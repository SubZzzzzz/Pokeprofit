package common

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"
)

// RetryConfig configures the retry behavior.
type RetryConfig struct {
	MaxRetries     int
	InitialDelay   time.Duration
	MaxDelay       time.Duration
	BackoffFactor  float64
	Jitter         bool
	RetryableCheck func(error) bool
}

// DefaultRetryConfig returns the default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialDelay:   2 * time.Second,
		MaxDelay:       30 * time.Second,
		BackoffFactor:  2.0,
		Jitter:         true,
		RetryableCheck: nil, // Retry all errors by default
	}
}

// Retryer handles retry logic with exponential backoff.
type Retryer struct {
	config RetryConfig
}

// NewRetryer creates a new Retryer with the given configuration.
func NewRetryer(config RetryConfig) *Retryer {
	return &Retryer{config: config}
}

// NewDefaultRetryer creates a new Retryer with default configuration.
func NewDefaultRetryer() *Retryer {
	return NewRetryer(DefaultRetryConfig())
}

// Do executes the function with retry logic.
// The function is retried until it succeeds, max retries is reached, or context is cancelled.
func (r *Retryer) Do(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if r.config.RetryableCheck != nil && !r.config.RetryableCheck(err) {
			return err
		}

		// Don't wait after the last attempt
		if attempt == r.config.MaxRetries {
			break
		}

		delay := r.calculateDelay(attempt)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return &RetryError{
		Attempts: r.config.MaxRetries + 1,
		Err:      lastErr,
	}
}

// DoWithResult executes a function that returns a result with retry logic.
func (r *Retryer) DoWithResult(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	var result interface{}
	var lastErr error

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		var err error
		result, err = fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if r.config.RetryableCheck != nil && !r.config.RetryableCheck(err) {
			return nil, err
		}

		// Don't wait after the last attempt
		if attempt == r.config.MaxRetries {
			break
		}

		delay := r.calculateDelay(attempt)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}

	return nil, &RetryError{
		Attempts: r.config.MaxRetries + 1,
		Err:      lastErr,
	}
}

func (r *Retryer) calculateDelay(attempt int) time.Duration {
	delay := float64(r.config.InitialDelay) * math.Pow(r.config.BackoffFactor, float64(attempt))

	if r.config.Jitter {
		// Add jitter: ±25% of the delay
		jitterFactor := 0.75 + rand.Float64()*0.5
		delay *= jitterFactor
	}

	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}

	return time.Duration(delay)
}

// RetryError represents an error after all retries have been exhausted.
type RetryError struct {
	Attempts int
	Err      error
}

func (e *RetryError) Error() string {
	return "max retries exceeded after " + string(rune(e.Attempts+'0')) + " attempts: " + e.Err.Error()
}

func (e *RetryError) Unwrap() error {
	return e.Err
}

// IsRetryError checks if the error is a RetryError.
func IsRetryError(err error) bool {
	var retryErr *RetryError
	return errors.As(err, &retryErr)
}

// Retry is a convenience function that executes fn with default retry settings.
func Retry(ctx context.Context, fn func() error) error {
	return NewDefaultRetryer().Do(ctx, fn)
}

// RetryWithConfig is a convenience function that executes fn with custom retry settings.
func RetryWithConfig(ctx context.Context, config RetryConfig, fn func() error) error {
	return NewRetryer(config).Do(ctx, fn)
}
