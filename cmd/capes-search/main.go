// Main package for the CAPES search application
package main

import (
	stderrors "errors" // standard library errors for As function
	"fmt"
	"os"
	"time"

	"github.com/alexandreffaria/reviu/internal/browser"
	"github.com/alexandreffaria/reviu/internal/cli"
	"github.com/alexandreffaria/reviu/internal/config"
	"github.com/alexandreffaria/reviu/internal/errors"
	"github.com/alexandreffaria/reviu/internal/logger"
	"github.com/alexandreffaria/reviu/internal/result"
	"github.com/alexandreffaria/reviu/internal/search"
)

func main() {
	// Initialize logger
	log := logger.NewLogger(logger.WithLevel(logger.INFO))
	log.Info("Starting CAPES Search Tool")

	// Run the application and handle errors
	if err := run(log); err != nil {
		// Determine error handling based on error type
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			switch appErr.Type {
			case errors.Configuration:
				log.Error("Configuration error: %v", err)
				fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
				os.Exit(1)

			case errors.UserInput:
				log.Error("Input error: %v", err)
				fmt.Fprintf(os.Stderr, "Input error: %v\n", err)
				// Show usage information
				os.Exit(2)

			case errors.Browser:
				log.Error("Browser error: %v", err)
				fmt.Fprintf(os.Stderr, "Browser error: %v\n", err)
				os.Exit(3)

			default:
				log.Error("Application error: %v", err)
				fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
				os.Exit(1)
			}
		} else {
			log.Error("Unexpected error: %v", err)
			fmt.Fprintf(os.Stderr, "Unexpected error: %v\n", err)
			os.Exit(1)
		}
	}

	log.Info("Application completed successfully")
}

// run contains the main application logic
func run(log logger.Logger) error {
	// Create component-specific loggers
	cliLog := log.WithPrefix("CLI")
	configLog := log.WithPrefix("Config")
	searchLog := log.WithPrefix("Search")
	browserLog := log.WithPrefix("Browser")
	resultLog := log.WithPrefix("Result")

	// Initialize CLI
	cli := cli.NewCLI(cliLog)

	// Parse command-line flags
	configLog.Info("Parsing command-line flags")
	params := config.SetupFlags(configLog)

	// Ensure required parameters are provided
	configLog.Debug("Ensuring required parameters")
	if err := cli.EnsureRequiredParameters(params); err != nil {
		return err
	}

	// Validate parameters
	configLog.Debug("Validating parameters")
	validator := &config.DefaultValidator{}
	if err := validator.ValidateSearchParams(params); err != nil {
		return err
	}

	// Print search report
	cli.PrintSearchReport(params)

	// Create URL builder
	urlBuilder := search.NewCAPESURLBuilder(searchLog)

	// Build search URL
	searchLog.Info("Building search URL")
	searchURL, err := urlBuilder.BuildSearchURL(params)
	if err != nil {
		return err
	}

	// Log the URL
	searchLog.Info("Search URL: %s", searchURL)
	cli.PrintSearchURL(searchURL)

	// Initialize browser
	browserLog.Info("Initializing browser")
	// Configure browser options based on parameters
	browserOptions := browser.DefaultBrowserOptions
	
	// Apply user-configured options
	browserOptions = browserOptions.
		WithStealthMode(params.StealthMode).
		WithRandomUserAgent(params.RandomUserAgent).
		WithSlowMotion(params.SlowMotion)
	
	// Set proxy if provided
	if params.Proxy != "" {
		browserOptions = browserOptions.WithProxy(params.Proxy)
	}
	
	// Create the browser instance with configured options
	browserLog.Info("Creating browser with anti-blocking measures")
	if params.StealthMode {
		browserLog.Info("Stealth mode enabled to avoid detection")
	}
	
	browser := browser.NewBrowser(browserLog, &browserOptions)

	// Ensure browser is closed even if errors occur
	defer func() {
		browserLog.Info("Closing browser")
		if err := browser.Close(); err != nil {
			log.Error("Failed to close browser: %v", err)
		}
	}()

	// Log browser anti-blocking configuration
	browserLog.Info("Browser configuration: stealth=%v, random-ua=%v, delay=%v, proxy=%s",
		params.StealthMode, params.RandomUserAgent, params.SlowMotion,
		params.Proxy)
	
	// Determine if we're doing a simple view or exporting results
	if params.ExportResults && params.OutputFile != "" {
		// We're exporting results - use the result processor
		resultLog.Info("Starting result export to %s", params.OutputFile)
		cli.PrintBrowserInfo(fmt.Sprintf("Iniciando exportação de resultados para: %s", params.OutputFile))
		cli.PrintBrowserInfo("Este processo pode demorar alguns minutos dependendo do número de resultados...")

		// Create result processor
		processor := result.NewResultProcessor(browser, resultLog)
		
		// Set browser to headless mode for export (optional)
		// This could be made configurable with a flag
		//browser.WithHeadless(true)
		
		// Process and export results
		err := processor.ProcessSearchResults(params, searchURL)
		if err != nil {
			return err
		}
		
		// Show success message
		cli.PrintBrowserInfo(fmt.Sprintf("Exportação concluída com sucesso para: %s", params.OutputFile))
		cli.PrintBrowserInfo("Você pode abrir o arquivo CSV em um editor de planilhas como Excel ou LibreOffice Calc.")

		return nil
	} else {
		// Simple view mode - just open the browser to show results
		cli.PrintBrowserInfo("Abrindo navegador com a URL de busca...")
		if err := browser.Open(searchURL); err != nil {
			return err
		}

		// Keep browser open for viewing results
		cli.PrintBrowserInfo("Busca realizada com sucesso.")
		cli.PrintBrowserInfo("Mantendo navegador aberto por 30 segundos para visualização dos resultados.")

		return browser.Wait(30 * time.Second)
	}
}