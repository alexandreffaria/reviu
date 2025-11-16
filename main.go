package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
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
	// Definir flag para o termo de busca
	searchTerm := flag.String("termo", "", "Termo para pesquisar")
	flag.Parse()

	// Se o termo de busca não foi fornecido como flag, solicitar ao usuário
	termo := *searchTerm
	if termo == "" {
		termo = promptTextRequired("TERMOS DE BUSCA", "texto livre (obrigatório)")
	}

	// Exibir termo de busca
	fmt.Println("\n========================================")
	fmt.Println(" RELATÓRIO DA BUSCA")
	fmt.Println("========================================")
	fmt.Printf("Termos de busca: %s\n", termo)
	fmt.Println("========================================\n")

	// URL base da página de busca
	baseURL := "https://www.periodicos.capes.gov.br/index.php/acervo/buscador.html"

	// Iniciar o navegador
	u := launcher.New().Headless(false).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()
	page := browser.MustPage(baseURL).MustWaitLoad()

	fmt.Println("Realizando busca pelo termo:", termo)

	// Localizar o campo de busca
	inputField := page.MustElement("input[name='q']")
	
	// Focar no campo e preencher
	inputField.MustFocus()
	inputField.MustInput(termo)
	
	// Pressionar Enter usando o teclado da página
	page.Keyboard.Press(input.Enter)
	
	// Aguardar o carregamento da página
	page.MustWaitLoad()
	
	fmt.Println("Busca realizada com sucesso. Mantendo navegador aberto por 30 segundos.")
	
	// Manter o navegador aberto por 30 segundos
	time.Sleep(30 * time.Second)
}
