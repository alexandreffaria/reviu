package result

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/alexandreffaria/reviu/internal/browser"
	"github.com/alexandreffaria/reviu/internal/errors"
	"github.com/alexandreffaria/reviu/internal/logger"
)

// Constants for CSS selectors and pagination settings
const (
	ResultLinkSelector  = "a.titulo-busca"
	NextPageSelector    = "button.br-button.circle.page-buscador[aria-label=\"Página seguinte\"]"
	ResultCountSelector = "span.fw-semibold.text-up-01.text-gray-60"
	ResultsPerPage      = 30 // Number of results per page

	DetailYearSelector   = "#item-ano"
	DetailAuthorSelector = "a.view-autor"
)

// CAPESResultExtractor extracts search results from CAPES search pages
type CAPESResultExtractor struct {
	log        logger.Logger
	browser    browser.Browser
	options    ProcessorOptions
	collection *SearchCollection
}

// NewCAPESResultExtractor creates a new extractor
func NewCAPESResultExtractor(browser browser.Browser, log logger.Logger) *CAPESResultExtractor {
	if log == nil {
		log = logger.NewLogger() // Default logger
	}

	return &CAPESResultExtractor{
		log:        log.WithPrefix("Extractor"),
		browser:    browser,
		options:    DefaultProcessorOptions(),
		collection: nil,
	}
}

// SetOptions configures the extractor options
func (e *CAPESResultExtractor) SetOptions(options ProcessorOptions) {
	e.options = options
}

// extractTotalResults extracts the total number of search results from the page
func (e *CAPESResultExtractor) extractTotalResults() (int, error) {
	// Get the text from the result count element
	resultCountText, err := e.browser.GetElementText(ResultCountSelector)
	if err != nil {
		return 0, errors.NewBrowserError("failed to find result count element", err)
	}

	// Parse the text to extract the number
	// The text format is typically like "3.016 resultados"
	resultCountText = strings.Replace(resultCountText, ".", "", -1) // Remove thousands separator
	var count int
	_, err = fmt.Sscanf(resultCountText, "%d resultados", &count)
	if err != nil {
		e.log.Warn("Failed to parse result count from '%s': %v", resultCountText, err)
		// Return a default value
		return 100, nil
	}

	return count, nil
}

// buildPageURL constructs a URL for a specific page
func (e *CAPESResultExtractor) buildPageURL(baseURL string, page int) string {
	// Check if the URL already has query parameters
	if strings.Contains(baseURL, "?") {
		// If URL already has parameters, add the page parameter
		if strings.Contains(baseURL, "page=") {
			// Replace existing page parameter
			re := regexp.MustCompile(`page=\d+`)
			return re.ReplaceAllString(baseURL, fmt.Sprintf("page=%d", page))
		} else {
			// Add page parameter
			return fmt.Sprintf("%s&page=%d", baseURL, page)
		}
	} else {
		// URL has no parameters, add page as first parameter
		return fmt.Sprintf("%s?page=%d", baseURL, page)
	}
}

// Process extracts search results from all pages using URL-based pagination
func (e *CAPESResultExtractor) Process(ctx context.Context, searchTerm string, searchURL string) (*SearchCollection, error) {
	// Initialize collection
	e.collection = NewSearchCollection(searchTerm)

	// Create a context with timeout
	if e.options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(e.options.Timeout)*time.Second)
		defer cancel()
	}

	// Navigate to the initial search URL
	e.log.Info("Navigating to initial search URL")
	if err := e.browser.Open(searchURL); err != nil {
		return nil, errors.NewBrowserError("failed to open initial search URL", err)
	}

	// Extract total results to calculate total pages
	totalResults, err := e.extractTotalResults()
	if err != nil {
		e.log.Warn("Could not determine total results: %v", err)
		totalResults = 100 // Default value
	}

	// Calculate total pages
	totalPages := (totalResults + ResultsPerPage - 1) / ResultsPerPage
	e.log.Info("Found approximately %d total results across %d pages", totalResults, totalPages)

	// Determine max pages to process
	maxPagesToProcess := totalPages
	if e.options.MaxPages > 0 && e.options.MaxPages < totalPages {
		maxPagesToProcess = e.options.MaxPages
		e.log.Info("Will process up to %d pages as specified by max-pages parameter", maxPagesToProcess)
	}

	// Process all pages using URL pagination
	for currentPage := 1; currentPage <= maxPagesToProcess; currentPage++ {
		select {
		case <-ctx.Done():
			e.log.Warn("Processing stopped due to context cancellation or timeout")
			return e.collection, ctx.Err()
		default:
			// Continue processing
		}

		pageURL := searchURL
		// For the first page, we're already on the correct page
		if currentPage > 1 {
			// Navigate to the specific page using URL parameter
			pageURL = e.buildPageURL(searchURL, currentPage)
			e.log.Info("Navigating to page %d using URL: %s", currentPage, pageURL)

			// Close the previous browser to avoid resource leaks
			if err := e.browser.Close(); err != nil {
				e.log.Warn("Error closing previous browser instance: %v", err)
			}

			// Open a new browser for this page
			if err := e.browser.Open(pageURL); err != nil {
				e.log.Error("Failed to open page %d: %v", currentPage, err)
				break
			}
		}

		// Log current page
		e.log.Info("Processing page %d", currentPage)

		// Extract results from current page
		results, err := e.extractResultsFromCurrentPage(currentPage, pageURL)
		if err != nil {
			e.log.Error("Failed to extract results from page %d: %v", currentPage, err)
			// Continue to next page despite errors
		} else {
			// Add results to collection
			e.collection.AddResults(results)
			e.log.Info("Extracted %d results from page %d", len(results), currentPage)
		}

		// Update collection metadata
		e.collection.UpdatePageCount(currentPage)

		// Delay between page navigations to avoid being blocked
		if currentPage < maxPagesToProcess {
			if e.options.PageDelay > 0 {
				e.log.Info("Waiting %v between pages to avoid blocking...", e.options.PageDelay)
				time.Sleep(e.options.PageDelay)
			}
		}
	}

	e.log.Info("Finished processing %d pages with a total of %d results",
		e.collection.TotalPages, e.collection.TotalResults)

	return e.collection, nil
}

// extractResultsFromCurrentPage extracts results from the current page
func (e *CAPESResultExtractor) extractResultsFromCurrentPage(pageNum int, pageURL string) ([]SearchResult, error) {
	// Get all result links on the page
	links, err := e.browser.ExtractLinks(ResultLinkSelector)
	if err != nil {
		return nil, errors.NewBrowserError("failed to extract result links", err)
	}

	if len(links) == 0 {
		e.log.Warn("No results found on page %d", pageNum)
		return []SearchResult{}, nil
	}

	// Process each link into a search result
	results := make([]SearchResult, 0, len(links))

	for i, link := range links {
		// Create result from link
		result := SearchResult{
			Title:     cleanTitle(link.Text),
			URL:       absoluteURL(link.URL),
			ID:        extractIDFromURL(link.URL),
			Source:    "CAPES",
			PageFound: pageNum,
			Position:  i + 1,
		}

		// Navigate to the detail page to extract author and year metadata
		author, year := e.extractMetadataForResult(result.URL)
		result.Author = author
		result.Year = year

		results = append(results, result)
	}

	return results, nil
}

// extractMetadataForResult navigates to the publication page and collects metadata
// using a dedicated browser instance so the main search page remains untouched.
func (e *CAPESResultExtractor) extractMetadataForResult(detailURL string) (string, string) {
	if detailURL == "" {
		return "", ""
	}

	// Use a separate headless browser for detail extraction to avoid disrupting the
	// main search page state while still visiting every article page for metadata.
	detailBrowserOptions := browser.DefaultBrowserOptions
	detailBrowserOptions.Headless = true
	detailBrowser := browser.NewBrowser(e.log, &detailBrowserOptions)

	if err := detailBrowser.Open(detailURL); err != nil {
		e.log.Warn("Failed to open details page %s: %v", detailURL, err)
		return "", ""
	}

	defer func() {
		if err := detailBrowser.Close(); err != nil {
			e.log.Warn("Failed to close detail browser for %s: %v", detailURL, err)
		}
	}()

	timeout := time.Duration(e.options.PageTimeout) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	// Wait for the details to load
	if err := detailBrowser.WaitForElement(DetailYearSelector, timeout); err != nil {
		e.log.Debug("Year element not found on detail page %s: %v", detailURL, err)
	}

	author := e.extractAuthorsFromDetail(detailBrowser)
	year := e.extractYearFromDetail(detailBrowser)

	return author, year
}

// extractAuthorsFromDetail collects author names from the details page
func (e *CAPESResultExtractor) extractAuthorsFromDetail(detailBrowser browser.Browser) string {
	authorElements, err := detailBrowser.GetElements(DetailAuthorSelector)
	if err != nil {
		e.log.Warn("Could not extract authors from detail page: %v", err)
		return ""
	}

	var authors []string
	for _, element := range authorElements {
		name, err := element.Text()
		if err != nil {
			continue
		}

		name = strings.TrimSpace(name)
		if name != "" {
			authors = append(authors, name)
		}
	}

	return strings.Join(authors, ", ")
}

// extractYearFromDetail collects the publication year from the details page
func (e *CAPESResultExtractor) extractYearFromDetail(detailBrowser browser.Browser) string {
	yearText, err := detailBrowser.GetElementText(DetailYearSelector)
	if err != nil {
		e.log.Warn("Could not extract year from detail page: %v", err)
		return ""
	}

	year := strings.TrimSpace(yearText)
	year = strings.TrimSuffix(year, ";")
	return strings.TrimSpace(year)
}

// hasNextPage checks if there's a next page button
func (e *CAPESResultExtractor) hasNextPage() (bool, error) {
	// Check if next page button exists
	exists, err := e.browser.ElementExists(NextPageSelector)
	if err != nil {
		return false, errors.NewBrowserError("failed to check for next page button", err)
	}

	return exists, nil
}

// goToNextPage clicks the next page button with retry logic
func (e *CAPESResultExtractor) goToNextPage() error {
	// Get configuration values from options
	maxRetries := e.options.RetryAttempts
	if maxRetries <= 0 {
		maxRetries = 3 // Fallback if not properly configured
	}

	baseTimeout := time.Duration(e.options.NavigationTimeout) * time.Second
	if baseTimeout <= 0 {
		baseTimeout = 20 * time.Second // Fallback if not properly configured
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		e.log.Debug("Pagination attempt %d of %d", attempt, maxRetries)

		// First, scroll to ensure the button is in view (lazy-loading)
		e.log.Debug("Scrolling to ensure next page button is loaded")

		// Try scrolling to bottom first
		if err := e.browser.ScrollToBottom(); err != nil {
			e.log.Warn("Error scrolling to bottom: %v", err)
			// Continue anyway
		}

		// Then continuous scrolling to trigger lazy loading
		e.log.Debug("Performing continuous scrolling for 3 seconds")
		if err := e.browser.ScrollForDuration(3 * time.Second); err != nil {
			e.log.Warn("Error during continuous scrolling: %v", err)
			// Continue anyway
		}

		// Small delay after scrolling
		time.Sleep(1 * time.Second)

		// Click next page button
		if err := e.browser.ClickElement(NextPageSelector); err != nil {
			e.log.Warn("Failed to click next page button (attempt %d): %v", attempt, err)
			if attempt == maxRetries {
				return errors.NewBrowserError("failed to click next page button after multiple attempts", err)
			}
			time.Sleep(1 * time.Second)
			continue
		}

		// Increase timeout for each retry
		timeout := baseTimeout + time.Duration(attempt-1)*5*time.Second

		// Wait for navigation with the configured timeout
		navigationTimeout := time.Duration(e.options.NavigationTimeout) * time.Second
		if navigationTimeout <= 0 {
			navigationTimeout = timeout // Use fallback if not configured
		}

		if err := e.browser.WaitForNavigation(navigationTimeout); err != nil {
			e.log.Warn("Failed waiting for navigation (attempt %d): %v", attempt, err)
			if attempt == maxRetries {
				return errors.NewBrowserError("failed waiting for navigation after multiple attempts", err)
			}
			continue
		}

		// Wait for results to load using page timeout
		resultTimeout := time.Duration(e.options.PageTimeout) * time.Second
		if resultTimeout <= 0 {
			resultTimeout = timeout + 5*time.Second // Use fallback if not configured
		}

		if err := e.browser.WaitForElement(ResultLinkSelector, resultTimeout); err != nil {
			e.log.Warn("Failed waiting for results to load (attempt %d): %v", attempt, err)
			if attempt == maxRetries {
				return errors.NewBrowserError("failed waiting for results to load after multiple attempts", err)
			}
			continue
		}

		// Successful navigation
		e.log.Debug("Successfully navigated to next page on attempt %d", attempt)

		// Add a small delay to ensure page is stable
		time.Sleep(1 * time.Second)
		return nil
	}

	return errors.NewBrowserError("failed to navigate to next page after all retry attempts", nil)
}

// Helper functions

// cleanTitle removes extra whitespace and cleans up the title
func cleanTitle(title string) string {
	// Remove extra whitespace
	title = strings.TrimSpace(title)
	// Replace multiple spaces with a single space
	title = strings.Join(strings.Fields(title), " ")
	return title
}

// absoluteURL converts relative URLs to absolute URLs
func absoluteURL(urlStr string) string {
	if strings.HasPrefix(urlStr, "http") {
		return urlStr
	}

	// If it's a relative URL, make it absolute
	baseURL := "https://www.periodicos.capes.gov.br"
	if strings.HasPrefix(urlStr, "/") {
		return baseURL + urlStr
	}

	return baseURL + "/" + urlStr
}
