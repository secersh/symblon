# Stage 1: build both binaries (CGO enabled for go-duckdb)
FROM golang:1.25-alpine AS builder

RUN apk --no-cache add gcc g++ musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /ingestor  ./cmd/ingestor
RUN CGO_ENABLED=1 GOOS=linux go build -o /paquetier ./cmd/paquetier

# Stage 2: ingestor runtime
FROM alpine:3.20 AS ingestor

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /ingestor .

EXPOSE 8080

ENTRYPOINT ["./ingestor"]

# Stage 3: paquetier runtime
FROM alpine:3.20 AS paquetier

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /paquetier .

VOLUME ["/data"]

ENTRYPOINT ["./paquetier"]
