# CAPES Periódicos - Ferramenta de Busca

Ferramenta de linha de comando para realizar buscas no Portal de Periódicos da CAPES, construída em Go.

## Requisitos

- Go 1.15 ou superior
- Pacote "github.com/go-rod/rod" para automação de navegador

## Instalação

Clone este repositório e instale as dependências:

```bash
git clone https://github.com/seu-usuario/capes-periodicos.git
cd capes-periodicos
go get github.com/go-rod/rod
```

## Como Usar

Execute o programa usando `go run` ou compile-o primeiro:

```bash
# Usando go run
go run main.go [flags]

# Ou compile e execute
go build -o capes-search
./capes-search [flags]
```

### Flags Disponíveis

| Flag | Descrição | Exemplo | Observação |
|------|-----------|---------|------------|
| `-search` | Termo de busca | `-search "inteligência artificial"` | Obrigatório |
| `-oa` | Filtro de acesso aberto | `-oa sim` ou `-oa nao` | Opcional |
| `-t` | Tipo de publicação | `-t "Artigo"` | Opcional |
| `-pymin` | Ano mínimo de publicação | `-pymin 2010` | Opcional |
| `-pymax` | Ano máximo de publicação | `-pymax 2023` | Opcional |

### Exemplos

**1. Busca básica por um termo:**
```bash
go run main.go -search "violência contra mulheres"
```

**2. Busca com filtro de acesso aberto:**
```bash
go run main.go -search "inteligência artificial" -oa sim
```

**3. Busca por tipo de publicação:**
```bash
go run main.go -search "vacinas covid" -t "Artigo"
```

**4. Busca por período específico:**
```bash
go run main.go -search "mudanças climáticas" -pymin 2015 -pymax 2023
```

**5. Combinando múltiplos filtros:**
```bash
go run main.go -search "pandemia" -oa sim -t "Artigo" -pymin 2020 -pymax 2023
```

## Funcionamento

A ferramenta constrói uma URL de busca com os parâmetros especificados e abre um navegador automatizado que carrega a página de resultados. O navegador permanece aberto por 30 segundos para permitir a visualização dos resultados.

## Observações

- Se nenhum valor for fornecido para os filtros opcionais, eles não serão incluídos na URL de busca.
- O navegador é aberto em modo visível (não headless) para facilitar a visualização dos resultados.
- A busca é realizada diretamente via URL, sem interação com elementos da página.