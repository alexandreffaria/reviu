package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

var in = bufio.NewReader(os.Stdin)

func promptTextRequired(label, hint string) string {
	for {
		if hint != "" {
			fmt.Printf("\n%s (%s): ", label, hint)
		} else {
			fmt.Printf("\n%s: ", label)
		}
		s, _ := in.ReadString('\n')
		s = strings.TrimSpace(s)
		if s != "" {
			return s
		}
		fmt.Println("Campo obrigatório. Por favor, preencha.")
	}
}

func main() {
	// Definir flags para busca
	searchTerm := flag.String("search", "Violência contra mulheres", "Termo para pesquisar")
	acessoAberto := flag.String("oa", "", "Acesso aberto: 'sim', 'nao' ou omitir para qualquer")
	flag.Parse()

	// Se o termo de busca não foi fornecido como flag, solicitar ao usuário
	termo := *searchTerm
	if termo == "" {
		termo = promptTextRequired("TERMOS DE BUSCA", "texto livre (obrigatório)")
	}

	// Validar e normalizar valor de acesso-aberto (se fornecido)
	acesso := strings.ToLower(*acessoAberto)
	if acesso != "" && acesso != "sim" && acesso != "nao" {
		fmt.Println("Valor inválido para -oa. O valor será ignorado.")
		acesso = ""
	}

	// Exibir relatório
	fmt.Println("\n========================================")
	fmt.Println(" RELATÓRIO DA BUSCA")
	fmt.Println("========================================")
	fmt.Printf("Termos de busca:   %s\n", termo)
	if acesso != "" {
		fmt.Printf("Acesso aberto:     %s\n", acesso)
	} else {
		fmt.Printf("Acesso aberto:     qualquer\n")
	}
	fmt.Println("========================================\n")

	// URL base da página de busca
	baseURL := "https://www.periodicos.capes.gov.br/index.php/acervo/buscador.html"

	// Construir os parâmetros de query
	params := url.Values{}
	
	// Adicionar termo de busca
	params.Add("q", termo)
	
	// Adicionar fonte expandida
	params.Add("source", "expanded")
	
	// Adicionar parâmetro de acesso aberto apenas se o flag foi especificado
	if acesso == "sim" {
		params.Add("open_access[]", "open_access==1")
	} else if acesso == "nao" {
		params.Add("open_access[]", "open_access==0")
	}
	// Se acesso estiver vazio, não adiciona nenhum parâmetro de acesso aberto
	
	// Construir a URL completa
	searchURL := baseURL + "?" + params.Encode()
	fmt.Println("URL da busca:", searchURL)

	// Iniciar o navegador
	u := launcher.New().Headless(false).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()
	
	// Abrir a página com a URL de busca
	fmt.Println("Abrindo navegador com a URL de busca...")
	_ = browser.MustPage(searchURL).MustWaitLoad()
	
	fmt.Println("Busca realizada com sucesso.")
	fmt.Println("Mantendo navegador aberto por 30 segundos para visualização dos resultados.")
	
	// Manter o navegador aberto por 30 segundos
	time.Sleep(30 * time.Second)
}
