package integration

import (
	"context"
	"testing"
	"time"

	"github.com/SubZzzzzz/pokeprofit/internal/discord"
	"github.com/SubZzzzzz/pokeprofit/internal/scraper/common"
)

// TestRateLimiterBasic tests basic rate limiter functionality.
func TestRateLimiterBasic(t *testing.T) {
	cfg := discord.RateLimitConfig{
		AnalyzeLimit:    1,
		AnalyzeWindow:   100 * time.Millisecond,
		ResultsLimit:    3,
		ResultsWindow:   100 * time.Millisecond,
		FilterLimit:     5,
		FilterWindow:    100 * time.Millisecond,
		DefaultLimit:    10,
		DefaultWindow:   100 * time.Millisecond,
		CleanupInterval: 1 * time.Hour, // Long cleanup for tests
	}

	rl := discord.NewRateLimiter(cfg)
	defer rl.Stop()

	userID := "test_user_123"

	// Test analyze command - should allow 1 request
	if !rl.Allow(userID, "analyze") {
		t.Error("First analyze request should be allowed")
	}

	// Second analyze should be blocked
	if rl.Allow(userID, "analyze") {
		t.Error("Second analyze request should be blocked")
	}

	// Test results command - should allow 3 requests
	for i := 0; i < 3; i++ {
		if !rl.Allow(userID, "results") {
			t.Errorf("Results request %d should be allowed", i+1)
		}
	}

	// Fourth results should be blocked
	if rl.Allow(userID, "results") {
		t.Error("Fourth results request should be blocked")
	}
}

// TestRateLimiterWindowReset tests that rate limits reset after the window expires.
func TestRateLimiterWindowReset(t *testing.T) {
	cfg := discord.RateLimitConfig{
		AnalyzeLimit:    1,
		AnalyzeWindow:   50 * time.Millisecond,
		ResultsLimit:    3,
		ResultsWindow:   50 * time.Millisecond,
		FilterLimit:     5,
		FilterWindow:    50 * time.Millisecond,
		DefaultLimit:    10,
		DefaultWindow:   50 * time.Millisecond,
		CleanupInterval: 1 * time.Hour,
	}

	rl := discord.NewRateLimiter(cfg)
	defer rl.Stop()

	userID := "test_user_456"

	// Use up the limit
	if !rl.Allow(userID, "analyze") {
		t.Error("First analyze request should be allowed")
	}

	if rl.Allow(userID, "analyze") {
		t.Error("Second analyze request should be blocked")
	}

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	// Should be allowed again
	if !rl.Allow(userID, "analyze") {
		t.Error("Analyze request after window reset should be allowed")
	}
}

// TestRateLimiterMultipleUsers tests that rate limits are per-user.
func TestRateLimiterMultipleUsers(t *testing.T) {
	cfg := discord.DefaultRateLimitConfig()
	cfg.AnalyzeLimit = 1
	cfg.AnalyzeWindow = 1 * time.Hour
	cfg.CleanupInterval = 1 * time.Hour

	rl := discord.NewRateLimiter(cfg)
	defer rl.Stop()

	user1 := "user_1"
	user2 := "user_2"

	// User 1 uses their limit
	if !rl.Allow(user1, "analyze") {
		t.Error("User 1 first analyze should be allowed")
	}

	// User 2 should still be able to make requests
	if !rl.Allow(user2, "analyze") {
		t.Error("User 2 first analyze should be allowed")
	}

	// User 1 is blocked, user 2 is also blocked now
	if rl.Allow(user1, "analyze") {
		t.Error("User 1 second analyze should be blocked")
	}

	if rl.Allow(user2, "analyze") {
		t.Error("User 2 second analyze should be blocked")
	}
}

// TestRateLimiterTimeUntilAllowed tests the TimeUntilAllowed function.
func TestRateLimiterTimeUntilAllowed(t *testing.T) {
	cfg := discord.RateLimitConfig{
		AnalyzeLimit:    1,
		AnalyzeWindow:   100 * time.Millisecond,
		ResultsLimit:    10,
		ResultsWindow:   100 * time.Millisecond,
		FilterLimit:     10,
		FilterWindow:    100 * time.Millisecond,
		DefaultLimit:    10,
		DefaultWindow:   100 * time.Millisecond,
		CleanupInterval: 1 * time.Hour,
	}

	rl := discord.NewRateLimiter(cfg)
	defer rl.Stop()

	userID := "test_user_789"

	// Before any usage, should return 0
	waitTime := rl.TimeUntilAllowed(userID, "analyze")
	if waitTime != 0 {
		t.Errorf("Expected 0 wait time for unused command, got %v", waitTime)
	}

	// Use up the limit
	rl.Allow(userID, "analyze")

	// Now should return non-zero
	waitTime = rl.TimeUntilAllowed(userID, "analyze")
	if waitTime <= 0 {
		t.Error("Expected positive wait time after rate limit hit")
	}

	if waitTime > 100*time.Millisecond {
		t.Errorf("Wait time should be less than window duration, got %v", waitTime)
	}
}

// TestRateLimiterCheck tests the Check function (non-modifying).
func TestRateLimiterCheck(t *testing.T) {
	cfg := discord.DefaultRateLimitConfig()
	cfg.AnalyzeLimit = 1
	cfg.AnalyzeWindow = 1 * time.Hour
	cfg.CleanupInterval = 1 * time.Hour

	rl := discord.NewRateLimiter(cfg)
	defer rl.Stop()

	userID := "test_user_check"

	// Check should return true without using the limit
	if !rl.Check(userID, "analyze") {
		t.Error("Check should return true for unused limit")
	}

	// Limit should still be available
	if !rl.Allow(userID, "analyze") {
		t.Error("Allow should succeed after Check")
	}

	// Now Check should return false
	if rl.Check(userID, "analyze") {
		t.Error("Check should return false after limit used")
	}
}

// TestRateLimiterReset tests the Reset functions.
func TestRateLimiterReset(t *testing.T) {
	cfg := discord.DefaultRateLimitConfig()
	cfg.AnalyzeLimit = 1
	cfg.AnalyzeWindow = 1 * time.Hour
	cfg.CleanupInterval = 1 * time.Hour

	rl := discord.NewRateLimiter(cfg)
	defer rl.Stop()

	userID := "test_user_reset"

	// Use up the limit
	rl.Allow(userID, "analyze")

	// Should be blocked
	if rl.Allow(userID, "analyze") {
		t.Error("Should be blocked after using limit")
	}

	// Reset the command
	rl.Reset(userID, "analyze")

	// Should be allowed again
	if !rl.Allow(userID, "analyze") {
		t.Error("Should be allowed after reset")
	}
}

// TestRateLimiterStats tests the Stats function.
func TestRateLimiterStats(t *testing.T) {
	cfg := discord.DefaultRateLimitConfig()
	cfg.CleanupInterval = 1 * time.Hour

	rl := discord.NewRateLimiter(cfg)
	defer rl.Stop()

	// Initially empty
	stats := rl.Stats()
	if stats.TotalUsers != 0 || stats.TotalCommands != 0 {
		t.Errorf("Expected empty stats, got users=%d, commands=%d", stats.TotalUsers, stats.TotalCommands)
	}

	// Add some usage
	rl.Allow("user1", "analyze")
	rl.Allow("user1", "results")
	rl.Allow("user2", "filter")

	stats = rl.Stats()
	if stats.TotalUsers != 2 {
		t.Errorf("Expected 2 users, got %d", stats.TotalUsers)
	}
	if stats.TotalCommands != 3 {
		t.Errorf("Expected 3 commands, got %d", stats.TotalCommands)
	}
}

// TestScraperRateLimiter tests the scraper rate limiter (1 req/s max for eBay).
func TestScraperRateLimiter(t *testing.T) {
	// Create rate limiter with 50ms delay for faster testing
	delay := 50 * time.Millisecond
	rl := common.NewRateLimiter(delay)

	ctx := context.Background()
	requestCount := 5
	start := time.Now()

	// Make requests
	for i := 0; i < requestCount; i++ {
		if err := rl.Wait(ctx); err != nil {
			t.Fatalf("Wait failed: %v", err)
		}
	}

	elapsed := time.Since(start)

	// Should take at least (requestCount-1) * delay
	minExpected := time.Duration(requestCount-1) * delay
	if elapsed < minExpected {
		t.Errorf("Rate limiter too fast: expected at least %v, got %v", minExpected, elapsed)
	}
}

// TestScraperRateLimiterCancellation tests rate limiter respects context cancellation.
func TestScraperRateLimiterCancellation(t *testing.T) {
	delay := 1 * time.Second
	rl := common.NewRateLimiter(delay)

	// Make one request to start the limiter
	ctx := context.Background()
	if err := rl.Wait(ctx); err != nil {
		t.Fatalf("First wait failed: %v", err)
	}

	// Create a context that will be cancelled quickly
	cancelCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	// Try to wait - should fail due to cancellation
	err := rl.Wait(cancelCtx)
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}
}
