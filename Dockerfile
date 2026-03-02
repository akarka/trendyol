# Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY . .
RUN go mod tidy
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/print-relay ./cmd/app

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/print-relay .

CMD ["./print-relay"]
