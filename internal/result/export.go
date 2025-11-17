package result

import (
	"fmt"
	"path/filepath"
	
	"github.com/alexandreffaria/reviu/internal/logger"
)

// ExportFormat defines the format for exporting search results
type ExportFormat string

const (
	FormatCSV  ExportFormat = "csv"
	FormatJSON ExportFormat = "json"
	FormatText ExportFormat = "txt"
)

// ExportConfig holds configuration for the export process
type ExportConfig struct {
	// File path for export
	FilePath string
	
	// Format to use for export
	Format ExportFormat
	
	// CSV-specific options
	Delimiter   rune   // Character to use as delimiter in CSV
	IncludeHeader bool  // Whether to include header row in CSV
	
	// Encoding options
	CharacterEncoding string // e.g., "utf-8", "iso-8859-1"
}

// DefaultCSVConfig returns a default configuration for CSV export
func DefaultCSVConfig(filePath string) ExportConfig {
	return ExportConfig{
		FilePath:          filePath,
		Format:            FormatCSV,
		Delimiter:         ',',
		IncludeHeader:     true,
		CharacterEncoding: "utf-8",
	}
}

// ExportStats captures statistics about an export operation
type ExportStats struct {
	StartTime       string
	EndTime         string
	Duration        string
	TotalResults    int
	ResultsWritten  int
	BytesWritten    int64
	ErrorCount      int
	FilePath        string
}

// String returns a formatted string with export statistics
func (s *ExportStats) String() string {
	return fmt.Sprintf(
		"Export completed in %s. Wrote %d/%d results (%d bytes) to %s with %d errors.",
		s.Duration,
		s.ResultsWritten,
		s.TotalResults,
		s.BytesWritten,
		s.FilePath,
		s.ErrorCount,
	)
}

// ResultWriter defines a generic interface for exporting search results
type ResultWriter interface {
	// Initialize prepares the writer
	Initialize() error
	
	// WriteHeader writes the header row
	WriteHeader() error
	
	// WriteResult writes a single result
	WriteResult(result SearchResult) error
	
	// WriteResults writes multiple results
	WriteResults(results []SearchResult) error
	
	// WriteCollection writes an entire search collection
	WriteCollection(collection *SearchCollection) error
	
	// Close finalizes the export and releases resources
	Close() error
}

// NewWriter creates the appropriate ResultWriter based on export config
func NewWriter(config ExportConfig, log logger.Logger) (ResultWriter, error) {
	// Ensure the file extension matches the format
	config.FilePath = ensureExtension(config.FilePath, string(config.Format))

	switch config.Format {
	case FormatCSV:
		return NewCSVWriter(config, log)
	case FormatJSON, FormatText:
		// Placeholder for future implementation
		return nil, fmt.Errorf("format %s not yet implemented", config.Format)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", config.Format)
	}
}

// ensureExtension ensures the filepath has the correct extension
func ensureExtension(filePath, ext string) string {
	currentExt := filepath.Ext(filePath)
	if currentExt == "" {
		return filePath + "." + ext
	}
	
	// If the extension doesn't match, replace it
	if currentExt[1:] != ext {
		return filePath[:len(filePath)-len(currentExt)] + "." + ext
	}
	
	return filePath
}