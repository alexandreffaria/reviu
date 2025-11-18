package config

import (
	"flag"
	"strings"
	"time"

	"github.com/alexandreffaria/reviu/internal/logger"
)

// Flag name constants are defined below instead of using a struct
// to avoid potential redefinition issues when flags are registered

// Flag constants
const (
	// Default flags
	searchTermFlag      = "search"
	accessTypeFlag      = "oa"
	publicationTypeFlag = "t"
	yearMinFlag         = "pymin"
	yearMaxFlag         = "pymax"
	peerReviewedFlag    = "pr"
	languagesFlag       = "lang"
	
	// Flags for output formatting
	outputFileFlag      = "output"
	formatFlag          = "format"
	maxPagesFlag        = "max-pages"
	noHeadersFlag       = "no-headers"
	
	// Browser options
	rodOptionsFlag      = "rod-options"
	stealthModeFlag     = "stealth"
	randomUserAgentFlag = "random-ua"
	slowMotionFlag      = "slow"
	proxyFlag           = "proxy"
	pageDelayFlag       = "delay"
)

// SetupFlags configures and parses command-line flags
// Returns a populated SearchParams struct
func SetupFlags(log logger.Logger) *SearchParams {
	params := NewSearchParams()
	
	// Define all flag pointers
	// Define flags using the constants - NOT the DefaultFlagNames struct
	searchTerm := flag.String(searchTermFlag, "",
	                            "Termo para pesquisar")
	accessType := flag.String(accessTypeFlag, "",
	                            "Acesso aberto: 'sim', 'nao' ou omitir para qualquer")
	publicationType := flag.String(publicationTypeFlag, "",
	                                 "Tipo de publicação (ex: 'Artigo')")
	yearMin := flag.Int(yearMinFlag, 0,
	                      "Ano mínimo de publicação")
	yearMax := flag.Int(yearMaxFlag, 0,
	                      "Ano máximo de publicação")
	peerReviewed := flag.String(peerReviewedFlag, "",
	                              "Revisão por pares: 'sim', 'nao' ou omitir para qualquer")
	languages := flag.String(languagesFlag, "",
	                           "Idiomas separados por '/' (ex: 'Português/Inglês/Espanhol')")
	
	// Export flags
	outputFile := flag.String(outputFileFlag, "",
	                            "Arquivo de saída para resultados (ex: 'resultados.csv')")
	exportFormat := flag.String(formatFlag, "csv",
	                              "Formato de exportação (csv)")
	maxPages := flag.Int(maxPagesFlag, 0,
	                       "Número máximo de páginas a processar (0 = todas)")
	noHeaders := flag.Bool(noHeadersFlag, false,
	                         "Não incluir linha de cabeçalho no arquivo CSV")
	
	// Browser anti-blocking options
	rodOptions := flag.String(rodOptionsFlag, "",
	                            "Set the default value of options used by rod.")
	stealthMode := flag.Bool(stealthModeFlag, true,
	                           "Enable stealth mode to avoid detection")
	randomUserAgent := flag.Bool(randomUserAgentFlag, true,
	                               "Use random user-agent string")
	slowMotion := flag.Duration(slowMotionFlag, 200*time.Millisecond,
	                              "Add delay between browser actions (e.g. '200ms')")
	pageDelay := flag.Duration(pageDelayFlag, 2*time.Second,
	                             "Delay between pages to avoid being blocked (e.g. '2s', '5s')")
	proxy := flag.String(proxyFlag, "",
	                       "Use proxy for browser (format: 'http://user:pass@host:port')")
	
	// Parse the flags
	flag.Parse()
	
	if log != nil {
		log.Debug("Flags parsed: search=%s, oa=%s, t=%s, pymin=%d, pymax=%d, pr=%s, lang=%s, output=%s, format=%s, max-pages=%d, no-headers=%v",
			*searchTerm, *accessType, *publicationType, *yearMin, *yearMax, *peerReviewed, *languages,
			*outputFile, *exportFormat, *maxPages, *noHeaders)
	}
	
	// Populate the SearchParams
	params.SearchTerm = *searchTerm
	params.AccessType = strings.ToLower(*accessType)
	params.PublicationType = *publicationType
	params.YearMin = *yearMin
	params.YearMax = *yearMax
	params.PeerReviewed = strings.ToLower(*peerReviewed)
	
	// Special handling for languages
	if *languages != "" {
		rawLanguages := strings.Split(*languages, "/")
		params.Languages = make([]string, len(rawLanguages))
		for i, lang := range rawLanguages {
			params.Languages[i] = strings.TrimSpace(lang)
		}
	}
	
	// Populate export parameters
	params.OutputFile = *outputFile
	params.ExportFormat = *exportFormat
	params.MaxPages = *maxPages
	params.IncludeHeaders = !*noHeaders
	
	// Set ExportResults based on whether OutputFile is provided
	params.ExportResults = params.OutputFile != ""
	
	// Set browser options
	params.RodOptions = *rodOptions
	params.StealthMode = *stealthMode
	params.RandomUserAgent = *randomUserAgent
	params.SlowMotion = *slowMotion
	params.PageDelay = *pageDelay
	params.Proxy = *proxy
	
	return params
}