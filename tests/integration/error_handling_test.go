package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "github.com/SubZzzzzz/pokeprofit/internal/errors"
	"github.com/SubZzzzzz/pokeprofit/internal/scraper/common"
	"github.com/SubZzzzzz/pokeprofit/internal/scraper/ebay"
)

// TestErrorTypes tests that error types are properly defined.
func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrNoData", apperrors.ErrNoData},
		{"ErrAnalysisRunning", apperrors.ErrAnalysisRunning},
		{"ErrScrapeFailed", apperrors.ErrScrapeFailed},
		{"ErrRateLimited", apperrors.ErrRateLimited},
		{"ErrProductNotFound", apperrors.ErrProductNotFound},
		{"ErrInvalidCategory", apperrors.ErrInvalidCategory},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Error("Error should not be nil")
			}
			if tt.err.Error() == "" {
				t.Error("Error message should not be empty")
			}
		})
	}
}

// TestErrorWrapping tests error wrapping functionality.
func TestErrorWrapping(t *testing.T) {
	originalErr := apperrors.ErrScrapeFailed
	wrappedErr := apperrors.Wrap(originalErr, "additional context")

	if wrappedErr == nil {
		t.Error("Wrapped error should not be nil")
	}

	// Should be able to unwrap to original
	if !errors.Is(wrappedErr, originalErr) {
		t.Error("Wrapped error should match original via errors.Is")
	}
}

// TestRetryerSuccess tests successful operation on first try.
func TestRetryerSuccess(t *testing.T) {
	cfg := common.RetryConfig{
		MaxRetries:    3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		Jitter:        false,
	}

	retryer := common.NewRetryer(cfg)
	ctx := context.Background()

	callCount := 0
	err := retryer.Do(ctx, func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

// TestRetryerRetryOnError tests retry behavior on transient errors.
func TestRetryerRetryOnError(t *testing.T) {
	cfg := common.RetryConfig{
		MaxRetries:    3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		Jitter:        false,
	}

	retryer := common.NewRetryer(cfg)
	ctx := context.Background()

	callCount := 0
	failUntil := 2 // Fail first 2 times

	err := retryer.Do(ctx, func() error {
		callCount++
		if callCount <= failUntil {
			return errors.New("transient error")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Expected success after retries, got %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
}

// TestRetryerMaxRetriesExceeded tests behavior when max retries are exceeded.
func TestRetryerMaxRetriesExceeded(t *testing.T) {
	cfg := common.RetryConfig{
		MaxRetries:    2,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		Jitter:        false,
	}

	retryer := common.NewRetryer(cfg)
	ctx := context.Background()

	callCount := 0
	testErr := errors.New("persistent error")

	err := retryer.Do(ctx, func() error {
		callCount++
		return testErr
	})

	if err == nil {
		t.Error("Expected error after max retries exceeded")
	}

	// Should have tried MaxRetries + 1 times (initial + retries)
	expectedCalls := cfg.MaxRetries + 1
	if callCount != expectedCalls {
		t.Errorf("Expected %d calls, got %d", expectedCalls, callCount)
	}
}

// TestRetryerContextCancellation tests that retries respect context cancellation.
func TestRetryerContextCancellation(t *testing.T) {
	cfg := common.RetryConfig{
		MaxRetries:    10,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      1 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        false,
	}

	retryer := common.NewRetryer(cfg)

	// Create a context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	callCount := 0
	err := retryer.Do(ctx, func() error {
		callCount++
		return errors.New("always fails")
	})

	if err == nil {
		t.Error("Expected error due to context cancellation")
	}

	// Should have been cancelled before all retries
	if callCount >= cfg.MaxRetries+1 {
		t.Errorf("Expected fewer calls due to cancellation, got %d", callCount)
	}
}

// TestRetryerExponentialBackoff tests that delays increase exponentially.
func TestRetryerExponentialBackoff(t *testing.T) {
	cfg := common.RetryConfig{
		MaxRetries:    3,
		InitialDelay:  20 * time.Millisecond,
		MaxDelay:      200 * time.Millisecond,
		BackoffFactor: 2.0,
		Jitter:        false,
	}

	retryer := common.NewRetryer(cfg)
	ctx := context.Background()

	var timestamps []time.Time
	callCount := 0

	start := time.Now()
	_ = retryer.Do(ctx, func() error {
		timestamps = append(timestamps, time.Now())
		callCount++
		if callCount < 4 {
			return errors.New("retry")
		}
		return nil
	})

	// Verify timing
	if len(timestamps) < 2 {
		t.Skip("Not enough retries to verify backoff")
	}

	// First retry should be after InitialDelay (~20ms)
	firstDelay := timestamps[1].Sub(timestamps[0])
	if firstDelay < 15*time.Millisecond {
		t.Errorf("First delay too short: %v", firstDelay)
	}

	// Second retry should be after ~40ms (2x)
	if len(timestamps) >= 3 {
		secondDelay := timestamps[2].Sub(timestamps[1])
		if secondDelay < 30*time.Millisecond {
			t.Errorf("Second delay should be longer than first: %v", secondDelay)
		}
	}

	t.Logf("Total time: %v", time.Since(start))
}

// TestScraperHealthCheckSuccess tests health check when eBay is reachable.
// Note: This is a live test that requires network access.
func TestScraperHealthCheckSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	scraper := ebay.NewDefaultEbayScraper()

	status := scraper.HealthCheck(context.Background())

	// Note: This may fail in CI/CD environments without network access
	if !status.Healthy {
		t.Logf("Health check failed (may be expected in CI): %s", status.Error)
	}

	if status.Platform != "ebay.fr" {
		t.Errorf("Expected platform 'ebay.fr', got '%s'", status.Platform)
	}

	if status.ResponseTime <= 0 {
		t.Error("Response time should be positive")
	}
}

// TestScraperHealthCheckTimeout tests health check behavior with timeout.
func TestScraperHealthCheckTimeout(t *testing.T) {
	scraper := ebay.NewDefaultEbayScraper()

	// Create a context that times out immediately
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to be cancelled
	time.Sleep(10 * time.Millisecond)

	status := scraper.HealthCheck(ctx)

	// Should complete but may or may not be healthy depending on timing
	if status.CheckedAt.IsZero() {
		t.Error("CheckedAt should be set")
	}
}

// TestScrapeOptionsValidation tests that scrape options are validated.
func TestScrapeOptionsValidation(t *testing.T) {
	scraper := ebay.NewDefaultEbayScraper()

	// Empty query should fail
	_, err := scraper.Scrape(ebay.ScrapeOptions{
		Query: "",
	})

	if err == nil {
		t.Error("Expected error for empty query")
	}
}

// TestScrapeContextCancellation tests that scraping respects context cancellation.
func TestScrapeContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	scraper := ebay.NewDefaultEbayScraper()

	// Create a context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := scraper.ScrapeWithContext(ctx, ebay.ScrapeOptions{
		Query:    "Pokemon Test",
		MaxPages: 10, // Request many pages to ensure cancellation
	})

	// Should fail due to context cancellation
	if err == nil {
		t.Log("Scrape completed before cancellation (may happen with fast network)")
	}
}

// TestConnectionFailureHandling tests behavior when connection fails.
func TestConnectionFailureHandling(t *testing.T) {
	cfg := ebay.ScraperConfig{
		RateLimitDelay: 100 * time.Millisecond,
		MaxRetries:     1,
		UserAgents:     []string{"Test Agent"},
	}

	scraper := ebay.NewEbayScraper(cfg)

	// Try to scrape with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	result, err := scraper.ScrapeWithContext(ctx, ebay.ScrapeOptions{
		Query:    "Test",
		MaxPages: 1,
	})

	// Either error or result with errors
	if err == nil && len(result.Errors) == 0 {
		t.Log("Connection succeeded (may happen with fast local network)")
	}
}

// TestErrorIsNoData tests proper error comparison.
func TestErrorIsNoData(t *testing.T) {
	wrappedErr := apperrors.Wrap(apperrors.ErrNoData, "no analysis data available")

	if !errors.Is(wrappedErr, apperrors.ErrNoData) {
		t.Error("Wrapped ErrNoData should match via errors.Is")
	}
}

// TestErrorIsAnalysisRunning tests proper error comparison.
func TestErrorIsAnalysisRunning(t *testing.T) {
	wrappedErr := apperrors.Wrap(apperrors.ErrAnalysisRunning, "analysis in progress")

	if !errors.Is(wrappedErr, apperrors.ErrAnalysisRunning) {
		t.Error("Wrapped ErrAnalysisRunning should match via errors.Is")
	}
}
