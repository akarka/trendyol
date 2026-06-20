# Dockerfile
FROM node:20-alpine AS web-builder
WORKDIR /web
COPY web/package*.json ./
RUN npm ci
COPY web/ .
RUN npm run build

FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# embed için React build çıktısı (web/.gitignore'da; bu yüzden builder stage'den kopyalanır)
COPY --from=web-builder /web/dist ./web/dist

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/print-relay ./cmd/app

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/print-relay .

CMD ["./print-relay"]
