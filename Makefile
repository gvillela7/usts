.PHONY: help
help: ## print make targets 
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: get-install-air
get-install-air: ## Installs the air build reload system using cUrl
	curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

.PHONY: npm-install
npm-install: ## install frontend dependencies
	cd web && npm install

.PHONY: npm-build
npm-build: ## build frontend assets
	cd web && npm run build

.PHONY: npm-watch
npm-watch: ## watch frontend assets
	cd web && npm run watch

.PHONY: build
build: ## build the project
	cd web && npm run build
	templ generate
	go build -o ./tmp/$(APP_NAME) ./cmd/main.go

.PHONY: watch
watch: ## build and watch the project with air
	go build -o ./tmp/$(APP_NAME) ./cmd/main.go && air

.PHONY: templ-generate
templ-generate:
	templ generate

.PHONY: templ-watch
templ-watch:
	templ generate --watch