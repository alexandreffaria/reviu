// Package config provides types and functions for application configuration
package config

import (
	"fmt"
	"time"
)

// SearchParams contains all possible search parameters
type SearchParams struct {
	// Required parameters
	SearchTerm string

	// Optional parameters
	AccessType     string // "sim", "nao", or "" (any)
	PublicationType string
	YearMin        int
	YearMax        int
	PeerReviewed   string // "sim", "nao", or "" (any)
	Languages      []string

	// Computed parameters (populated during validation)
	EffectiveYearMax int // Calculated max year value
	CurrentYear      int // Current year (for relative calculations)
	Valid            bool // Indicates if parameters have been validated
}

// AccessOption defines valid options for access type
type AccessOption string

const (
	AccessOpen   AccessOption = "sim"
	AccessClosed AccessOption = "nao"
	AccessAny    AccessOption = ""
)

// PeerReviewOption defines valid options for peer review status
type PeerReviewOption string

const (
	PeerReviewYes PeerReviewOption = "sim"
	PeerReviewNo  PeerReviewOption = "nao"
	PeerReviewAny PeerReviewOption = ""
)

// NewSearchParams creates a new SearchParams instance with current year set
func NewSearchParams() *SearchParams {
	return &SearchParams{
		CurrentYear: time.Now().Year(),
	}
}

// String returns a string representation of SearchParams for reporting
func (p *SearchParams) String() string {
	if p == nil {
		return "<nil>"
	}

	yearRange := "qualquer"
	if p.YearMin > 0 || p.EffectiveYearMax > 0 {
		minYear := "não especificado"
		maxYear := "não especificado"
		
		if p.YearMin > 0 {
			minYear = fmt.Sprintf("%d", p.YearMin)
		}
		
		if p.EffectiveYearMax > 0 {
			maxYear = fmt.Sprintf("%d", p.EffectiveYearMax)
		}
		
		yearRange = minYear + " até " + maxYear
	}

	access := "qualquer"
	if p.AccessType != "" {
		access = p.AccessType
	}

	peerReview := "qualquer"
	if p.PeerReviewed != "" {
		peerReview = p.PeerReviewed
	}

	pubType := "qualquer"
	if p.PublicationType != "" {
		pubType = p.PublicationType
	}

	languages := "qualquer"
	if len(p.Languages) > 0 {
		languages = ""
		for i, lang := range p.Languages {
			if i > 0 {
				languages += ", "
			}
			languages += lang
		}
	}

	return "SearchParams{" +
		"SearchTerm: " + p.SearchTerm +
		", AccessType: " + access +
		", PublicationType: " + pubType +
		", YearRange: " + yearRange +
		", PeerReviewed: " + peerReview +
		", Languages: " + languages +
		", Valid: " + fmt.Sprintf("%v", p.Valid) +
		"}"
}