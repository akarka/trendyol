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
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/seed ./cmd/seed
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/import-products ./cmd/import-products
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/export-products ./cmd/export-products

FROM alpine:latest

WORKDIR /app

# mysqldump/mysql CLI: DB yedek/geri yükleme (internal/server/admin.go)
RUN apk add --no-cache mysql-client

COPY --from=builder /app/print-relay .
COPY --from=builder /app/seed .
COPY --from=builder /app/import-products .
COPY --from=builder /app/export-products .
# import-products varsayılan kaynağı (bootstrap); işletme xlsx'i sonra mount edilerek re-import edilir
COPY data/urun_listesi.utf8.csv ./data/urun_listesi.utf8.csv

CMD ["./print-relay"]
