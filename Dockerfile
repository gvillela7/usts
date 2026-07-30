# Stage 1: build the frontend assets with Parcel.
# The output (web/static) is embedded into the Go binary via //go:embed.
FROM node:22-alpine AS frontend

WORKDIR /build/web

COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/.postcssrc ./
# Static files (icons, manifest) live in the same dist dir Parcel writes to,
# so copy them in before building.
COPY web/static ./static
# Tailwind scans the templates for class names (see source() in styles.css).
COPY web/templates ./templates
COPY web/assets ./assets

RUN npm run build

# Stage 2: generate the templ files and build the Go binary.
FROM golang:1.24-alpine AS builder

# github.com/mattn/go-sqlite3 is a cgo package, so a C toolchain is required.
RUN apk add --no-cache gcc musl-dev

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

# Keep in sync with the templ version in go.mod.
RUN go install github.com/a-h/templ/cmd/templ@v0.3.960

COPY . .
COPY --from=frontend /build/web/static ./web/static

ENV CGO_ENABLED=1

RUN templ generate && \
    go build -ldflags="-s -w" -o usts ./cmd/main.go

# Stage 3: runtime image.
# Alpine (not scratch) because the cgo binary links against musl.
FROM alpine:3.21

# The app resolves config.toml and database/*.sql relative to the working directory.
WORKDIR /app

COPY --from=builder /build/usts ./usts
COPY config.toml ./config.toml
COPY database/schema.sql database/seed.sql ./database/

EXPOSE 8000

ENTRYPOINT ["/app/usts"]
