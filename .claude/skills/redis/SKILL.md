---
name: redis
description: Patterns Redis pour caching et queues. Cache de prix, rate limiting, job queues.
allowed-tools: Read, Grep, Glob, Edit, Write, Bash
---

# Redis Patterns (Go)

## Stack
- **go-redis/redis** : Client Redis standard
- **asynq** : Queue de jobs (alternative à Sidekiq)
- **go-redsync** : Distributed locks

## Setup de base

```go
package cache

import (
    "github.com/redis/go-redis/v9"
    "context"
)

func NewRedisClient(addr string) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:         addr,
        Password:     "", // no password
        DB:           0,
        PoolSize:     10,
        MaxRetries:   3,
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
    })
}
```

## 1. Caching Pattern

### Cache de prix (éviter re-scraping)

```go
type PriceCache struct {
    client *redis.Client
    ttl    time.Duration
}

func NewPriceCache(client *redis.Client) *PriceCache {
    return &PriceCache{
        client: client,
        ttl:    6 * time.Hour, // Cache 6h
    }
}

// Key pattern: price:{platform}:{product_id}
func (pc *PriceCache) cacheKey(platform, productID string) string {
    return fmt.Sprintf("price:%s:%s", platform, productID)
}

func (pc *PriceCache) Get(ctx context.Context, platform, productID string) (*Price, error) {
    key := pc.cacheKey(platform, productID)

    data, err := pc.client.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return nil, nil // Cache miss
    }
    if err != nil {
        return nil, err
    }

    var price Price
    if err := json.Unmarshal(data, &price); err != nil {
        return nil, err
    }

    return &price, nil
}

func (pc *PriceCache) Set(ctx context.Context, platform, productID string, price *Price) error {
    key := pc.cacheKey(platform, productID)

    data, err := json.Marshal(price)
    if err != nil {
        return err
    }

    return pc.client.Set(ctx, key, data, pc.ttl).Err()
}

// Cache-aside pattern
func (pc *PriceCache) GetOrFetch(ctx context.Context, platform, productID string,
    fetchFn func() (*Price, error)) (*Price, error) {

    // 1. Check cache
    price, err := pc.Get(ctx, platform, productID)
    if err != nil {
        return nil, err
    }
    if price != nil {
        return price, nil // Cache hit
    }

    // 2. Cache miss: fetch from source
    price, err = fetchFn()
    if err != nil {
        return nil, err
    }

    // 3. Store in cache
    if err := pc.Set(ctx, platform, productID, price); err != nil {
        // Log error but don't fail
        log.Printf("failed to cache price: %v", err)
    }

    return price, nil
}
```

## 2. Rate Limiting

### Rate limiter pour scrapers

```go
type RateLimiter struct {
    client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
    return &RateLimiter{client: client}
}

// Sliding window rate limiter
// Limite: maxRequests requêtes par window
func (rl *RateLimiter) Allow(ctx context.Context, key string, maxRequests int, window time.Duration) (bool, error) {
    now := time.Now().UnixNano()
    windowStart := now - window.Nanoseconds()

    pipe := rl.client.Pipeline()

    // 1. Supprimer entrées expirées
    pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

    // 2. Compter requêtes dans la fenêtre
    countCmd := pipe.ZCard(ctx, key)

    // 3. Ajouter cette requête
    pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

    // 4. Expiration de la clé
    pipe.Expire(ctx, key, window)

    _, err := pipe.Exec(ctx)
    if err != nil {
        return false, err
    }

    count := countCmd.Val()
    return count < int64(maxRequests), nil
}

// Usage dans scraper
func (s *Scraper) scrapeWithRateLimit(ctx context.Context, url string) error {
    rateLimitKey := fmt.Sprintf("ratelimit:scraper:%s", s.Name())

    // Max 10 requêtes par minute
    allowed, err := s.rateLimiter.Allow(ctx, rateLimitKey, 10, time.Minute)
    if err != nil {
        return err
    }

    if !allowed {
        return fmt.Errorf("rate limit exceeded, waiting...")
    }

    return s.doScrape(ctx, url)
}
```

## 3. Job Queue (avec Asynq)

### Définir les tasks

```go
package tasks

import (
    "github.com/hibiken/asynq"
    "encoding/json"
)

const (
    TypeScrapeEbay    = "scrape:ebay"
    TypeScrapeVinted  = "scrape:vinted"
    TypeAnalyzeVolume = "analyze:volume"
    TypeSendAlert     = "alert:send"
)

type ScrapePayload struct {
    Platform  string `json:"platform"`
    SearchURL string `json:"search_url"`
    ProductID string `json:"product_id,omitempty"`
}

func NewScrapeTask(payload ScrapePayload) (*asynq.Task, error) {
    data, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }
    return asynq.NewTask(TypeScrapeEbay, data), nil
}
```

### Producer (enqueue jobs)

```go
type TaskProducer struct {
    client *asynq.Client
}

func NewTaskProducer(redisAddr string) *TaskProducer {
    client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
    return &TaskProducer{client: client}
}

func (tp *TaskProducer) ScheduleScraping(ctx context.Context) error {
    // Scrape eBay toutes les heures
    task, err := NewScrapeTask(ScrapePayload{
        Platform:  "ebay",
        SearchURL: "https://ebay.fr/...",
    })
    if err != nil {
        return err
    }

    info, err := tp.client.Enqueue(task,
        asynq.Queue("scrapers"),
        asynq.MaxRetry(3),
        asynq.Timeout(5*time.Minute),
    )
    if err != nil {
        return err
    }

    log.Printf("enqueued task: id=%s queue=%s", info.ID, info.Queue)
    return nil
}
```

### Consumer (process jobs)

```go
type TaskHandler struct {
    scrapers map[string]Scraper
    analyzer *VolumeAnalyzer
}

func (h *TaskHandler) HandleScrapeTask(ctx context.Context, t *asynq.Task) error {
    var payload ScrapePayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return fmt.Errorf("unmarshal payload: %w", err)
    }

    scraper, ok := h.scrapers[payload.Platform]
    if !ok {
        return fmt.Errorf("unknown platform: %s", payload.Platform)
    }

    sales, err := scraper.Scrape(ctx)
    if err != nil {
        return fmt.Errorf("scrape failed: %w", err)
    }

    log.Printf("scraped %d sales from %s", len(sales), payload.Platform)
    return nil
}

func StartWorker(redisAddr string, handler *TaskHandler) error {
    srv := asynq.NewServer(
        asynq.RedisClientOpt{Addr: redisAddr},
        asynq.Config{
            Concurrency: 10,
            Queues: map[string]int{
                "critical": 6, // 60% des workers
                "default":  3, // 30%
                "low":      1, // 10%
            },
        },
    )

    mux := asynq.NewServeMux()
    mux.HandleFunc(TypeScrapeEbay, handler.HandleScrapeTask)
    mux.HandleFunc(TypeScrapeVinted, handler.HandleScrapeTask)

    return srv.Run(mux)
}
```

## 4. Distributed Locks

### Lock pour éviter double-scraping

```go
import "github.com/go-redsync/redsync/v4"

type LockManager struct {
    rs *redsync.Redsync
}

func (lm *LockManager) WithLock(ctx context.Context, key string, fn func() error) error {
    mutex := lm.rs.NewMutex(key,
        redsync.WithExpiry(5*time.Minute),
        redsync.WithTries(3),
    )

    if err := mutex.LockContext(ctx); err != nil {
        return fmt.Errorf("failed to acquire lock: %w", err)
    }
    defer mutex.Unlock()

    return fn()
}

// Usage
func (s *Scraper) ScrapeWithLock(ctx context.Context) error {
    lockKey := fmt.Sprintf("lock:scraper:%s", s.Name())

    return s.lockManager.WithLock(ctx, lockKey, func() error {
        return s.doScrape(ctx)
    })
}
```

## 5. Pub/Sub (Real-time alerts)

```go
type AlertPublisher struct {
    client *redis.Client
}

func (ap *AlertPublisher) Publish(ctx context.Context, alert Alert) error {
    data, err := json.Marshal(alert)
    if err != nil {
        return err
    }

    return ap.client.Publish(ctx, "alerts", data).Err()
}

type AlertSubscriber struct {
    client *redis.Client
}

func (as *AlertSubscriber) Subscribe(ctx context.Context, handler func(Alert)) error {
    pubsub := as.client.Subscribe(ctx, "alerts")
    defer pubsub.Close()

    ch := pubsub.Channel()

    for msg := range ch {
        var alert Alert
        if err := json.Unmarshal([]byte(msg.Payload), &alert); err != nil {
            log.Printf("unmarshal error: %v", err)
            continue
        }

        handler(alert)
    }

    return nil
}
```

## Best Practices

1. **Toujours set TTL** sur les clés (éviter memory leaks)
2. **Prefixer les clés** par namespace (ex: `price:`, `lock:`)
3. **Pipeline** pour opérations groupées
4. **Retry logic** sur failures Redis
5. **Monitoring** : track hit rate, latency, memory usage
