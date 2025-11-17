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
	tipoPublicacao := flag.String("t", "", "Tipo de publicação (ex: 'Artigo')")
	anoMinimo := flag.Int("pymin", 0, "Ano mínimo de publicação")
	anoMaximo := flag.Int("pymax", 0, "Ano máximo de publicação")
	revisaoPares := flag.String("pr", "", "Revisão por pares: 'sim', 'nao' ou omitir para qualquer")
	linguagens := flag.String("lang", "", "Idiomas separados por '/' (ex: 'Português/Inglês/Espanhol')")
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
	
	// Validar e normalizar valor de revisão por pares (se fornecido)
	revisao := strings.ToLower(*revisaoPares)
	if revisao != "" && revisao != "sim" && revisao != "nao" {
		fmt.Println("Valor inválido para -pr. O valor será ignorado.")
		revisao = ""
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
	
	if *tipoPublicacao != "" {
		fmt.Printf("Tipo de publicação: %s\n", *tipoPublicacao)
	} else {
		fmt.Printf("Tipo de publicação: qualquer\n")
	}
	
	// Obter o ano atual
	anoAtual := time.Now().Year()
	
	// Ajustar o ano máximo para o ano atual se apenas o mínimo foi especificado
	anoMaximoEfetivo := *anoMaximo
	if *anoMinimo > 0 && *anoMaximo == 0 {
		anoMaximoEfetivo = anoAtual
	}
	
	// Mostrar anos se pelo menos um deles foi especificado
	if *anoMinimo > 0 || anoMaximoEfetivo > 0 {
		anoMinStr := "não especificado"
		anoMaxStr := "não especificado"
		
		if *anoMinimo > 0 {
			anoMinStr = fmt.Sprintf("%d", *anoMinimo)
		}
		
		if anoMaximoEfetivo > 0 {
			anoMaxStr = fmt.Sprintf("%d", anoMaximoEfetivo)
		}
		
		fmt.Printf("Anos de publicação: %s até %s\n", anoMinStr, anoMaxStr)
	} else {
		fmt.Printf("Anos de publicação: qualquer\n")
	}
	
	if revisao != "" {
		fmt.Printf("Revisão por pares:  %s\n", revisao)
	} else {
		fmt.Printf("Revisão por pares:  qualquer\n")
	}
	
	// Processar linguagens
	var listaLinguagens []string
	if *linguagens != "" {
		listaLinguagens = strings.Split(*linguagens, "/")
		for i, lang := range listaLinguagens {
			listaLinguagens[i] = strings.TrimSpace(lang)
		}
		fmt.Printf("Idiomas:            %s\n", strings.Join(listaLinguagens, ", "))
	} else {
		fmt.Printf("Idiomas:            qualquer\n")
	}
	fmt.Println("========================================\n")

	// URL base da página de busca
	baseURL := "https://www.periodicos.capes.gov.br/index.php/acervo/buscador.html"

	// Construir os parâmetros de query manualmente para controlar a ordem exata
	var urlParams []string
	
	// Adicionar termo de busca (primeiro parâmetro)
	termoBusca := url.QueryEscape(termo)
	// Substituir %20 por + para match exato com a URL de exemplo
	termoBusca = strings.ReplaceAll(termoBusca, "%20", "+")
	urlParams = append(urlParams, "q="+termoBusca)
	
	// Adicionar source vazio (segundo parâmetro)
	urlParams = append(urlParams, "source=")
	
	// Adicionar parâmetro de acesso aberto apenas se o flag foi especificado
	if acesso == "sim" {
		urlParams = append(urlParams, "open_access%5B%5D=open_access%3D%3D1")
	} else if acesso == "nao" {
		urlParams = append(urlParams, "open_access%5B%5D=open_access%3D%3D0")
	}
	
	// Adicionar tipo de publicação apenas se o flag foi especificado
	if *tipoPublicacao != "" {
		tipoEncoded := url.QueryEscape("type=="+*tipoPublicacao)
		urlParams = append(urlParams, "type%5B%5D="+tipoEncoded)
	}
	
	// Adicionar anos de publicação apenas se especificados
	if *anoMinimo > 0 {
		urlParams = append(urlParams, fmt.Sprintf("publishyear_min%%5B%%5D=%d", *anoMinimo))
	}
	
	// Usar o ano máximo especificado ou o ano atual se apenas o mínimo foi fornecido
	if anoMaximoEfetivo > 0 {
		urlParams = append(urlParams, fmt.Sprintf("publishyear_max%%5B%%5D=%d", anoMaximoEfetivo))
	}
	
	// Adicionar parâmetro de revisão por pares apenas se o flag foi especificado
	if revisao == "sim" {
		urlParams = append(urlParams, "peer_reviewed%5B%5D=peer_reviewed%3D%3D1")
	} else if revisao == "nao" {
		urlParams = append(urlParams, "peer_reviewed%5B%5D=peer_reviewed%3D%3D0")
	}
	
	// Adicionar parâmetros de idioma
	for _, lang := range listaLinguagens {
		// Escape especial para o ê em "Português"
		langEncoded := strings.ReplaceAll(lang, "ê", "%C3%AA")
		urlParams = append(urlParams, fmt.Sprintf("language%%5B%%5D=language%%3D%%3D%s", langEncoded))
	}
	
	// Construir a URL completa com parâmetros na ordem específica
	searchURL := baseURL + "?" + strings.Join(urlParams, "&")
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
