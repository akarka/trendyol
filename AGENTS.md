# AGENTS.md — Agentic Handoff Dökümanı

Bu dosya bu codebase'i devralan AI agent'lar içindir.

---

## Proje Kimliği

**Rol:** Dükkan yönetim sisteminin termal yazıcı modülü. Bağımsız Go daemon; dış sisteme printer interface olarak sunulacak.
**Repo:** `github.com/akarka/trendyol`
**Dil/Runtime:** Go 1.22, Docker (Alpine), Supabase (Deno Edge Fn + PostgreSQL Realtime)

---

## Mimari — Tek Cümle

```
Trendyol POST → Supabase Edge Fn (auth + idempotency) → trendyol_orders(JSONB)
→ Realtime WebSocket → Go listener → parser → printer (ESC/POS veya txt)
```

---

## Dosya Haritası (önem sırasıyla)

```
cmd/app/main.go                          # entry point; pipeline kurulumu
config/config.go                         # env var yükleme
internal/listener/supabase_client.go     # WebSocket/Phoenix; exponential backoff reconnect
internal/parser/payload_parser.go        # DBRow → Order struct; nested JSONB unwrap
internal/printer/txt_printer.go          # AKTİF — output.txt'e append
internal/printer/escpos_printer.go       # STUB — ESC/POS impl yok
internal/printer/digital_printer.go      # UNUSED — main'de çağrılmıyor
internal/alerter/system_alerter.go       # log sarmalayıcı; OS notify TODO
supabase/functions/trendyol-webhook/index.ts  # Deno Edge Fn; Basic Auth + DB insert
supabase/migrations/20260301_init.sql    # tablo şeması, RLS, indeksler
supabase/migrations/20260302_enable_realtime.sql  # realtime publication
.env.example                             # tüm config anahtarları burada
docker-compose.yml                       # volume: ./output.txt, devices: yorum satırında
```

---

## Veri Akışı Detayı

**Webhook → DB:**
- Edge Fn: Basic Auth (`WEBHOOK_USERNAME`/`WEBHOOK_PASSWORD`)
- `trendyol_orders` INSERT; `UNIQUE(order_id, package_status)` idempotency
- Trendyol'a her zaman 200 dön (retry önleme); hata internal log'a

**DB → Printer:**
- `DBRow.payload` (JSONB) → `Order` struct (nested unwrap, `payload_parser.go:43`)
- `PackageStatus` değerleri: `Created`, `Cancelled`, `Delivered`, `UnSupplied`
- `main.go` sadece `PrintToTXT` çağırıyor (`txt_printer.go:13`)

**WebSocket:**
- Phoenix protocol, `realtime:public:trendyol_orders` topic, `INSERT` event
- `gorilla/websocket` — `supabase-community/realtime-go` kullanılmıyor (Alpine'de derleme sorunu)
- Heartbeat 30s, reconnect: 5s → exp backoff → 60s max

---

## Tamamlanan / Bekleyen

| Durum | Konu |
|-------|------|
| ✅ | Supabase Edge Fn (auth, idempotency, DB insert) |
| ✅ | Realtime WebSocket listener + reconnect |
| ✅ | Order parser (DBRow → Order) |
| ✅ | TXT printer (output.txt, TEST_MODE) |
| ✅ | Docker deployment (multi-stage, Alpine) |
| ✅ | Test scripts (test_webhook_status.ps1) |
| 🔲 | ESC/POS printer impl (`escpos_printer.go` stub) |
| 🔲 | OS notification/ses (`system_alerter.go` stub) |
| 🔲 | Gerçek yazıcı Docker device mount (docker-compose.yml'de yorum satırında) |
| 🔲 | `digital_printer.go` → ya sil ya da aktif et |

---

## Kritik Kısıtlar

1. `SUPABASE_URL` ve `SUPABASE_ANON_KEY` olmadan servis panikler (`config.go:26`)
2. `TEST_MODE=false` iken `PRINTER_DEVICE` zorunlu (`config.go:29`)
3. Phoenix mesaj yapısı iki farklı payload formatını handle ediyor (`supabase_client.go:116-130`) — değiştirme
4. ESC/POS için library seçilmedi; `escpos_printer.go` düz skeleton
5. `output.txt` Docker volume ile host'a bağlı — container restart'ta kaybolmaz
6. RLS: anon key sadece SELECT; Edge Fn service_role key ile INSERT

---

## Görev Başlangıç Noktaları

**ESC/POS implement et:**
→ `internal/printer/escpos_printer.go` → library ekle → `main.go`'daki `PrintToTXT`'i `Print` ile değiştir → `docker-compose.yml`'de `devices:` bloğunu aç

**OS bildirim ekle:**
→ `internal/alerter/system_alerter.go` → `NotifySuccess`/`NotifyError` fonksiyonlarını doldur

**Yeni sipariş alanı ekle:**
→ `internal/parser/payload_parser.go`'daki `Order` struct'a alan ekle → printer'da kullan

**Yeni printer tipi ekle:**
→ `internal/printer/` altına yeni dosya → `main.go`'da switch et

---

## Test Komutu

```powershell
powershell.exe -ExecutionPolicy Bypass -File .\test_webhook_status.ps1 [-Status "Created"|"Cancelled"|"Delivered"|"UnSupplied"]
```

Çıktı: `output.txt` — `tail -f output.txt` ile izle.
