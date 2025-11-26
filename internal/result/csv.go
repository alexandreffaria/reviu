package result

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexandreffaria/reviu/internal/config"
	"github.com/alexandreffaria/reviu/internal/errors"
	"github.com/alexandreffaria/reviu/internal/logger"
)

// CSVHeader defines the column names for the CSV export
var CSVHeader = []string{
	"Título",
	"Autor",
	"Ano",
	"Link de acesso",
}

// SummaryCSVHeader defines the column names for the summary CSV export
var SummaryCSVHeader = []string{
	"Responsável",
	"Base de dados",
	"Termos de busca",
	"Data da busca",
	"No de artigos encontrados",
	"Filtros usados",
}

// CSVWriter implements ResultWriter for CSV format
type CSVWriter struct {
	config        ExportConfig
	file          *os.File
	writer        *csv.Writer
	log           logger.Logger
	rowCount      int
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

	// Convert result to row format with new structure
	row := []string{
		r.Title,  // Título
		r.Author, // Autor
		r.Year,   // Ano
		r.URL,    // Link de acesso
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

// WriteSummaryToCSV writes or appends a summary of the search to a CSV file
// The summary includes metadata about the search parameters and results
func WriteSummaryToCSV(collection *SearchCollection, params interface{}, outputPath string, log logger.Logger) error {
	if collection == nil {
		return errors.NewConfigError("search collection cannot be nil", nil)
	}

	// If not specified, create a default path by adding "_summary" before the extension
	if outputPath == "" {
		return errors.NewConfigError("output path for summary CSV is required", nil)
	}

	// Format current date in local time
	currentTime := collection.SearchDate.Local()
	formattedDate := currentTime.Format("02/01/2006")

	// Determine if file exists to decide whether to write header
	fileExists := false
	if _, err := os.Stat(outputPath); err == nil {
		fileExists = true
	}

	// Create directories if needed
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.NewConfigError(fmt.Sprintf("failed to create directory %s", dir), err)
		}
	}

	// Open file in append mode or create it if it doesn't exist
	file, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.NewConfigError(fmt.Sprintf("failed to open summary file %s", outputPath), err)
	}
	defer file.Close()

	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header if file is new
	if !fileExists {
		if err := writer.Write(SummaryCSVHeader); err != nil {
			return errors.NewExternalError("failed to write summary CSV header", err)
		}
	}

	// Extract filters from params if possible
	var filtersDescription string
	if configParams, ok := params.(*config.SearchParams); ok {
		filtersDescription = extractFiltersDescription(configParams)
	} else {
		filtersDescription = "Filtros não disponíveis"
	}

	// Create summary row
	summaryRow := []string{
		"",                    // Responsável (empty)
		"Periódicos Capes",    // Base de dados
		collection.SearchTerm, // Termos de busca
		formattedDate,         // Data da busca
		fmt.Sprintf("%d", collection.TotalResults), // No de artigos encontrados
		filtersDescription,                         // Filtros usados
	}

	// Write the summary row
	if err := writer.Write(summaryRow); err != nil {
		return errors.NewExternalError("failed to write summary CSV row", err)
	}

	if log != nil {
		log.Info("Search summary appended to %s", outputPath)
	}

	return nil
}

// extractFiltersDescription generates a human-readable description of the search filters in Portuguese
func extractFiltersDescription(params *config.SearchParams) string {
	var filters []string

	// Access Type
	if params.AccessType == "sim" {
		filters = append(filters, "Acesso aberto: Sim")
	} else if params.AccessType == "nao" {
		filters = append(filters, "Acesso aberto: Não")
	}

	// Publication Type
	if params.PublicationType != "" {
		filters = append(filters, fmt.Sprintf("Tipo de publicação: %s", params.PublicationType))
	}

	// Year Range
	if params.YearMin > 0 || params.EffectiveYearMax > 0 {
		yearStr := "Ano: "
		if params.YearMin > 0 {
			yearStr += fmt.Sprintf("%d", params.YearMin)
			if params.EffectiveYearMax > 0 {
				yearStr += fmt.Sprintf(" até %d", params.EffectiveYearMax)
			}
		} else if params.EffectiveYearMax > 0 {
			yearStr += fmt.Sprintf("Até %d", params.EffectiveYearMax)
		}
		filters = append(filters, yearStr)
	}

	// Peer Reviewed
	if params.PeerReviewed == "sim" {
		filters = append(filters, "Revisão por pares: Sim")
	} else if params.PeerReviewed == "nao" {
		filters = append(filters, "Revisão por pares: Não")
	}

	// Languages
	if len(params.Languages) > 0 {
		langStr := "Idiomas: " + strings.Join(params.Languages, ", ")
		filters = append(filters, langStr)
	}

	// Max Pages
	if params.MaxPages > 0 {
		filters = append(filters, fmt.Sprintf("Máximo de páginas: %d", params.MaxPages))
	}

	if len(filters) == 0 {
		return "Nenhum filtro aplicado"
	}

	return strings.Join(filters, "; ")
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
