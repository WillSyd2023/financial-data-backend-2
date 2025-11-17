FROM golang:1.24-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/go-api-service ./cmd/go-api-service
COPY ./internal ./internal

RUN go build -o /app/api-service ./cmd/go-api-service

FROM alpine:latest

WORKDIR /app

# grab compiled code from the top image
COPY --from=builder /app/api-service .

EXPOSE 8000

CMD ["./api-service"]