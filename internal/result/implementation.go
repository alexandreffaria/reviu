package result

import (
	"context"
	"time"

	"github.com/alexandreffaria/reviu/internal/browser"
	"github.com/alexandreffaria/reviu/internal/config"
	"github.com/alexandreffaria/reviu/internal/errors"
	"github.com/alexandreffaria/reviu/internal/logger"
)

// MainResultProcessor coordinates the extraction and export of search results
type MainResultProcessor struct {
	log       logger.Logger
	extractor *CAPESResultExtractor
	options   ProcessorOptions
}

// NewResultProcessor creates a new processor
func NewResultProcessor(browser browser.Browser, log logger.Logger) *MainResultProcessor {
	if log == nil {
		log = logger.NewLogger() // Default logger
	}
	
	return &MainResultProcessor{
		log:       log.WithPrefix("Processor"),
		extractor: NewCAPESResultExtractor(browser, log),
		options:   DefaultProcessorOptions(),
	}
}

// SetOptions configures the processor options
func (p *MainResultProcessor) SetOptions(options ProcessorOptions) {
	p.options = options
	p.extractor.SetOptions(options)
}

// SetLogger sets the logger for the processor
func (p *MainResultProcessor) SetLogger(log logger.Logger) {
	if log != nil {
		p.log = log.WithPrefix("Processor")
		p.extractor.log = log.WithPrefix("Extractor")
	}
}

// ProcessAndExport extracts results and exports them to the configured format
func (p *MainResultProcessor) ProcessAndExport(ctx context.Context, searchParams *config.SearchParams, searchURL string) error {
	// Create context for the entire operation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	// Start timing
	startTime := time.Now()
	
	// Extract results
	p.log.Info("Starting result extraction for search: %s", searchParams.SearchTerm)
	collection, err := p.extractor.Process(ctx, searchParams.SearchTerm, searchURL)
	if err != nil {
		return errors.NewBrowserError("failed during result extraction", err)
	}
	
	// If export is enabled, export the results
	if searchParams.OutputFile != "" {
		p.log.Info("Exporting %d results to %s", collection.TotalResults, searchParams.OutputFile)
		
		// Create export configuration
		exportConfig := ExportConfig{
			FilePath:          searchParams.OutputFile,
			Format:            FormatCSV,
			Delimiter:         ',',
			IncludeHeader:     true, // We'll always include headers for now
			CharacterEncoding: "utf-8",
		}
		
		// Create writer
		writer, err := NewWriter(exportConfig, p.log)
		if err != nil {
			return errors.NewConfigError("failed to create export writer", err)
		}
		
		// Initialize writer
		if err := writer.Initialize(); err != nil {
			return errors.NewConfigError("failed to initialize export writer", err)
		}
		
		// Ensure writer is closed when done
		defer func() {
			if err := writer.Close(); err != nil {
				p.log.Error("Failed to close export writer: %v", err)
			}
		}()
		
		// Export collection
		if err := writer.WriteCollection(collection); err != nil {
			return errors.NewExternalError("failed to export results", err)
		}
		
		// Report success
		duration := time.Since(startTime)
		p.log.Info("Successfully exported %d results from %d pages in %v", 
			collection.TotalResults, collection.TotalPages, duration)
	}
	
	return nil
}

// ProcessSearchResults is a convenience method that handles the entire process
func (p *MainResultProcessor) ProcessSearchResults(searchParams *config.SearchParams, searchURL string) error {
	// Create a background context
	ctx := context.Background()
	
	// Create processor options from search params
	options := ProcessorOptions{
		MaxPages:          searchParams.MaxPages,
		Timeout:           600, // 10 minutes default
		RetryAttempts:     3,
		PageTimeout:       30,  // 30 seconds per page
		NavigationTimeout: 30,  // 30 seconds for navigation
		PageDelay:         searchParams.PageDelay, // Use the delay specified in search params
	}
	
	// Set options
	p.SetOptions(options)
	
	// Process and export
	return p.ProcessAndExport(ctx, searchParams, searchURL)
}