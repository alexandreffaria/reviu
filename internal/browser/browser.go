// Package browser provides functionality for browser automation
package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/alexandreffaria/reviu/internal/errors"
	"github.com/alexandreffaria/reviu/internal/logger"
)

// Browser defines the interface for browser interactions
type Browser interface {
	// Open launches a browser and navigates to the provided URL
	// Returns an error if the browser fails to launch or navigate
	Open(url string) error
	
	// Wait keeps the browser open for the specified duration
	// If duration is 0, the browser remains open until Close is called
	Wait(duration time.Duration) error
	
	// Close closes the browser instance and cleans up resources
	// Returns an error if cleanup fails
	Close() error
}

// BrowserOptions contains configuration options for the browser
type BrowserOptions struct {
	// Headless determines whether the browser runs without a visible UI
	Headless bool
	
	// DefaultWaitTime is the default time to keep the browser open
	DefaultWaitTime time.Duration
	
	// Timeout for browser operations
	Timeout time.Duration
}

// DefaultBrowserOptions provides sensible defaults
var DefaultBrowserOptions = BrowserOptions{
	Headless:        false,
	DefaultWaitTime: 30 * time.Second,
	Timeout:         60 * time.Second,
}

// RodBrowser implements Browser using the Rod library
type RodBrowser struct {
	browser *rod.Browser
	page    *rod.Page
	log     logger.Logger
	options BrowserOptions
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewBrowser creates a new browser with the provided options
func NewBrowser(log logger.Logger, options *BrowserOptions) Browser {
	opts := DefaultBrowserOptions
	if options != nil {
		opts = *options
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	if log == nil {
		// Create a null logger if none provided
		log = logger.NewLogger(logger.WithWriter(nil))
	}
	
	return &RodBrowser{
		log:     log.WithPrefix("Browser"),
		options: opts,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Open launches a browser and navigates to the specified URL
func (b *RodBrowser) Open(url string) error {
	b.log.Info("Launching browser...")
	
	// Create a timeout context for browser operations
	timeoutCtx, cancel := context.WithTimeout(b.ctx, b.options.Timeout)
	defer cancel()
	
	// Configure and launch the browser
	l := launcher.New().Headless(b.options.Headless)
	
	launchURL, err := l.Launch()
	if err != nil {
		return errors.NewBrowserError("failed to launch browser", err)
	}
	
	// Connect to the browser
	browser := rod.New().ControlURL(launchURL)
	err = browser.Connect()
	if err != nil {
		return errors.NewBrowserError("failed to connect to browser", err)
	}
	b.browser = browser
	
	// Create a new page
	b.log.Info("Opening URL: %s", url)
	page, err := browser.Page(timeoutCtx)
	if err != nil {
		b.Close() // Clean up on error
		return errors.NewBrowserError("failed to create page", err)
	}
	b.page = page
	
	// Navigate to the URL
	err = page.Navigate(url)
	if err != nil {
		b.Close() // Clean up on error
		return errors.NewBrowserError("failed to navigate to URL", err)
	}
	
	// Wait for page to load
	err = page.WaitLoad()
	if err != nil {
		b.Close() // Clean up on error
		return errors.NewBrowserError("failed to wait for page load", err)
	}
	
	b.log.Info("Page loaded successfully")
	return nil
}

// Wait keeps the browser open for the specified duration
func (b *RodBrowser) Wait(duration time.Duration) error {
	if b.browser == nil {
		return errors.NewBrowserError("browser not initialized, call Open first", nil)
	}
	
	if duration == 0 {
		duration = b.options.DefaultWaitTime
	}
	
	b.log.Info("Keeping browser open for %v", duration)
	
	// Create a timer that will be canceled if Close is called
	timer := time.NewTimer(duration)
	
	select {
	case <-timer.C:
		b.log.Debug("Wait timer expired")
		return nil
	case <-b.ctx.Done():
		b.log.Debug("Wait canceled")
		if !timer.Stop() {
			<-timer.C // Drain the channel if timer already fired
		}
		return fmt.Errorf("wait canceled: %w", b.ctx.Err())
	}
}

// Close closes the browser and cleans up resources
func (b *RodBrowser) Close() error {
	b.cancel() // Cancel any ongoing operations
	
	b.log.Info("Closing browser...")
	
	var errs []error
	
	// Close page if it exists
	if b.page != nil {
		if err := b.page.Close(); err != nil {
			errs = append(errs, errors.NewBrowserError("failed to close page", err))
		}
		b.page = nil
	}
	
	// Close browser if it exists
	if b.browser != nil {
		if err := b.browser.Close(); err != nil {
			errs = append(errs, errors.NewBrowserError("failed to close browser", err))
		}
		b.browser = nil
	}
	
	if len(errs) > 0 {
		// Return the first error for simplicity
		// In a more complex implementation, we might want to combine errors
		return errs[0]
	}
	
	b.log.Info("Browser closed successfully")
	return nil
}

// WithHeadless creates a copy of options with headless setting modified
func (o BrowserOptions) WithHeadless(headless bool) BrowserOptions {
	o.Headless = headless
	return o
}

// WithDefaultWaitTime creates a copy of options with wait time modified
func (o BrowserOptions) WithDefaultWaitTime(duration time.Duration) BrowserOptions {
	o.DefaultWaitTime = duration
	return o
}

// WithTimeout creates a copy of options with timeout modified
func (o BrowserOptions) WithTimeout(duration time.Duration) BrowserOptions {
	o.Timeout = duration
	return o
}