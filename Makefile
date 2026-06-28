# =============================================================================
#  Makefile — comandos de desarrollo, build y deploy del dashboard
# =============================================================================

BINARY      := polaris-dashboard
IMAGE       := polaris-dashboard:latest
FRONTEND    := frontend
DIST        := internal/server/dist

.PHONY: help dev dev-backend dev-frontend build build-frontend build-backend \
        deps run docker docker-run clean test fmt

help: ## Muestra esta ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

## --- Desarrollo ------------------------------------------------------------

deps: ## Instala dependencias de Go y del frontend
	go mod download
	cd $(FRONTEND) && npm install

dev: ## Levanta backend (:3000) y frontend Vite (:5173) en paralelo
	@$(MAKE) -j2 dev-backend dev-frontend

dev-backend: ## Solo el backend de Go con recarga manual
	CONFIG_PATH=config.yaml go run ./cmd/server

dev-frontend: ## Solo el dev server de Vite (proxy /api -> :3000)
	cd $(FRONTEND) && npm run dev

## --- Build -----------------------------------------------------------------

build: build-frontend build-backend ## Build completo (frontend embebido en el binario)

build-frontend: ## Compila el frontend hacia $(DIST)
	cd $(FRONTEND) && npm run build

build-backend: ## Compila el binario de Go (requiere el frontend ya compilado)
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY) ./cmd/server

run: ## Ejecuta el binario compilado
	./$(BINARY)

## --- Docker ----------------------------------------------------------------

docker: ## Construye la imagen Docker de producción
	docker build -t $(IMAGE) .

docker-run: ## Corre la imagen montando config.yaml
	docker run --rm -p 3000:3000 -v $$(pwd)/config.yaml:/app/config.yaml $(IMAGE)

## --- Utilidades ------------------------------------------------------------

test: ## Corre los tests de Go
	go test ./...

fmt: ## Formatea el código Go
	go fmt ./...

clean: ## Elimina binario y build del frontend
	rm -f $(BINARY)
	rm -rf $(DIST)/static $(DIST)/index.html $(FRONTEND)/node_modules/.vite
