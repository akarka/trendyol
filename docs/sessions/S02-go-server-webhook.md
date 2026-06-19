# S02 — Go HTTP Server + Webhook

**Durum:** 🔲 Bekliyor
**Bağımlılık:** S01 tamamlanmış olmalı
**Sonraki:** S03-rest-api-auth.md

---

## Bu Session Yapılacaklar

### 1. `internal/server/` paketi oluştur

```
internal/server/
  server.go       # chi router kurulumu, route tanımları
  webhook.go      # POST /webhook/trendyol handler
  middleware.go   # Basic Auth (webhook için), JWT (API için)
```

### 2. Webhook handler (`webhook.go`)

```
POST /webhook/trendyol
  → Basic Auth doğrula (WEBHOOK_USERNAME / WEBHOOK_PASSWORD)
  → Body'i parse et (Trendyol JSON)
  → parser.ParseOrder() çağır
  → trendyol_orders'a INSERT (idempotency: UNIQUE(order_id, package_status) → duplicate'i 200 ile geç)
  → print_jobs kaydı aç (status: queued)
  → printer goroutine'e gönder (buffered channel)
  → Her durumda Trendyol'a 200 dön
```

### 3. Printer goroutine (`cmd/app/main.go`)

S01'deki stub'ı gerçek hale getir:
```go
printCh := make(chan *parser.Order, 64)
go runPrinter(printCh, cfg)   // PrintToTXT veya Print(cfg.PrinterDevice)
```
Goroutine: `print_jobs` kaydını `success`/`failed` olarak güncelle.

### 4. `internal/db/queries.go`

Şu an gereken query'ler:
```go
InsertOrder(order) error
InsertPrintJob(orderID, status, errMsg) (int64, error)
UpdatePrintJob(id int64, status, errMsg string) error
```

### 5. `main.go` güncelle

```go
db  := db.Connect(cfg)
srv := server.New(cfg, db, printCh)
srv.Start(":8080")
```

---

## Giriş Noktası

`internal/server/server.go` → `internal/server/webhook.go` → `internal/db/queries.go` → `cmd/app/main.go`

---

## Çıkış Kriteri

- [ ] `POST /webhook/trendyol` Basic Auth korumalı
- [ ] Webhook alınınca `trendyol_orders` + `print_jobs` kayıtları oluşuyor
- [ ] `output.txt`'e baskı yapılıyor (TEST_MODE=true)
- [ ] Duplicate webhook → 200, DB'de tekrar kayıt yok
- [ ] `test_webhook_status.ps1` çalıştırınca uçtan uca flow tamamlanıyor
- [ ] `go build ./...` hatasız

---

## Dikkat Edilecekler

- chi router: `r.Use(middleware.Logger)` koy — her request loglanır
- Trendyol'a **daima 200** dön — parse hatası bile olsa, sadece loga yaz
- Printer channel dolunca (`len == cap`) uyar ama block etme; `select { case ch <- order: default: log.Warn }`
- `internal/parser/payload_parser.go` **dokunma**
