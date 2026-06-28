# =============================================================================
#  Dockerfile multistage — imagen final mínima con el frontend embebido
# =============================================================================

# --- Stage 1: build del frontend ---------------------------------------------
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
# Vite escribe en ../internal/server/dist (ver vite.config.js)
COPY internal/server/dist/.gitkeep ../internal/server/dist/.gitkeep
RUN npm run build

# --- Stage 2: build del binario de Go ----------------------------------------
FROM golang:1.26-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Trae el dist generado por el stage anterior para que //go:embed lo incluya
COPY --from=frontend /app/internal/server/dist ./internal/server/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /polaris ./cmd/server

# --- Stage 3: imagen final ---------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=backend /polaris /app/polaris
# El config.yaml se monta como volumen en runtime
EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT ["/app/polaris", "-config", "/app/config.yaml"]
