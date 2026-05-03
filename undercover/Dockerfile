# Stage 1: build all binaries (CGO enabled for go-duckdb / DuckDB requires glibc)
FROM golang:1.25-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc g++ \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /ingestor   ./cmd/ingestor
RUN CGO_ENABLED=1 GOOS=linux go build -o /paquetier  ./cmd/paquetier
RUN CGO_ENABLED=1 GOOS=linux go build -o /overseer   ./cmd/overseer
RUN CGO_ENABLED=0 GOOS=linux go build -o /registrar  ./cmd/registrar

# Stage 2: ingestor runtime (needs glibc for cgo binary)
FROM debian:bookworm-slim AS ingestor

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /ingestor .

EXPOSE 8080

ENTRYPOINT ["./ingestor"]

# Stage 3: paquetier runtime (needs glibc for cgo binary)
FROM debian:bookworm-slim AS paquetier

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /paquetier .

VOLUME ["/data"]

ENTRYPOINT ["./paquetier"]

# Stage 4: registrar runtime
FROM debian:bookworm-slim AS registrar

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /registrar .

EXPOSE 8082

ENTRYPOINT ["./registrar"]

# Stage 5: overseer runtime (pure Go, scratch-compatible but slim for consistency)
FROM debian:bookworm-slim AS overseer

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /overseer .

EXPOSE 8081

ENTRYPOINT ["./overseer"]

