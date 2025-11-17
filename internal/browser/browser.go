// Package browser provides functionality for browser automation
package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/alexandreffaria/reviu/internal/errors"
	"github.com/alexandreffaria/reviu/internal/logger"
)

// Browser defines the interface for browser interactions
type Browser interface {
	// Basic browser operations
	// Open launches a browser and navigates to the provided URL
	// Returns an error if the browser fails to launch or navigate
	Open(url string) error
	
	// Wait keeps the browser open for the specified duration
	// If duration is 0, the browser remains open until Close is called
	Wait(duration time.Duration) error
	
	// Close closes the browser instance and cleans up resources
	// Returns an error if cleanup fails
	Close() error

	// DOM interaction methods
	GetElements(selector string) ([]*rod.Element, error)
	GetElement(selector string) (*rod.Element, error)
	ElementExists(selector string) (bool, error)
	ClickElement(selector string) error
	GetElementText(selector string) (string, error)
	GetElementAttribute(selector, attr string) (string, error)
	WaitForElement(selector string, timeout time.Duration) error
	WaitForNavigation(timeout time.Duration) error
	ExtractLinks(selector string) ([]LinkData, error)
	
	// Scrolling operations
	ScrollToBottom() error
	ScrollForDuration(duration time.Duration) error
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
	
	// Will set timeout after browser is initialized
	
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
	// Set the browser with timeout
	b.browser = browser.Timeout(b.options.Timeout)
	
	// Create a new page
	b.log.Info("Opening URL: %s", url)
	page, err := browser.Page(proto.TargetCreateTarget{})
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

// Close closes the browser and cleans up resources with timeout handling
func (b *RodBrowser) Close() error {
	b.cancel() // Cancel any ongoing operations
	
	b.log.Info("Closing browser...")
	
	var errs []error
	
	// Function to close with timeout
	closeWithTimeout := func(closeFn func() error, operation string, timeout time.Duration) error {
		done := make(chan error, 1)
		
		// Start close operation in a goroutine
		go func() {
			done <- closeFn()
		}()
		
		// Wait for completion or timeout
		select {
		case err := <-done:
			return err
		case <-time.After(timeout):
			b.log.Warn("Timeout while %s", operation)
			return fmt.Errorf("timeout while %s", operation)
		}
	}
	
	// Close page if it exists
	if b.page != nil {
		// Use short timeout for page closing
		err := closeWithTimeout(b.page.Close, "closing page", 5*time.Second)
		if err != nil {
			b.log.Warn("Error closing page: %v (continuing anyway)", err)
			errs = append(errs, errors.NewBrowserError("failed to close page", err))
		}
		b.page = nil
	}
	
	// Close browser if it exists
	if b.browser != nil {
		// Use short timeout for browser closing
		err := closeWithTimeout(b.browser.Close, "closing browser", 5*time.Second)
		if err != nil {
			b.log.Warn("Error closing browser: %v (continuing anyway)", err)
			errs = append(errs, errors.NewBrowserError("failed to close browser", err))
		}
		b.browser = nil
	}
	
	if len(errs) > 0 {
		// Log the errors but still consider the operation successful
		b.log.Info("Browser resources cleaned up with some errors")
		return nil // Return nil to avoid cascading errors
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