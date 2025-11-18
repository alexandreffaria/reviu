# CAPES Periódicos - Ferramenta de Busca e Exportação

Ferramenta de linha de comando para realizar buscas no Portal de Periódicos da CAPES, extrair resultados de múltiplas páginas e exportar para CSV, construída em Go.

## Características

- Busca avançada com múltiplos filtros (acesso aberto, tipo de publicação, ano, idioma, etc.)
- Extração de resultados de múltiplas páginas de busca
- Exportação automática para arquivo CSV
- Medidas anti-bloqueio para evitar detecção durante a coleta de dados
- Arquitetura modular e manutenível
- Interface de linha de comando amigável

## Requisitos

- Go 1.18 ou superior
- Navegador Chrome/Chromium instalado (usado pelo rod em segundo plano)

## Instalação

Clone este repositório e instale as dependências:

```bash
git clone https://github.com/seu-usuario/capes-periodicos.git
cd capes-periodicos
go mod download
```

## Compilando o Binário

Para compilar um binário executável para distribuição:

```bash
# Compilar para o sistema atual
go build -o capes-search cmd/capes-search/main.go

# Cross-compilação para Windows 64-bit
GOOS=windows GOARCH=amd64 go build -o capes-search.exe cmd/capes-search/main.go

# Cross-compilação para Linux 64-bit
GOOS=linux GOARCH=amd64 go build -o capes-search cmd/capes-search/main.go

# Cross-compilação para macOS
GOOS=darwin GOARCH=amd64 go build -o capes-search cmd/capes-search/main.go
```

## Como Usar

Execute o programa usando `go run` ou utilize o binário compilado:

```bash
# Usando go run
go run cmd/capes-search/main.go [flags]

# Ou utilize o binário compilado
./capes-search [flags]
```

### Flags de Busca

| Flag | Descrição | Exemplo | Observação |
|------|-----------|---------|------------|
| `-search` | Termo de busca | `-search "inteligência artificial"` | Obrigatório |
| `-oa` | Filtro de acesso aberto | `-oa sim` ou `-oa nao` | Opcional |
| `-t` | Tipo de publicação | `-t "Artigo"` | Opcional |
| `-pymin` | Ano mínimo de publicação | `-pymin 2010` | Opcional |
| `-pymax` | Ano máximo de publicação | `-pymax 2023` | Opcional, se omitido com `-pymin` definido, usa o ano atual |
| `-pr` | Revisão por pares | `-pr sim` ou `-pr nao` | Opcional |
| `-lang` | Filtro de idiomas | `-lang "Português/Inglês/Espanhol"` | Opcional, múltiplos idiomas separados por `/` |

### Flags de Exportação

| Flag | Descrição | Exemplo | Observação |
|------|-----------|---------|------------|
| `-output` | Arquivo de saída | `-output "resultados.csv"` | Habilita a exportação de resultados |
| `-format` | Formato de exportação | `-format csv` | Atualmente apenas CSV é suportado |
| `-max-pages` | Máximo de páginas | `-max-pages 5` | Limita o número de páginas processadas (0 = todas) |
| `-no-headers` | Sem cabeçalhos | `-no-headers` | Remove a linha de cabeçalho do CSV |

### Flags Anti-Bloqueio

| Flag | Descrição | Exemplo | Observação |
|------|-----------|---------|------------|
| `-delay` | Delay entre páginas | `-delay 5s` | Espera entre páginas para evitar bloqueio |
| `-stealth` | Modo stealth | `-stealth=false` | Desativa o modo stealth (ativado por padrão) |
| `-random-ua` | Agente aleatório | `-random-ua=false` | Desativa o agente de usuário aleatório (ativado por padrão) |
| `-proxy` | Proxy | `-proxy "http://user:pass@host:port"` | Define um proxy para requisições |

### Exemplos de Uso

**1. Busca básica por um termo:**
```bash
./capes-search -search "violência contra mulheres"
```

**2. Exportar resultados para CSV:**
```bash
./capes-search -search "inteligência artificial" -output "resultados.csv"
```

**3. Limitar número de páginas e configurar delay:**
```bash
./capes-search -search "machine learning" -max-pages 5 -delay 3s -output "ml_results.csv"
```

**4. Busca com filtro de acesso aberto:**
```bash
./capes-search -search "inteligência artificial" -oa sim -output "ia_open_access.csv"
```

**5. Busca por tipo de publicação com anti-bloqueio avançado:**
```bash
./capes-search -search "vacinas covid" -t "Artigo" -delay 5s -proxy "http://myproxy:8080" -output "vacinas.csv"
```

**6. Busca por período específico com múltiplas páginas:**
```bash
./capes-search -search "mudanças climáticas" -pymin 2015 -pymax 2023 -max-pages 10 -output "clima.csv"
```

**7. Busca por idioma específico:**
```bash
./capes-search -search "educação" -lang "Português" -output "educacao_pt.csv"
```

**8. Exportação sem cabeçalhos:**
```bash
./capes-search -search "economia" -lang "Português/Inglês" -no-headers -output "economia.csv"
```

**9. Combinando múltiplos filtros com anti-bloqueio:**
```bash
./capes-search -search "pandemia" -oa sim -t "Artigo" -pr sim -pymin 2020 -lang "Português/Inglês" -delay 5s -max-pages 20 -output "pandemia.csv"
```

## Funcionamento

A ferramenta opera nos seguintes passos:

1. Constrói uma URL de busca com os parâmetros especificados
2. Abre um navegador automatizado com medidas anti-bloqueio
3. Navega para a URL de busca inicial
4. Extrai os resultados da primeira página
5. Se a exportação estiver habilitada e houver mais páginas a processar:
   - Espera o tempo definido por `-delay` entre as páginas
   - Navega para a próxima página
   - Extrai os resultados
   - Repete até atingir o limite de páginas ou o final dos resultados
6. Exporta todos os resultados para o arquivo CSV especificado
7. Fecha o navegador automaticamente

## Observações

- As medidas anti-bloqueio (stealth mode, agentes aleatórios, delay entre páginas) são essenciais para evitar detecção durante a extração de múltiplas páginas.
- Recomenda-se utilizar um valor adequado para `-delay` (por exemplo, 3-5 segundos) para reduzir o risco de bloqueio.
- Para coletas extensas, considere limitar o número de páginas com `-max-pages`.
- O arquivo CSV resultante pode ser aberto em Excel, LibreOffice Calc, Google Sheets, etc.

## Estrutura do Projeto

O projeto foi refatorado para uma arquitetura modular:

- `cmd/capes-search`: Ponto de entrada do programa
- `internal/browser`: Gerenciamento de navegador e interação com páginas
- `internal/cli`: Interface de linha de comando
- `internal/config`: Configuração e processamento de flags
- `internal/errors`: Tratamento estruturado de erros
- `internal/logger`: Sistema de logging
- `internal/result`: Extração e exportação de resultados
- `internal/search`: Construção de URLs de busca