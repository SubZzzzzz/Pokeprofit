package errors

import "errors"

// Domain errors for the application.
var (
	// ErrNoData indicates no analysis data is available.
	ErrNoData = errors.New("no analysis data available")

	// ErrAnalysisRunning indicates an analysis is already in progress.
	ErrAnalysisRunning = errors.New("analysis already running")

	// ErrScrapeFailed indicates the scraper could not reach the target.
	ErrScrapeFailed = errors.New("scraping failed")

	// ErrRateLimited indicates too many requests.
	ErrRateLimited = errors.New("rate limited by target platform")

	// ErrProductNotFound indicates the product doesn't exist.
	ErrProductNotFound = errors.New("product not found")

	// ErrInvalidCategory indicates an unknown product category.
	ErrInvalidCategory = errors.New("invalid product category")

	// ErrDuplicateSale indicates a sale with this URL already exists.
	ErrDuplicateSale = errors.New("duplicate sale URL")

	// ErrInvalidInput indicates invalid user input.
	ErrInvalidInput = errors.New("invalid input")

	// ErrTimeout indicates an operation timed out.
	ErrTimeout = errors.New("operation timed out")

	// ErrDatabaseConnection indicates a database connection error.
	ErrDatabaseConnection = errors.New("database connection failed")

	// ErrCacheConnection indicates a Redis connection error.
	ErrCacheConnection = errors.New("cache connection failed")
)

// Is checks if the error is of a specific type.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Wrap wraps an error with additional context.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return &wrappedError{
		msg: message,
		err: err,
	}
}

type wrappedError struct {
	msg string
	err error
}

func (e *wrappedError) Error() string {
	return e.msg + ": " + e.err.Error()
}

func (e *wrappedError) Unwrap() error {
	return e.err
}

// ScrapeError represents an error during scraping.
type ScrapeError struct {
	URL     string
	Message string
	Err     error
}

func (e *ScrapeError) Error() string {
	if e.Err != nil {
		return e.Message + " (" + e.URL + "): " + e.Err.Error()
	}
	return e.Message + " (" + e.URL + ")"
}

func (e *ScrapeError) Unwrap() error {
	return e.Err
}

// NewScrapeError creates a new ScrapeError.
func NewScrapeError(url, message string, err error) *ScrapeError {
	return &ScrapeError{
		URL:     url,
		Message: message,
		Err:     err,
	}
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
