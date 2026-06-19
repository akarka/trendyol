# Teknik Mimari — Trendyol Termal Yazıcı Modülü

## Proje Rolü

Bu servis, dükkan yönetim sisteminin **termal yazıcı modülüdür**. Trendyol'dan gelen sipariş webhook'larını yerel ağda çalışan termal yazıcıya iletir. Supabase, yerel ağı internete açmadan cloud→local tüneli sağlar.

---

## Stack

| Katman | Teknoloji |
|--------|-----------|
| Webhook alıcı | Supabase Edge Function (Deno/TypeScript) |
| Veritabanı | Supabase PostgreSQL |
| Realtime kanal | Supabase Realtime (PostgreSQL CDC → WebSocket) |
| Local agent | Go 1.22, `gorilla/websocket` |
| Containerization | Docker + Docker Compose (Alpine multi-stage) |
| Printer protokolü | ESC/POS (stub) / txt dosya (aktif) |

---

## Mimari Pipeline

```
Trendyol
  │  HTTP POST (Basic Auth)
  ▼
Supabase Edge Fn  ──(idempotency check)──▶  trendyol_orders (PostgreSQL)
                                                    │
                                              Realtime CDC
                                                    │ WebSocket/Phoenix
                                                    ▼
                                          Go Daemon (Docker)
                                          ├─ listener  (WebSocket client)
                                          ├─ parser    (JSON → Order struct)
                                          └─ printer   (ESC/POS veya txt)
```

---

## Paketler

### `supabase/functions/trendyol-webhook` (Deno)
- Basic Auth doğrulama (`WEBHOOK_USERNAME` / `WEBHOOK_PASSWORD`)
- `trendyol_orders` INSERT; `UNIQUE(order_id, package_status)` ile idempotency
- Trendyol'a daima 200 döner (retry bastırma); hata loga yazılır

### `config`
- Env var yükleme; `SUPABASE_URL` + `SUPABASE_ANON_KEY` yoksa panic
- `TEST_MODE=false` iken `PRINTER_DEVICE` zorunlu

### `internal/listener`
- Phoenix protocol WebSocket istemcisi
- Reconnect: 5s başlangıç, exponential backoff, 60s max
- Heartbeat: 30s
- `gorilla/websocket` — `supabase-community/realtime-go` yerine (Alpine derleme uyumsuzluğu)

### `internal/parser`
- `DBRow` (top-level DB kolonları) + `payload` (JSONB) → `Order` struct
- Zorunlu alanlar: `order_id`, `order_number`, en az 1 ürün satırı

### `internal/printer`
| Dosya | Durum | Açıklama |
|-------|-------|----------|
| `txt_printer.go` | **Aktif** | `output.txt`'e append; TEST_MODE |
| `escpos_printer.go` | **Stub** | ESC/POS iskelet; library seçilmedi |
| `digital_printer.go` | **Unused** | Çağrılmıyor |

### `internal/alerter`
- `log.Printf` sarmalayıcı; OS ses/bildirim TODO

---

## Veritabanı Şeması

```sql
trendyol_orders (
  uuid            UUID PRIMARY KEY,
  order_id        VARCHAR(255) NOT NULL,
  order_number    VARCHAR(255) NOT NULL,
  package_status  VARCHAR(50)  NOT NULL,   -- Created|Cancelled|Delivered|UnSupplied
  payload         JSONB        NOT NULL,
  created_at      TIMESTAMPTZ,
  updated_at      TIMESTAMPTZ,
  UNIQUE (order_id, package_status)
)
```

**RLS:**
- `service_role` → INSERT (Edge Fn)
- `anon` → SELECT (Go listener Realtime)

---

## Ortam Değişkenleri

| Değişken | Zorunlu | Açıklama |
|----------|---------|----------|
| `SUPABASE_URL` | ✅ | `https://<ref>.supabase.co` |
| `SUPABASE_ANON_KEY` | ✅ | Realtime aboneliği için |
| `SUPABASE_SERVICE_ROLE_KEY` | ✅ (Edge Fn) | DB insert için |
| `WEBHOOK_USERNAME` | ✅ (Edge Fn) | Basic Auth kullanıcı |
| `WEBHOOK_PASSWORD` | ✅ (Edge Fn) | Basic Auth şifre |
| `TEST_MODE` | — | `true` → txt dosyası; `false` → ESC/POS |
| `PRINTER_DEVICE` | TEST_MODE=false ise ✅ | `/dev/usb/lp0` veya dosya yolu |
| `LOG_LEVEL` | — | Varsayılan: `info` |

---

## Kurulum

```bash
# 1. Env dosyasını hazırla
cp .env.example .env
# .env içini doldur

# 2. Başlat
docker-compose up -d --build

# 3. Test (Trendyol webhook simülasyonu)
powershell.exe -ExecutionPolicy Bypass -File .\test_webhook_status.ps1

# Belirli durum testi
powershell.exe -ExecutionPolicy Bypass -File .\test_webhook_status.ps1 -Status "Cancelled"

# Çıktı izle
tail -f output.txt
```

---

## Gerçek Yazıcıya Geçiş

1. `.env`: `TEST_MODE=true` satırını kaldır, `PRINTER_DEVICE=/dev/usb/lp0` yaz
2. `docker-compose.yml`: `devices:` bloğunu yorumdan çıkar
3. `main.go`: `PrintToTXT` → `printer.Print(cfg.PrinterDevice, order)` olarak değiştir
4. `escpos_printer.go`'ya ESC/POS library implement et
5. `docker-compose up -d --build`

---

## Kritik Notlar

- Phoenix protokolü iki farklı Supabase payload formatını handle ediyor (`supabase_client.go:116-130`) — dokunma
- `output.txt` Docker bind-mount ile host'ta tutulur (`docker-compose.yml` volume tanımı)
- Trendyol'a her zaman HTTP 200 dön; hata durumunda internal log, retry isteme
