package config

import (
	"flag"
	"strings"

	"github.com/alexandreffaria/reviu/internal/logger"
)

// FlagNames contains the names of all command-line flags
type FlagNames struct {
	SearchTerm      string
	AccessType      string
	PublicationType string
	YearMin         string
	YearMax         string
	PeerReviewed    string
	Languages       string
}

// DefaultFlagNames provides the standard flag names
var DefaultFlagNames = FlagNames{
	SearchTerm:      "search",
	AccessType:      "oa",
	PublicationType: "t",
	YearMin:         "pymin",
	YearMax:         "pymax",
	PeerReviewed:    "pr",
	Languages:       "lang",
}

// SetupFlags configures and parses command-line flags
// Returns a populated SearchParams struct
func SetupFlags(log logger.Logger) *SearchParams {
	params := NewSearchParams()
	
	// Define all flag pointers
	searchTerm := flag.String(DefaultFlagNames.SearchTerm, "", 
                             "Termo para pesquisar")
	accessType := flag.String(DefaultFlagNames.AccessType, "", 
                             "Acesso aberto: 'sim', 'nao' ou omitir para qualquer")
	publicationType := flag.String(DefaultFlagNames.PublicationType, "", 
                                  "Tipo de publicação (ex: 'Artigo')")
	yearMin := flag.Int(DefaultFlagNames.YearMin, 0, 
                       "Ano mínimo de publicação")
	yearMax := flag.Int(DefaultFlagNames.YearMax, 0, 
                       "Ano máximo de publicação")
	peerReviewed := flag.String(DefaultFlagNames.PeerReviewed, "", 
                               "Revisão por pares: 'sim', 'nao' ou omitir para qualquer")
	languages := flag.String(DefaultFlagNames.Languages, "", 
                            "Idiomas separados por '/' (ex: 'Português/Inglês/Espanhol')")
	
	// Parse the flags
	flag.Parse()
	
	if log != nil {
		log.Debug("Flags parsed: search=%s, oa=%s, t=%s, pymin=%d, pymax=%d, pr=%s, lang=%s",
			*searchTerm, *accessType, *publicationType, *yearMin, *yearMax, *peerReviewed, *languages)
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
	
	return params
}