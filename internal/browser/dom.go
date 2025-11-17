package browser

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/alexandreffaria/reviu/internal/errors"
)

// LinkData represents data extracted from an anchor element
type LinkData struct {
	Text       string
	URL        string
	Attributes map[string]string
}

// ScrollToBottom scrolls the page to the bottom
func (b *RodBrowser) ScrollToBottom() error {
	if b.page == nil {
		return errors.NewBrowserError("browser page not initialized, call Open first", nil)
	}
	
	b.log.Debug("Scrolling to bottom of page...")
	
	// Execute JavaScript to scroll to the bottom
	_, err := b.page.Eval(`window.scrollTo(0, document.body.scrollHeight)`)
	if err != nil {
		return errors.NewBrowserError("failed to scroll to bottom", err)
	}
	
	return nil
}

// ScrollForDuration scrolls the page repeatedly for a specified duration
func (b *RodBrowser) ScrollForDuration(duration time.Duration) error {
	if b.page == nil {
		return errors.NewBrowserError("browser page not initialized, call Open first", nil)
	}
	
	b.log.Debug("Scrolling continuously for %v...", duration)
	
	// Get start time
	startTime := time.Now()
	
	// Scroll multiple times until the duration is reached
	for time.Since(startTime) < duration {
		// Scroll down
		_, err := b.page.Eval(`window.scrollBy(0, 500)`)
		if err != nil {
			return errors.NewBrowserError("failed to scroll page", err)
		}
		
		// Brief pause between scrolls
		time.Sleep(200 * time.Millisecond)
	}
	
	b.log.Debug("Completed scrolling for %v", duration)
	return nil
}

// GetElements returns all elements matching the provided CSS selector
func (b *RodBrowser) GetElements(selector string) ([]*rod.Element, error) {
	if b.page == nil {
		return nil, errors.NewBrowserError("browser page not initialized, call Open first", nil)
	}
	
	// Set a timeout for this operation
	b.page = b.page.Timeout(5 * time.Second)
	
	// Attempt to find the elements
	elements, err := b.page.Elements(selector)
	if err != nil {
		return nil, errors.NewBrowserError(fmt.Sprintf("failed to find elements with selector: %s", selector), err)
	}
	
	b.log.Debug("Found %d elements matching selector: %s", len(elements), selector)
	return elements, nil
}

// GetElement returns the first element matching the provided CSS selector
func (b *RodBrowser) GetElement(selector string) (*rod.Element, error) {
	if b.page == nil {
		return nil, errors.NewBrowserError("browser page not initialized, call Open first", nil)
	}
	
	// Set a timeout for this operation
	b.page = b.page.Timeout(5 * time.Second)
	
	// Attempt to find the element
	element, err := b.page.Element(selector)
	if err != nil {
		return nil, errors.NewBrowserError(fmt.Sprintf("failed to find element with selector: %s", selector), err)
	}
	
	if element == nil {
		return nil, errors.NewBrowserError(fmt.Sprintf("element not found with selector: %s", selector), nil)
	}
	
	return element, nil
}

// ElementExists checks if an element exists in the page
func (b *RodBrowser) ElementExists(selector string) (bool, error) {
	if b.page == nil {
		return false, errors.NewBrowserError("browser page not initialized, call Open first", nil)
	}
	
	// Attempt to find the element
	element, err := b.page.Element(selector)
	
	// Element not found
	if err != nil {
		if strings.Contains(err.Error(), "element not found") {
			return false, nil
		}
		return false, errors.NewBrowserError(fmt.Sprintf("error checking if element exists: %s", selector), err)
	}
	
	return element != nil, nil
}

// ClickElement clicks on an element matching the provided selector
func (b *RodBrowser) ClickElement(selector string) error {
	if b.page == nil {
		return errors.NewBrowserError("browser page not initialized, call Open first", nil)
	}
	
	// Get the element
	element, err := b.GetElement(selector)
	if err != nil {
		return err
	}
	
	// Scroll element into view
	err = element.ScrollIntoView()
	if err != nil {
		return errors.NewBrowserError(fmt.Sprintf("failed to scroll element into view: %s", selector), err)
	}
	
	// Click the element (left button, 1 click)
	err = element.Click(proto.InputMouseButtonLeft, 1)
	if err != nil {
		return errors.NewBrowserError(fmt.Sprintf("failed to click element: %s", selector), err)
	}
	
	b.log.Debug("Clicked element with selector: %s", selector)
	return nil
}

// GetElementText returns the text content of an element
func (b *RodBrowser) GetElementText(selector string) (string, error) {
	if b.page == nil {
		return "", errors.NewBrowserError("browser page not initialized, call Open first", nil)
	}
	
	// Get the element
	element, err := b.GetElement(selector)
	if err != nil {
		return "", err
	}
	
	// Get the element's text
	text, err := element.Text()
	if err != nil {
		return "", errors.NewBrowserError(fmt.Sprintf("failed to get text from element: %s", selector), err)
	}
	
	return text, nil
}

// GetElementAttribute returns the value of an attribute on an element
func (b *RodBrowser) GetElementAttribute(selector, attr string) (string, error) {
	if b.page == nil {
		return "", errors.NewBrowserError("browser page not initialized, call Open first", nil)
	}
	
	// Get the element
	element, err := b.GetElement(selector)
	if err != nil {
		return "", err
	}
	
	// Get the attribute value
	value, err := element.Attribute(attr)
	if err != nil {
		return "", errors.NewBrowserError(fmt.Sprintf("failed to get attribute '%s' from element: %s", attr, selector), err)
	}
	
	if value == nil {
		return "", nil // Attribute doesn't exist
	}
	
	return *value, nil
}

// WaitForElement waits for an element to appear in the page
func (b *RodBrowser) WaitForElement(selector string, timeout time.Duration) error {
	if b.page == nil {
		return errors.NewBrowserError("browser page not initialized, call Open first", nil)
	}
	
	if timeout == 0 {
		timeout = 10 * time.Second // Default timeout
	}
	
	// Set timeout for this operation
	b.page = b.page.Timeout(timeout)
	
	// Wait for the element to appear
	err := b.page.Timeout(timeout).WaitElementsMoreThan(selector, 0)
	if err != nil {
		return errors.NewBrowserError(fmt.Sprintf("timeout waiting for element: %s", selector), err)
	}
	
	b.log.Debug("Element appeared: %s", selector)
	return nil
}

// WaitForNavigation waits for page navigation to complete
func (b *RodBrowser) WaitForNavigation(timeout time.Duration) error {
	if b.page == nil {
		return errors.NewBrowserError("browser page not initialized, call Open first", nil)
	}
	
	if timeout == 0 {
		timeout = 30 * time.Second // Default timeout
	}
	
	b.log.Debug("Waiting for page navigation (timeout: %v)...", timeout)
	
	// First try with WaitLoad which is more reliable
	err := b.page.Timeout(timeout).WaitLoad()
	if err == nil {
		b.log.Debug("Navigation completed successfully")
		return nil
	}
	
	// If WaitLoad fails, try with WaitIdle as a fallback
	// This handles cases where the page is still processing after initial load
	b.log.Debug("WaitLoad failed, trying WaitIdle: %v", err)
	err = b.page.Timeout(timeout).WaitIdle(timeout)
	if err == nil {
		b.log.Debug("Navigation completed with WaitIdle")
		return nil
	}
	
	// As a last resort, just wait a fixed time
	b.log.Debug("Navigation waiting failed, using fixed delay: %v", err)
	time.Sleep(timeout / 2)
	
	// Check if we can still interact with the page
	_, err = b.page.Element("body")
	if err != nil {
		return errors.NewBrowserError("timeout waiting for navigation", err)
	}
	
	b.log.Debug("Navigation assumed complete after delay")
	return nil
}

// ExtractLinks extracts all links (anchor elements) matching the selector
func (b *RodBrowser) ExtractLinks(selector string) ([]LinkData, error) {
	if b.page == nil {
		return nil, errors.NewBrowserError("browser page not initialized, call Open first", nil)
	}
	
	// Get elements matching selector
	elements, err := b.GetElements(selector)
	if err != nil {
		return nil, err
	}
	
	var links []LinkData
	
	// Process each element
	for i, element := range elements {
		link := LinkData{
			Attributes: make(map[string]string),
		}
		
		// Extract text
		text, err := element.Text()
		if err == nil {
			link.Text = strings.TrimSpace(text)
		} else {
			b.log.Warn("Could not extract text from link %d: %v", i, err)
		}
		
		// Extract href attribute
		href, err := element.Attribute("href")
		if err == nil && href != nil {
			link.URL = *href
		} else {
			b.log.Warn("Could not extract href from link %d: %v", i, err)
		}
		
		// Extract other common attributes
		for _, attr := range []string{"id", "title", "class", "rel", "target"} {
			value, err := element.Attribute(attr)
			if err == nil && value != nil {
				link.Attributes[attr] = *value
			}
		}
		
		links = append(links, link)
	}
	
	b.log.Debug("Extracted %d links matching selector: %s", len(links), selector)
	return links, nil
}