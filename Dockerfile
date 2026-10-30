FROM golang:1.26.5-alpine AS builder

RUN adduser -D appuser
USER appuser

WORKDIR /app

COPY go.mod go.sum /

RUN go mod download

COPY . .

RUN go build -o api ./cmd/api

FROM alpine:latest

RUN adduser -D appuser
USER appuser

WORKDIR /app

COPY --from=builder --chown=appuser:appuser /app/api .

EXPOSE 8080

CMD ["./api"]


