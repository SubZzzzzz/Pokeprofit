package common

import (
	"context"
	"sync"
	"time"
)

// RateLimiter controls the rate of operations.
type RateLimiter struct {
	mu       sync.Mutex
	delay    time.Duration
	lastCall time.Time
}

// NewRateLimiter creates a new rate limiter with the specified delay between operations.
func NewRateLimiter(delay time.Duration) *RateLimiter {
	return &RateLimiter{
		delay: delay,
	}
}

// Wait blocks until the rate limit allows the next operation.
// Returns an error if the context is cancelled.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.lastCall.IsZero() {
		rl.lastCall = time.Now()
		return nil
	}

	elapsed := time.Since(rl.lastCall)
	if elapsed < rl.delay {
		waitTime := rl.delay - elapsed

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}
	}

	rl.lastCall = time.Now()
	return nil
}

// Delay returns the configured delay between operations.
func (rl *RateLimiter) Delay() time.Duration {
	return rl.delay
}

// Reset resets the rate limiter state.
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.lastCall = time.Time{}
}

// TokenBucket implements a token bucket rate limiter for burst traffic.
type TokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

// NewTokenBucket creates a new token bucket rate limiter.
// maxTokens is the bucket capacity, refillRate is tokens added per second.
func NewTokenBucket(maxTokens, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Take attempts to take a token from the bucket.
// Returns true if a token was available, false otherwise.
func (tb *TokenBucket) Take() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}

	return false
}

// TakeWait blocks until a token is available or context is cancelled.
func (tb *TokenBucket) TakeWait(ctx context.Context) error {
	for {
		if tb.Take() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			// Check again
		}
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}
	tb.lastRefill = now
}

// Available returns the current number of available tokens.
func (tb *TokenBucket) Available() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	return tb.tokens
}
