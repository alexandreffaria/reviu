package result

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	
	"github.com/alexandreffaria/reviu/internal/errors"
	"github.com/alexandreffaria/reviu/internal/logger"
)

// CSVHeader defines the column names for the CSV export
var CSVHeader = []string{
	"Title",
	"URL",
	"ID",
	"Source",
	"PageFound",
	"Position",
}

// CSVWriter implements ResultWriter for CSV format
type CSVWriter struct {
	config       ExportConfig
	file         *os.File
	writer       *csv.Writer
	log          logger.Logger
	rowCount     int
	headerWritten bool
}

// NewCSVWriter creates a new CSV writer
func NewCSVWriter(config ExportConfig, log logger.Logger) (*CSVWriter, error) {
	if config.FilePath == "" {
		return nil, errors.NewConfigError("file path is required for CSV export", nil)
	}
	
	if log == nil {
		log = logger.NewLogger() // Default logger
	}
	
	return &CSVWriter{
		config: config,
		log:    log.WithPrefix("CSVExport"),
	}, nil
}

// Initialize opens the file and prepares the CSV writer
func (w *CSVWriter) Initialize() error {
	var err error
	
	// Create directories if they don't exist
	dir := filepath.Dir(w.config.FilePath)
	if dir != "" && dir != "." {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return errors.NewConfigError(fmt.Sprintf("failed to create directory %s", dir), err)
		}
	}
	
	// Open file for writing
	w.file, err = os.Create(w.config.FilePath)
	if err != nil {
		return errors.NewConfigError(fmt.Sprintf("failed to create file %s", w.config.FilePath), err)
	}
	
	// Create CSV writer
	w.writer = csv.NewWriter(w.file)
	
	// Set delimiter if custom one is specified
	if w.config.Delimiter != 0 && w.config.Delimiter != ',' {
		w.writer.Comma = w.config.Delimiter
	}
	
	w.log.Info("CSV export initialized: %s", w.config.FilePath)
	
	// Write header if configured
	if w.config.IncludeHeader {
		return w.WriteHeader()
	}
	
	return nil
}

// WriteHeader writes the header row to the CSV file
func (w *CSVWriter) WriteHeader() error {
	if w.writer == nil {
		return errors.NewConfigError("CSV writer not initialized, call Initialize first", nil)
	}
	
	if w.headerWritten {
		return nil // Header already written
	}
	
	err := w.writer.Write(CSVHeader)
	if err != nil {
		return errors.NewExternalError("failed to write CSV header", err)
	}
	
	w.writer.Flush() // Ensure header is written immediately
	w.headerWritten = true
	
	return w.writer.Error() // Check for delayed write errors
}

// WriteResult writes a single search result to the CSV file
func (w *CSVWriter) WriteResult(r SearchResult) error {
	if w.writer == nil {
		return errors.NewConfigError("CSV writer not initialized, call Initialize first", nil)
	}
	
	// Convert result to row format
	row := []string{
		r.Title,
		r.URL,
		r.ID,
		r.Source,
		strconv.Itoa(r.PageFound),
		strconv.Itoa(r.Position),
	}
	
	// Write the row
	err := w.writer.Write(row)
	if err != nil {
		return errors.NewExternalError("failed to write CSV row", err)
	}
	
	w.rowCount++
	
	// Periodically flush to avoid losing data in case of long-running processes
	if w.rowCount%10 == 0 {
		w.writer.Flush()
	}
	
	return nil
}

// WriteResults writes multiple results to the CSV file
func (w *CSVWriter) WriteResults(results []SearchResult) error {
	for _, r := range results {
		if err := w.WriteResult(r); err != nil {
			return err
		}
	}
	
	// Ensure data is written to disk
	w.writer.Flush()
	
	return w.writer.Error() // Check for delayed write errors
}

// WriteCollection writes an entire search collection to the CSV file
func (w *CSVWriter) WriteCollection(collection *SearchCollection) error {
	if collection == nil {
		return errors.NewConfigError("search collection cannot be nil", nil)
	}
	
	// Write all results
	err := w.WriteResults(collection.Results)
	if err != nil {
		return err
	}
	
	w.log.Info("Wrote %d search results to CSV", collection.TotalResults)
	
	return nil
}

// Close finalizes the CSV file and releases resources
func (w *CSVWriter) Close() error {
	if w.writer == nil {
		return nil // Nothing to close
	}
	
	// Flush any remaining data
	w.writer.Flush()
	
	// Check for write errors
	if err := w.writer.Error(); err != nil {
		return errors.NewExternalError("error flushing CSV data", err)
	}
	
	// Close the file
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return errors.NewExternalError("error closing CSV file", err)
		}
	}
	
	w.log.Info("CSV export completed: %s (%d rows)", w.config.FilePath, w.rowCount)
	
	return nil
}