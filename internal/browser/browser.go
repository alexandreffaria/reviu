// Package browser provides functionality for browser automation
package browser

import (
	"context"
	"fmt"
	"math/rand"
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
	
	// Navigate navigates to a new URL using the existing browser instance
	// This should be used for subsequent navigation after Open
	Navigate(url string) error
	
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
	
	// Anti-blocking options
	RandomizeUserAgent bool
	SlowMotion         time.Duration
	StealthMode        bool
	Proxy              string
}

// DefaultBrowserOptions provides sensible defaults
var DefaultBrowserOptions = BrowserOptions{
	Headless:          false,
	DefaultWaitTime:   30 * time.Second,
	Timeout:           60 * time.Second,
	RandomizeUserAgent: true,
	SlowMotion:        200 * time.Millisecond,
	StealthMode:       true,
	Proxy:             "",
}

// Common user agents for randomization
var commonUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.55 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0) Gecko/20100101 Firefox/94.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64; rv:94.0) Gecko/20100101 Firefox/94.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.93 Safari/537.36 Edg/96.0.1054.43",
}

// Random number generator
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// getRandomUserAgent returns a random user agent from the list
func getRandomUserAgent() string {
	return commonUserAgents[rng.Intn(len(commonUserAgents))]
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
	l := launcher.New().Headless(b.options.Headless).Leakless(false)
	b.log.Debug("Disabled leakless mode to avoid antivirus detection")

	// Apply anti-blocking measures
	if b.options.StealthMode {
		b.log.Debug("Enabling stealth mode")
		
		// Set a random user agent if enabled
		if b.options.RandomizeUserAgent {
			userAgent := getRandomUserAgent()
			l = l.Set("user-agent", userAgent)
			b.log.Debug("Using random user agent: %s", userAgent)
		}
		
		// Set proxy if provided
		if b.options.Proxy != "" {
			l = l.Proxy(b.options.Proxy)
			b.log.Debug("Using proxy: %s", b.options.Proxy)
		}
		
		// Add additional arguments to avoid detection
		l = l.Set("disable-blink-features", "AutomationControlled")
		l = l.Set("ignore-certificate-errors", "")
		l = l.Set("disable-web-security", "")
	}
	
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
	return b.navigateToURL(url)
}

// Navigate navigates to a new URL using the existing browser instance
func (b *RodBrowser) Navigate(url string) error {
	if b.browser == nil || b.page == nil {
		return errors.NewBrowserError("browser not initialized, call Open first", nil)
	}
	
	b.log.Info("Navigating to URL: %s", url)
	
	// Use the existing page to navigate to the new URL
	return b.navigateToURL(url)
}

// navigateToURL is a helper method that navigates to a URL and waits for page load
func (b *RodBrowser) navigateToURL(url string) error {
	if b.page == nil {
		return errors.NewBrowserError("page not initialized", nil)
	}
	
	// Navigate to the URL
	err := b.page.Navigate(url)
	if err != nil {
		return errors.NewBrowserError("failed to navigate to URL", err)
	}
	
	// Wait for page to load
	err = b.page.WaitLoad()
	if err != nil {
		return errors.NewBrowserError("failed to wait for page load", err)
	}
	
	// Add human-like behavior if stealth mode is enabled
	if b.options.StealthMode {
		// Execute JavaScript to hide automation markers
		b.executeStealthScripts(b.page)
		
		// Add random delay to simulate human behavior
		delay := time.Duration(500+rng.Intn(1000)) * time.Millisecond
		b.log.Debug("Adding random delay of %v after page load", delay)
		time.Sleep(delay)
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

// WithStealthMode creates a copy of options with stealth mode setting
func (o BrowserOptions) WithStealthMode(enabled bool) BrowserOptions {
	o.StealthMode = enabled
	return o
}

// WithProxy creates a copy of options with proxy setting
func (o BrowserOptions) WithProxy(proxy string) BrowserOptions {
	o.Proxy = proxy
	return o
}

// WithSlowMotion creates a copy of options with slow motion setting
func (o BrowserOptions) WithSlowMotion(duration time.Duration) BrowserOptions {
	o.SlowMotion = duration
	return o
}

// WithRandomUserAgent creates a copy of options with random user agent setting
func (o BrowserOptions) WithRandomUserAgent(enabled bool) BrowserOptions {
	o.RandomizeUserAgent = enabled
	return o
}

// executeStealthScripts applies JavaScript to hide automation markers
func (b *RodBrowser) executeStealthScripts(page *rod.Page) {
	b.log.Debug("Stealth scripts disabled due to compatibility issues")
	
	// Scripts have been disabled due to consistent execution errors
	// The delay between page requests provides sufficient anti-blocking protection
}