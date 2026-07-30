# USTS — Gestão de UST e Ordens de Serviço

Aplicação web para gestão de **UST** (Unidade de Serviço Técnico) e **OS** (Ordem de Serviço).
O fluxo é hierárquico: projetos → sprints → tarefas → itens de OS, sendo os itens precificados
a partir de um catálogo de serviços em UST.

Este arquivo reúne as informações necessárias sobre:

- [Visão geral](#visão-geral)
- [Pré-requisitos](#pré-requisitos)
- [Estrutura de pastas](#estrutura-de-pastas)
- [Comandos do Make](#comandos-do-make)
- [Executando o projeto](#executando-o-projeto)
- [Desenvolvendo o projeto](#desenvolvendo-o-projeto)
- [Configuração](#configuração)
- [Banco de dados](#banco-de-dados)
- [Deploy](#deploy)

## Visão geral

Backend:

- Módulo Go: `github.com/gvillela7/usts`
- Go: `1.24`
- Router: `Chi`
- Banco de dados: `SQLite` (via `go-sqlite3`, exige CGO)
- Configuração: `Viper` (arquivo TOML)
- Porta do servidor: `8000`

Frontend:

- Templates: `Templ` (arquivos `.templ` compilados para `*_templ.go`)
- Interatividade: `htmx`
- CSS: `Tailwind CSS` com componentes `daisyUI`
- Bundler: `Parcel`
- Exportação para Excel: `ExcelJS`

Ferramentas de desenvolvimento:

- `Air` para live-reload
- `templ` para geração de HTML
- `golangci-lint` para lint
- `Prettier` para formatação dos arquivos em `web/assets`

A interface é em português; código e comentários são em inglês.

## Pré-requisitos

> ❗️ Garanta que os executáveis abaixo estejam instalados antes de iniciar o projeto:

- **Go 1.24+**: [https://go.dev/dl](https://go.dev/dl)
- **Node.js e npm** (para o build do frontend)
- **Air**: [https://github.com/air-verse/air](https://github.com/air-verse/air) — pode ser instalado com `make get-install-air`
- **Templ**: [https://github.com/a-h/templ](https://github.com/a-h/templ)
- **golangci-lint**: [https://github.com/golangci/golangci-lint](https://github.com/golangci/golangci-lint)

## Estrutura de pastas

```console
.
├── cmd
│   └── main.go                 # ponto de entrada: config → migrate → seed → servidor
├── configs
│   ├── config.go               # leitura do config.toml via Viper
│   └── database.go             # abertura da conexão, Migrate e Seed
├── database
│   ├── catalog.db              # banco SQLite
│   ├── schema.sql              # DDL (catalog, projects, sprints, tasks, os)
│   └── seed.sql                # carga inicial do catálogo
├── internal
│   ├── handler                 # handlers de páginas e endpoints htmx
│   ├── repository              # interface Repository, SQLiteRepository e structs de domínio
│   └── routes                  # ponto único de wiring e registro das rotas
├── web
│   ├── assets                  # fontes do frontend (JS e CSS) processados pelo Parcel
│   │   ├── alerts.js
│   │   ├── dropdown.js
│   │   ├── excel-export.js
│   │   ├── modal-handler.js
│   │   ├── scripts.js
│   │   ├── styles.css
│   │   └── table.js
│   ├── static                  # assets gerados/minificados e arquivos estáticos (embed)
│   ├── templates
│   │   ├── components
│   │   ├── pages
│   │   └── main.templ          # Layout compartilhado por todas as páginas
│   ├── efs.go                  # //go:embed all:static
│   ├── package.json
│   └── prettier.config.js
├── .air.toml
├── .golangci.yml
├── config.toml
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

## Comandos do Make

O `Makefile` é o ponto central de automação do projeto. Rode `make help` para listar os alvos disponíveis.

| Alvo | Descrição |
| --- | --- |
| `make help` | Lista todos os alvos disponíveis. |
| `make get-install-air` | Instala o Air em `$(go env GOPATH)/bin` via script oficial. |
| `make npm-install` | Instala as dependências do frontend (`cd web && npm install`). |
| `make npm-build` | Gera os assets de produção com Parcel em `web/static`. |
| `make npm-watch` | Observa os assets do frontend e reconstrói sem otimização. |
| `make templ-generate` | Gera os arquivos `*_templ.go` a partir dos `.templ`. |
| `make templ-watch` | Gera os `*_templ.go` em modo watch. |
| `make build` | Build completo: assets do frontend → `templ generate` → `go build`. |
| `make watch` | Compila o binário e sobe o Air para desenvolvimento com live-reload. |

## Executando o projeto

Na primeira execução, instale as dependências do frontend:

```console
make npm-install
```

Depois, suba o ambiente de desenvolvimento com live-reload:

```console
make watch
```

A aplicação fica disponível em [http://localhost:8000](http://localhost:8000).

Para gerar um build completo sem live-reload:

```console
make build
```

> ❗️ A ordem do build importa. `web/static` é embutido no binário via `//go:embed all:static`
> em [web/efs.go](web/efs.go); se o Parcel nunca rodou, o `go build` falha. Execute
> `make npm-build` antes de qualquer `go build` direto. Da mesma forma, alterações em arquivos
> `.templ` só têm efeito depois de `make templ-generate` — o Air já executa as duas etapas
> automaticamente (`pre_cmd` e `cmd` em [.air.toml](.air.toml)).

> ❗️ A aplicação precisa ser executada a partir da raiz do repositório, pois lê `config.toml`,
> `database/schema.sql` e `database/seed.sql` por caminho relativo.

> ℹ️ A variável `APP_NAME` não está definida no `Makefile`, então `make build` e `make watch`
> geram `tmp/main`, enquanto o Air compila e executa `tmp/gowebly_chi`. Durante o
> desenvolvimento, o Air é a referência.

## Desenvolvendo o projeto

O backend está organizado em três camadas, com um único ponto de wiring em
[internal/routes/routes.go](internal/routes/routes.go), que constrói o `SQLiteRepository`,
injeta no `Handler` e registra todas as rotas:

- **repository** — interface `Repository` e implementação `SQLiteRepository` sobre `database/sql`
  puro (sem ORM), além das structs de domínio (`Project`, `Sprint`, `Task`, `Catalog`, `OS`).
- **handler** — struct `Handler` com os renderizadores de página e os endpoints htmx.
- **templates** — `Layout(title, currentPage, metaTags, bodyContent)` em `main.templ` envolve
  todas as páginas; `currentPage` controla o destaque do link ativo na navbar.

Para adicionar um novo método de repositório, declare-o primeiro na interface `Repository`
(o handler depende da interface, não do tipo concreto) e depois implemente em `SQLiteRepository`.

Sobre o frontend:

- Todo o JavaScript usa IIFE e globais em `window.*`, importados por
  [web/assets/scripts.js](web/assets/scripts.js) — não há sistema de módulos.
- O `main.templ` define um `Content-Security-Policy` restrito (apenas self, Google Fonts e
  `ws://localhost:*`). Qualquer asset externo novo precisa ser liberado ali.

Lint e formatação:

```console
golangci-lint run        # config em .golangci.yml (ignora *_templ.go)
cd web && npm run fmt    # Prettier para web/assets
```

> ℹ️ O projeto não possui testes automatizados (nenhum arquivo `_test.go` e nenhum alvo de teste).

## Configuração

As configurações ficam em [config.toml](config.toml), lido pelo Viper na inicialização:

```toml
[database]
  driver = "sqlite3"
  dsn = "./database/catalog.db"

[server]
  port = 8000
```

## Banco de dados

Na inicialização, a aplicação executa `Migrate` e depois `Seed`:

- `Migrate` executa [database/schema.sql](database/schema.sql) por inteiro a cada boot. Como
  são apenas comandos `CREATE TABLE IF NOT EXISTS`, somente alterações aditivas se aplicam
  de forma limpa. Não há ferramenta de migração: alterar uma tabela existente exige
  intervenção manual em `database/catalog.db`.
- `Seed` carrega [database/seed.sql](database/seed.sql) apenas quando a tabela `catalog`
  está vazia.

O esquema tem cinco tabelas — `catalog`, `projects`, `sprints`, `tasks` e `os` — na hierarquia
projeto → sprint → tarefa → OS, com `os` referenciando também `catalog`. Todas as tabelas,
exceto `catalog`, possuem `deleted_at` para soft delete, mas nenhuma consulta filtra por essa
coluna ainda — o soft delete existe apenas no esquema.

## Deploy

O deploy usa Docker, com as configurações em [Dockerfile](Dockerfile) e
[docker-compose.yml](docker-compose.yml):

```console
docker compose up -d --build
```

A aplicação fica disponível em [http://localhost:8000](http://localhost:8000).

O `Dockerfile` é multi-stage e faz todo o build dentro do container — não é necessário rodar
`make` antes:

1. **frontend** (`node:22-alpine`) — `npm ci` e `npm run build`, gerando os assets em
   `web/static`. Os templates são copiados porque o Tailwind escaneia as classes usadas neles.
2. **builder** (`golang:1.24-alpine`) — instala o `templ`, executa `templ generate` e compila
   `./cmd/main.go`. O toolchain C (`gcc`, `musl-dev`) e `CGO_ENABLED=1` são obrigatórios:
   `github.com/mattn/go-sqlite3` é um pacote cgo.
3. **runtime** (`alpine:3.21`) — recebe o binário, o `config.toml` e os arquivos SQL. A imagem
   é Alpine, e não `scratch`, porque o binário com cgo depende da musl.

O banco SQLite é persistido no volume nomeado `usts_data`, montado em `/app/database`.
O `schema.sql` e o `seed.sql` são montados por cima como bind mounts read-only, de modo que
alterações no repositório sempre chegam ao container sem precisar recriar o volume.

Para apagar os dados e começar de um banco vazio:

```console
docker compose down -v
```

Para deploy em um servidor remoto:

1. Provisione um VPS e instale Docker, Docker Compose e Git.
2. Clone o repositório e entre na pasta do projeto.
3. Execute `docker compose up -d --build`.