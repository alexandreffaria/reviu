// Package cli provides command-line interface utilities
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/alexandreffaria/reviu/internal/config"
	"github.com/alexandreffaria/reviu/internal/errors"
	"github.com/alexandreffaria/reviu/internal/logger"
)

// CLI handles user interaction via command line
type CLI struct {
	reader *bufio.Reader
	log    logger.Logger
}

// NewCLI creates a new CLI instance
func NewCLI(log logger.Logger) *CLI {
	if log == nil {
		log = logger.NewLogger() // Use default logger if none provided
	}

	return &CLI{
		reader: bufio.NewReader(os.Stdin),
		log:    log.WithPrefix("CLI"),
	}
}

// PromptTextRequired asks for user input with a required value
func (c *CLI) PromptTextRequired(label, hint string) (string, error) {
	for {
		var prompt string
		if hint != "" {
			prompt = fmt.Sprintf("\n%s (%s): ", label, hint)
		} else {
			prompt = fmt.Sprintf("\n%s: ", label)
		}

		fmt.Print(prompt)
		input, err := c.reader.ReadString('\n')
		if err != nil {
			return "", errors.NewUserInputError("failed to read input", err)
		}

		input = strings.TrimSpace(input)
		if input != "" {
			return input, nil
		}

		fmt.Println("Campo obrigatório. Por favor, preencha.")
	}
}

// EnsureRequiredParameters prompts for any missing required parameters
func (c *CLI) EnsureRequiredParameters(params *config.SearchParams) error {
	if params == nil {
		return errors.NewConfigError("search parameters cannot be nil", nil)
	}

	// Ensure search term is provided
	if params.SearchTerm == "" {
		c.log.Info("Search term not provided via flags, prompting user")
		term, err := c.PromptTextRequired("TERMOS DE BUSCA", "texto livre (obrigatório)")
		if err != nil {
			return err
		}
		params.SearchTerm = term
	}

	return nil
}

// PrintSearchReport prints a formatted report of search parameters
func (c *CLI) PrintSearchReport(params *config.SearchParams) {
	if params == nil {
		c.log.Error("Cannot print report: params is nil")
		return
	}

	fmt.Println("\n========================================")
	fmt.Println(" RELATÓRIO DA BUSCA")
	fmt.Println("========================================")
	fmt.Printf("Termos de busca:   %s\n", params.SearchTerm)

	// Access type
	if params.AccessType != "" {
		fmt.Printf("Acesso aberto:     %s\n", params.AccessType)
	} else {
		fmt.Printf("Acesso aberto:     qualquer\n")
	}

	// Publication type
	if params.PublicationType != "" {
		fmt.Printf("Tipo de publicação: %s\n", params.PublicationType)
	} else {
		fmt.Printf("Tipo de publicação: qualquer\n")
	}

	// Publication years
	if params.YearMin > 0 || params.EffectiveYearMax > 0 {
		anoMinStr := "não especificado"
		anoMaxStr := "não especificado"

		if params.YearMin > 0 {
			anoMinStr = fmt.Sprintf("%d", params.YearMin)
		}

		if params.EffectiveYearMax > 0 {
			anoMaxStr = fmt.Sprintf("%d", params.EffectiveYearMax)
		}

		fmt.Printf("Anos de publicação: %s até %s\n", anoMinStr, anoMaxStr)
	} else {
		fmt.Printf("Anos de publicação: qualquer\n")
	}

	// Peer review
	if params.PeerReviewed != "" {
		fmt.Printf("Revisão por pares:  %s\n", params.PeerReviewed)
	} else {
		fmt.Printf("Revisão por pares:  qualquer\n")
	}

	// Languages
	if len(params.Languages) > 0 {
		fmt.Printf("Idiomas:            %s\n", strings.Join(params.Languages, ", "))
	} else {
		fmt.Printf("Idiomas:            qualquer\n")
	}
	fmt.Println("========================================\n")
}

// PrintSearchURL prints the generated search URL
func (c *CLI) PrintSearchURL(url string) {
	fmt.Println("URL da busca:", url)
}

// PrintBrowserInfo prints information about the browser status
func (c *CLI) PrintBrowserInfo(message string) {
	fmt.Println(message)
}