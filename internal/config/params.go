package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/alexandreffaria/reviu/internal/errors"
)

// Validator provides methods to validate search parameters
type Validator interface {
	ValidateSearchParams(*SearchParams) error
}

// DefaultValidator implements parameter validation logic
type DefaultValidator struct{}

// ValidateSearchParams validates all search parameters
// Returns error if validation fails
func (v *DefaultValidator) ValidateSearchParams(params *SearchParams) error {
	if params == nil {
		return errors.NewConfigError("params cannot be nil", nil)
	}
	
	// Required parameter validation
	if params.SearchTerm == "" {
		return errors.NewConfigError("search term is required", nil)
	}
	
	// Validate and normalize access type
	if err := validateAccessType(params); err != nil {
		return err
	}
	
	// Validate and normalize peer review status
	if err := validatePeerReview(params); err != nil {
		return err
	}
	
	// Validate publication years
	if err := validateYears(params); err != nil {
		return err
	}
	
	// Normalize languages
	normalizeLanguages(params)
	
	// Validate export parameters if export is enabled
	if params.ExportResults {
		if err := validateExportParams(params); err != nil {
			return err
		}
	}
	
	// Mark params as validated
	params.Valid = true
	
	return nil
}

// validateAccessType validates and normalizes the access type parameter
func validateAccessType(params *SearchParams) error {
	if params.AccessType == "" {
		return nil // Empty value is valid
	}
	
	if params.AccessType != string(AccessOpen) && params.AccessType != string(AccessClosed) {
		return errors.NewConfigError(
			fmt.Sprintf("invalid access type value: %s (must be 'sim' or 'nao')", params.AccessType),
			nil,
		)
	}
	
	return nil
}

// validatePeerReview validates and normalizes the peer review parameter
func validatePeerReview(params *SearchParams) error {
	if params.PeerReviewed == "" {
		return nil // Empty value is valid
	}
	
	if params.PeerReviewed != string(PeerReviewYes) && params.PeerReviewed != string(PeerReviewNo) {
		return errors.NewConfigError(
			fmt.Sprintf("invalid peer review value: %s (must be 'sim' or 'nao')", params.PeerReviewed),
			nil,
		)
	}
	
	return nil
}

// validateYears validates and normalizes year parameters
func validateYears(params *SearchParams) error {
	// If no years specified, nothing to validate
	if params.YearMin == 0 && params.YearMax == 0 {
		return nil
	}
	
	currentYear := time.Now().Year()
	params.CurrentYear = currentYear
	
	// Validate minimum year if provided
	if params.YearMin < 0 {
		return errors.NewConfigError(
			fmt.Sprintf("invalid minimum year: %d (must be positive)", params.YearMin),
			nil,
		)
	}
	
	// Validate maximum year if provided
	if params.YearMax < 0 {
		return errors.NewConfigError(
			fmt.Sprintf("invalid maximum year: %d (must be positive)", params.YearMax),
			nil,
		)
	}
	
	// If only minimum year is specified, use current year as maximum
	if params.YearMin > 0 && params.YearMax == 0 {
		params.EffectiveYearMax = currentYear
	} else {
		params.EffectiveYearMax = params.YearMax
	}
	
	// Ensure minimum year is not greater than maximum year
	if params.YearMin > 0 && params.EffectiveYearMax > 0 && params.YearMin > params.EffectiveYearMax {
		return errors.NewConfigError(
			fmt.Sprintf("minimum year (%d) cannot be greater than maximum year (%d)",
				params.YearMin, params.EffectiveYearMax),
			nil,
		)
	}
	
	return nil
}

// normalizeLanguages ensures languages are properly formatted
func normalizeLanguages(params *SearchParams) {
	// Nothing to do if no languages
	if len(params.Languages) == 0 {
		return
	}
	
	// Trim whitespace from each language
	for i, lang := range params.Languages {
		params.Languages[i] = strings.TrimSpace(lang)
	}
}

// validateExportParams validates export-related parameters
func validateExportParams(params *SearchParams) error {
	// Validate output file
	if params.ExportResults && params.OutputFile == "" {
		return errors.NewConfigError("output file is required when export is enabled", nil)
	}
	
	// Validate export format
	if params.ExportFormat != "" && params.ExportFormat != "csv" {
		return errors.NewConfigError(
			fmt.Sprintf("unsupported export format: %s (only 'csv' is currently supported)",
						params.ExportFormat),
			nil,
		)
	}
	
	// Validate max pages
	if params.MaxPages < 0 {
		return errors.NewConfigError(
			fmt.Sprintf("invalid max pages: %d (must be 0 or positive)", params.MaxPages),
			nil,
		)
	}
	
	return nil
}