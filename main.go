package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"net/url"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

type Form struct {
	Escopo           string
	Campo            string
	Match            string
	Termos           string
	TiposMaterial    []string
	AcessoAberto     string
	TipoRecurso      []string
	TipoCaseDados    string
	AnoInicio        string
	AnoFim           string
	ProducaoNacional string
	RevisadoPares    string
	Areas            []string
	Idiomas          []string
}

var in = bufio.NewReader(os.Stdin)

func promptSingle(label string, opts []string, defIdx int) string {
	fmt.Println("\n" + label)
	for i, o := range opts {
		def := ""
		if i == defIdx {
			def = " *"
		}
		fmt.Printf("%d) %s%s\n", i+1, o, def)
	}
	fmt.Printf("Escolha (1-%d) [Enter=%d]: ", len(opts), defIdx+1)
	s, _ := in.ReadString('\n')
	s = strings.TrimSpace(s)
	if s == "" {
		return opts[defIdx]
	}
	i, err := strconv.Atoi(s)
	if err != nil || i < 1 || i > len(opts) {
		return opts[defIdx]
	}
	return opts[i-1]
}

func promptMulti(label string, opts []string) []string {
	fmt.Println("\n" + label)
	for i, o := range opts {
		fmt.Printf("%d) %s\n", i+1, o)
	}
	fmt.Print("Escolhas (ex: 1,3,7) [Enter=nenhum]: ")
	s, _ := in.ReadString('\n')
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	seen := map[int]bool{}
	var res []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		i, err := strconv.Atoi(part)
		if err != nil || i < 1 || i > len(opts) || seen[i] {
			continue
		}
		seen[i] = true
		res = append(res, opts[i-1])
	}
	return res
}

func promptText(label, hint string) string {
	if hint != "" {
		fmt.Printf("\n%s (%s): ", label, hint)
	} else {
		fmt.Printf("\n%s: ", label)
	}
	s, _ := in.ReadString('\n')
	return strings.TrimSpace(s)
}

func promptTextRequired(label, hint string) string {
	for {
		v := promptText(label, hint)
		if v != "" {
			return v
		}
		fmt.Println("Campo obrigatório. Por favor, preencha.")
	}
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

func joinOrDash(ss []string) string {
	if len(ss) == 0 {
		return "—"
	}
	return strings.Join(ss, ", ")
}

func printReport(f Form) {
	// defaults de ano se vazio
	if f.AnoInicio == "" {
		f.AnoInicio = "2020"
	}
	if f.AnoFim == "" {
		f.AnoFim = "2025"
	}

	fmt.Println("\n========================================")
	fmt.Println(" RELATÓRIO DA BUSCA")
	fmt.Println("========================================")
	fmt.Printf("Escopo:              %s\n", orDash(f.Escopo))
	fmt.Printf("Campo:               %s\n", orDash(f.Campo))
	fmt.Printf("Correspondência:     %s\n", orDash(f.Match))
	fmt.Printf("Termos de busca:     %s\n", orDash(f.Termos))
	fmt.Printf("Tipos de material:   %s\n", joinOrDash(f.TiposMaterial))
	fmt.Printf("Acesso aberto:       %s\n", orDash(f.AcessoAberto))
	fmt.Printf("Tipo do recurso:     %s\n", joinOrDash(f.TipoRecurso))
	fmt.Printf("Case de dados:       %s\n", orDash(f.TipoCaseDados))
	fmt.Printf("Ano (início–fim):    %s–%s\n", orDash(f.AnoInicio), orDash(f.AnoFim))
	fmt.Printf("Produção Nacional:   %s\n", orDash(f.ProducaoNacional))
	fmt.Printf("Revisado por pares:  %s\n", orDash(f.RevisadoPares))
	fmt.Printf("Áreas:               %s\n", joinOrDash(f.Areas))
	fmt.Printf("Idiomas:             %s\n", joinOrDash(f.Idiomas))
	fmt.Println("========================================\n")
}

func main() {
	f := Form{}

	// 1) Escopo da busca (single, default = Buscar tudo)
	f.Escopo = promptSingle("Escopo da busca",
		[]string{"Buscar tudo", "bases", "periódicos", "livros"}, 0)

	// 2) Campo (single, default = qualquer campo)
	f.Campo = promptSingle("Campo",
		[]string{"qualquer campo", "título", "autor", "assunto", "editor"}, 0)

	// 3) Tipo de correspondência (single, default = contém)
	f.Match = promptSingle("Tipo de correspondência",
		[]string{"contém", "é (exato)"}, 0)

	// 4) TERMOS DE BUSCA (obrigatório)
	f.Termos = promptTextRequired("TERMOS DE BUSCA", "texto livre (obrigatório)")

	// Tipos de material (multi)
	f.TiposMaterial = promptMulti("Tipos de material (multi)",
		[]string{"artigo", "capítulo de livro", "paratexto", "revisão", "pré-print", "livro", "dissertação", "conjunto de dados"})

	// Acesso aberto (single)
	f.AcessoAberto = promptSingle("Acesso aberto", []string{"Sim", "Não"}, 0)

	// Tipo do recurso (multi)
	f.TipoRecurso = promptMulti("Tipo do recurso (multi)",
		[]string{"Artigo", "Revisão", "Capítulo de Livro", "Jornais", "Carta", "Editorial"})

	// Tipo de case de dados (single)
	f.TipoCaseDados = promptSingle("Tipo de case de dados",
		[]string{"Livro", "Periódico", "Documento"}, 0)

	// Ano de criação (opcionais; defaultizados no relatório)
	f.AnoInicio = promptText("Ano de criação - início", "aaaa (Enter = 2020)")
	f.AnoFim = promptText("Ano de criação - fim", "aaaa (Enter = 2025)")

	// Produção Nacional (single)
	f.ProducaoNacional = promptSingle("Produção Nacional", []string{"sim", "não"}, 0)

	// Revisado por pares (single)
	f.RevisadoPares = promptSingle("Revisado por pares", []string{"sim", "não"}, 0)

	// Áreas (multi)
	f.Areas = promptMulti("Áreas (multi)",
		[]string{"ciências humanas", "ciências da saúde", "ciências sociais aplicadas", "multidisciplinar", "linguística, letras e artes", "ciências exatas e da terra", "engenharias", "ciências biológicas", "ciências agrárias"})

	// Idioma (multi)
	f.Idiomas = promptMulti("Idioma (multi)",
		[]string{"inglês", "português", "espanhol", "francês", "alemão", "italiano"})

	// Imprime o relatório “bonito”
	printReport(f)

	q := url.QueryEscape(fmt.Sprintf("all:contains(%s)", f.Termos))
	finalURL := "https://www.periodicos.capes.gov.br/index.php/acervo/buscador.html?mode=advanced&source=all&q=" + q

	// (Por enquanto) mantém o binário funcional abrindo o navegador:
	u := launcher.New().Headless(false).MustLaunch()
	rod.New().ControlURL(u).MustConnect().
		MustPage(finalURL).
		MustWaitLoad()
	time.Sleep(10 * time.Second)
}
