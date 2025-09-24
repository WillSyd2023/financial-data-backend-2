FROM golang:1.24-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal

RUN go build -o /app/ingestor ./cmd/go-ingestor

FROM alpine:latest

WORKDIR /app

# grab compiled code from the top image
COPY --from=builder /app/ingestor .

# grab config YML from local
COPY config/config.yml ./config/config.yml

CMD ["./ingestor"]