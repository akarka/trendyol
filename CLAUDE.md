# CLAUDE.md — trendyol termal yazıcı modülü

## Proje Rolü

Dükkan yönetim sisteminin **termal yazıcı modülü**. Bağımsız Go servisi + React admin SPA.
Trendyol webhook'larını Cloudflare Tunnel üzerinden alır, MySQL'e yazar, anlık basar, web arayüzü üzerinden yönetilir.

## Hedef Mimari

```
Trendyol
   │ HTTPS POST
   ▼
Cloudflare Tunnel ──▶ Go HTTP Server (Docker)
                          ├── POST /webhook/trendyol   → MySQL → print goroutine
                          ├── POST /api/auth/login     → JWT
                          ├── GET  /api/orders         → sipariş listesi
                          ├── POST /api/orders/:id/print → yeniden baskı
                          ├── GET  /api/printer/status
                          └── GET  /                   → React SPA (embed)

MySQL 8 (Docker)          Cloudflared (Docker)
```

## Hedef Dizin Yapısı

```
cmd/app/main.go
config/config.go
internal/
  server/        # HTTP router, middleware, handler'lar
  db/            # MySQL bağlantısı, query'ler
  auth/          # JWT üretme/doğrulama
  printer/       # ESC/POS ve txt printer
  parser/        # Trendyol payload → Order struct
  alerter/       # log sarmalayıcı
web/             # React + Tailwind kaynak
  src/
  dist/          # build çıktısı → Go'ya embed
docs/
  PLAN.md        # implementasyon planı
  sessions/      # session handoff dökümanları
```

## Veritabanı Şeması

```sql
trendyol_orders (uuid PK, order_id, order_number, package_status, payload JSONB, created_at, updated_at)
users           (id PK, username, password_hash, role, created_at)
print_jobs      (id PK, order_id FK, status, error_msg, attempted_at)
settings        (key PK, value, updated_at)
```

## Tech Stack

| Katman | Teknoloji |
|--------|-----------|
| Backend | Go 1.22, `chi` router, `sqlx` + `go-sql-driver/mysql` |
| Auth | `golang-jwt/jwt/v5`, `bcrypt` |
| DB | MySQL 8 (Docker) |
| Frontend | React 18, Vite, Tailwind CSS v3, TanStack Query |
| Tunnel | Cloudflare Tunnel (cloudflared Docker) |
| Deploy | Docker Compose |

## Mevcut Durum

S01–S05 tamamlandı. Kalan: **S06** (Cloudflare Tunnel + docker-compose full stack + e2e).

- `internal/listener/` → SİLİNDİ (Supabase WebSocket / Phoenix kaldırıldı)
- `internal/server/` → ✅ chi router, webhook, auth, api, spaHandler (SPA embed)
- `internal/db/` → ✅ sqlx bağlantı + query'ler
- `internal/auth/` → ✅ JWT (HS256, 24s) + middleware
- `cmd/seed/` → ✅ bcrypt admin kullanıcı seed
- `web/` → ✅ React 18 + Vite + Tailwind v3 SPA; `web/dist` Go binary'ye embed
- `config/config.go` → ✅ Supabase kaldırıldı, MySQL + JWT + webhook + printer env'leri

## Roller

Şimdilik: `admin` tek rol. JWT payload'ında `role` claim ile middleware hazır, genişletme tek satır.

## Env Değişkenleri (hedef)

```
MYSQL_DSN=user:pass@tcp(mysql:3306)/trendyol?parseTime=true
JWT_SECRET=...
WEBHOOK_USERNAME=...
WEBHOOK_PASSWORD=...
PRINTER_DEVICE=/dev/usb/lp0
TEST_MODE=true
LOG_LEVEL=info
CF_TUNNEL_TOKEN=...        # Cloudflare Tunnel token
```

## Asistan Davranış Kuralları

- Cevaplar kısa ve doğrudan. Gereksiz açıklama yok.
- Kod değişikliği: sadece ilgili satırları göster, tüm dosyayı tekrarlama.
- Övgü, teşekkür, "harika soru" gibi ifadeler yok.
- Yorum satırı yazma; yazmak gerekiyorsa tek satır, sadece "neden" açıklanır.
- Seçenekler sorulmadıkça sunma; doğrudan en iyi yaklaşımı uygula.
- Türkçe konuş.
- Session başında daima `docs/sessions/` klasörünü oku — aktif session'ı belirle.
