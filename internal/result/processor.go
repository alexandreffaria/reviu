package result

import (
	"context"
	"time"
	
	"github.com/alexandreffaria/reviu/internal/logger"
)

// ProcessorOptions defines options for the result processing
type ProcessorOptions struct {
	MaxPages          int           // Maximum number of pages to process (0 = all)
	Timeout           int           // Timeout in seconds for the entire operation
	RetryAttempts     int           // Number of retry attempts for page navigation
	PageTimeout       int           // Timeout in seconds for processing a single page
	NavigationTimeout int           // Timeout in seconds for page navigation operations
	PageDelay         time.Duration // Delay between pages to avoid being blocked
}

// DefaultProcessorOptions returns default options for the processor
func DefaultProcessorOptions() ProcessorOptions {
	return ProcessorOptions{
		MaxPages:          0,              // Process all pages
		Timeout:           600,            // 10 minutes timeout for entire operation
		RetryAttempts:     3,              // 3 retry attempts
		PageTimeout:       30,             // 30 seconds per page
		NavigationTimeout: 30,             // 30 seconds for navigation operations
		PageDelay:         2 * time.Second, // 2 seconds delay between pages
	}
}

// ResultProcessor defines the interface for processing search results
type ResultProcessor interface {
	// Process extracts results from all pages and returns a collection
	Process(ctx context.Context, searchURL string) (*SearchCollection, error)
	
	// SetOptions configures the processor
	SetOptions(options ProcessorOptions)
	
	// SetLogger sets the logger for the processor
	SetLogger(log logger.Logger)
}

// RetryOptions configures the retry behavior
type RetryOptions struct {
	MaxAttempts  int
	InitialDelay int // Delay in milliseconds
	MaxDelay     int // Max delay in milliseconds
	Factor       float64 // Backoff factor
}

// DefaultRetryOptions provides default retry options
func DefaultRetryOptions() RetryOptions {
	return RetryOptions{
		MaxAttempts:  3,
		InitialDelay: 1000,  // 1 second
		MaxDelay:     30000, // 30 seconds
		Factor:       2.0,
	}
}