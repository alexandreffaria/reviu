// Package result provides functionality for handling search results
package result

import (
	"fmt"
	"strings"
	"time"
)

// SearchResult represents a single research publication from search results
type SearchResult struct {
	// Core fields extracted directly from search results
	Title string // Publication title
	URL   string // URL to details page
	ID    string // Document ID (extracted from URL)

	// Detailed metadata extracted from the publication page
	Author string // Author name(s) extracted from the details page
	Year   string // Publication year

	// Additional metadata that might be available
	Source string // Source of the publication, if available

	// Collection metadata
	PageFound int // The page number where this result was found
	Position  int // Position in the result list (1-based)
}

// NewSearchResult creates a new search result with the given title and URL
func NewSearchResult(title, url string, pageNum, position int) SearchResult {
	// Extract ID from URL if possible
	id := extractIDFromURL(url)

	return SearchResult{
		Title:     title,
		URL:       url,
		ID:        id,
		PageFound: pageNum,
		Position:  position,
	}
}

// String returns a formatted string representation of the search result
func (r SearchResult) String() string {
	return fmt.Sprintf("%s [Page %d, Pos %d] - %s", r.Title, r.PageFound, r.Position, r.URL)
}

// SearchCollection represents a collection of search results from a search session
type SearchCollection struct {
	// Search metadata
	SearchTerm   string    // The search term used
	SearchDate   time.Time // When the search was performed
	TotalPages   int       // Total number of pages processed
	TotalResults int       // Total number of results collected

	// The actual results
	Results []SearchResult // All search results collected
}

// NewSearchCollection creates a new search collection
func NewSearchCollection(searchTerm string) *SearchCollection {
	return &SearchCollection{
		SearchTerm: searchTerm,
		SearchDate: time.Now(),
		Results:    make([]SearchResult, 0),
	}
}

// AddResult adds a search result to the collection
func (c *SearchCollection) AddResult(result SearchResult) {
	c.Results = append(c.Results, result)
	c.TotalResults = len(c.Results)
}

// AddResults adds multiple search results to the collection
func (c *SearchCollection) AddResults(results []SearchResult) {
	c.Results = append(c.Results, results...)
	c.TotalResults = len(c.Results)
}

// UpdatePageCount updates the total page count if the new count is higher
func (c *SearchCollection) UpdatePageCount(pageCount int) {
	if pageCount > c.TotalPages {
		c.TotalPages = pageCount
	}
}

// ResultsFromPage returns all results from a specific page
func (c *SearchCollection) ResultsFromPage(pageNum int) []SearchResult {
	var pageResults []SearchResult
	for _, result := range c.Results {
		if result.PageFound == pageNum {
			pageResults = append(pageResults, result)
		}
	}
	return pageResults
}

// extractIDFromURL extracts the document ID from the URL
// Example URL: "/index.php/acervo/buscador.html?task=detalhes&source=all&id=W2004342886"
func extractIDFromURL(urlStr string) string {
	// Find the position of "id=" in the URL
	idPos := strings.Index(urlStr, "id=")
	if idPos == -1 {
		return "" // No ID found
	}

	// Get everything after "id="
	idPart := urlStr[idPos+3:]

	// If there are other query parameters after this one, stop at the &
	if ampPos := strings.Index(idPart, "&"); ampPos != -1 {
		idPart = idPart[:ampPos]
	}

	return idPart
}
