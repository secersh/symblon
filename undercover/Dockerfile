# Stage 1: build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /ingestor ./cmd/ingestor

# Stage 2: minimal runtime image
FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /ingestor .

EXPOSE 8080

ENTRYPOINT ["./ingestor"]
