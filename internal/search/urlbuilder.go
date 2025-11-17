// Package search provides functionality for constructing and executing searches
package search

import (
	"fmt"
	"net/url"
	"strings"
	
	"github.com/alexandreffaria/reviu/internal/config"
	"github.com/alexandreffaria/reviu/internal/errors"
	"github.com/alexandreffaria/reviu/internal/logger"
)

// URLBuilder defines the interface for constructing search URLs
type URLBuilder interface {
	// BuildSearchURL constructs a complete search URL from validated parameters
	BuildSearchURL(params *config.SearchParams) (string, error)
}

// CAPESURLBuilder implements URLBuilder for the CAPES search portal
type CAPESURLBuilder struct {
	baseURL string
	log     logger.Logger
}

// NewCAPESURLBuilder creates a new URL builder for CAPES searches
func NewCAPESURLBuilder(log logger.Logger) *CAPESURLBuilder {
	return &CAPESURLBuilder{
		baseURL: "https://www.periodicos.capes.gov.br/index.php/acervo/buscador.html",
		log:     log,
	}
}

// BuildSearchURL constructs a CAPES search URL from validated parameters
func (b *CAPESURLBuilder) BuildSearchURL(params *config.SearchParams) (string, error) {
	if params == nil {
		return "", errors.NewConfigError("search parameters cannot be nil", nil)
	}
	
	if !params.Valid {
		return "", errors.NewConfigError("parameters must be validated before building URL", nil)
	}
	
	// Construct query parameters in the required order
	var urlParams []string
	
	// Add search term (required parameter)
	termEncoded := encodeSearchTerm(params.SearchTerm)
	urlParams = append(urlParams, "q="+termEncoded)
	
	// Add empty source parameter (required by CAPES)
	urlParams = append(urlParams, "source=")
	
	// Add optional parameters
	
	// Open Access parameter
	if params.AccessType != "" {
		openAccessParam := buildOpenAccessParam(params.AccessType)
		urlParams = append(urlParams, openAccessParam)
	}
	
	// Publication Type parameter
	if params.PublicationType != "" {
		typeParam := buildPublicationTypeParam(params.PublicationType)
		urlParams = append(urlParams, typeParam)
	}
	
	// Year parameters
	if params.YearMin > 0 {
		yearMinParam := fmt.Sprintf("publishyear_min%%5B%%5D=%d", params.YearMin)
		urlParams = append(urlParams, yearMinParam)
	}
	
	if params.EffectiveYearMax > 0 {
		yearMaxParam := fmt.Sprintf("publishyear_max%%5B%%5D=%d", params.EffectiveYearMax)
		urlParams = append(urlParams, yearMaxParam)
	}
	
	// Peer Review parameter
	if params.PeerReviewed != "" {
		peerReviewParam := buildPeerReviewParam(params.PeerReviewed)
		urlParams = append(urlParams, peerReviewParam)
	}
	
	// Language parameters
	for _, lang := range params.Languages {
		langParam := buildLanguageParam(lang)
		urlParams = append(urlParams, langParam)
	}
	
	// Construct final URL
	finalURL := b.baseURL + "?" + strings.Join(urlParams, "&")
	
	if b.log != nil {
		b.log.Debug("Built search URL: %s", finalURL)
	}
	
	return finalURL, nil
}

// Helper functions for parameter encoding and construction

// encodeSearchTerm properly encodes the search term for the CAPES portal
func encodeSearchTerm(term string) string {
	encoded := url.QueryEscape(term)
	// CAPES uses + instead of %20 for spaces
	return strings.ReplaceAll(encoded, "%20", "+")
}

// buildOpenAccessParam constructs the open access parameter
func buildOpenAccessParam(accessType string) string {
	if accessType == "sim" {
		return "open_access%5B%5D=open_access%3D%3D1"
	}
	return "open_access%5B%5D=open_access%3D%3D0"
}

// buildPublicationTypeParam constructs the publication type parameter
func buildPublicationTypeParam(pubType string) string {
	typeEncoded := url.QueryEscape("type==" + pubType)
	return "type%5B%5D=" + typeEncoded
}

// buildPeerReviewParam constructs the peer review parameter
func buildPeerReviewParam(peerReview string) string {
	if peerReview == "sim" {
		return "peer_reviewed%5B%5D=peer_reviewed%3D%3D1"
	}
	return "peer_reviewed%5B%5D=peer_reviewed%3D%3D0"
}

// buildLanguageParam constructs a language parameter
func buildLanguageParam(lang string) string {
	// Special handling for Portuguese and other diacritics
	langEncoded := strings.ReplaceAll(lang, "ê", "%C3%AA")
	return fmt.Sprintf("language%%5B%%5D=language%%3D%%3D%s", langEncoded)
}