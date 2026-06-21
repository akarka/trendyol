# Implementasyon Planı

## Mimari Hedef

```
Trendyol → Cloudflare Tunnel → Go HTTP Server → MySQL 8 → Printer
                                      ↕
                              React + Tailwind SPA (embed)
```

## Session İlerleme Tablosu

| # | Session | Kapsam | Durum |
|---|---------|--------|-------|
| S01 | [Infra + DB](sessions/S01-infra-db.md) | MySQL Docker, şema, go.mod, eski Supabase kodu temizliği | ✅ |
| S02 | [Go HTTP Server + Webhook](sessions/S02-go-server-webhook.md) | chi router, webhook endpoint, DB write, print goroutine | ✅ |
| S03 | [REST API + JWT Auth](sessions/S03-rest-api-auth.md) | login, orders, print-jobs, settings, printer-status endpoint'leri | ✅ |
| S04 | [React Setup + Auth UI](sessions/S04-react-setup.md) | Vite + React + Tailwind, login sayfası, API client, embed | ✅ |
| S05 | [React Pages](sessions/S05-react-pages.md) | Order list, order detail, printer status/log, settings | ✅ |
| S06 | [Cloudflare + Docker](sessions/S06-tunnel-docker.md) | cloudflared, docker-compose full stack, e2e test | ✅ |
| S07 | [Zeytuni Ürün İnceleme](sessions/S07-zeytuni-urunler-review.md) | ürün verisi kalite analizi + `products` şema/import taslağı (kod yok) → [rapor](zeytuni-urunler-rapor.md) | ✅ |
| S08 | [Manuel Sipariş + Katalog](sessions/S08-manuel-siparis-katalog.md) | normalize products şema/import/export (.xlsx round-trip), `/api/orders/manual`, canvas bitmap baskı, ManualOrderPage | ✅ |

## Durum Kodları
- 🔲 Bekliyor
- 🔄 Devam ediyor
- ✅ Tamamlandı

## Bağımlılık Zinciri

S01 → S02 → S03 → S04 → S05 → S06

Her session bir öncekinin çıktısına bağlı. S07 ana zincire bağımlı değil — paralel discovery (kod/şema değiştirmez).

## S07 — Zeytuni Ürün Verisi (discovery)

`docs/zeytuni-urunler-rapor.md`: 67 ürün / 14 kategori (15 AKTIF · 52 PASIF) kalite analizi.
Kararlar: join **SKU** üzerinden (Trendyol endpoint gelince barkoda geçiş); import **tüm 67** ürün (`is_active` flag); çift SKU `ZYT-TER-012` dükkana sorulacak.
`products` şema taslağı + idempotent `cmd/import-products` planı raporda — **uygulanmadı**, ayrı implementasyon session'ı bekliyor.

## Uygulanan Durum (S01–S05)

Aşağıdaki kısımlar fiilen kodlandı; bu bölüm planı değil mevcut gerçeği yansıtır.

**Backend (Go)**
- `config/config.go` — MySQL DSN + JWT + webhook + printer env'leri (Supabase kaldırıldı)
- `internal/db/` — `db.go` (sqlx bağlantı), `queries.go` (orders/print_jobs/settings/users + `InsertOrder` idempotent `INSERT IGNORE`)
- `internal/server/` — chi router; `webhook.go` (BasicAuth, daima 200), `auth.go` (bcrypt login → JWT), `api.go` (orders/reprint/printer-status/logs/settings), `middleware.go`
- `internal/auth/` — `jwt.go` (HS256, 24s), `middleware.go` (Bearer zorunlu)
- `cmd/seed/main.go` — `--username/--password` ile bcrypt admin kullanıcı
- `internal/listener/` (Supabase) kaldırıldı; parser/printer paketlerine dokunulmadı

**Frontend (React) + embed**
- `web/` — Vite + React 18 + TS + Tailwind v3; `src/api` (axios, Bearer + 401 auto-logout), `AuthContext`, `ToastContext`, router + `ProtectedRoute`, `Layout` (desktop sidebar / mobil alt nav)
- Sayfalar: `LoginPage`, `OrdersPage` (filtre + tablo/kart + sayfalama + baskı), `OrderDetailModal`, `PrinterPage` (10s refetch), `SettingsPage`
- `web/embed.go` `//go:embed all:dist` + `internal/server/server.go` `spaHandler` (bilinmeyen rota → index.html); `cmd/app/main.go` `web.Dist()` geçiyor
- `Dockerfile` — `node:20` web-builder multi-stage → `web/dist` Go builder'a kopyalanır

**Plandan sapmalar / notlar**
- Toast, `components/Toast.tsx` yerine `context/ToastContext.tsx` (provider + `useToast`)
- `web/dist` ve `node_modules` git'e dahil değil (`web/.gitignore`); embed çıktısı Docker build sırasında üretilir
- Go toolchain geliştirme makinesinde yok; `go build` doğrulaması Docker `golang:1.22-alpine` builder stage'inde yapılıyor

**Kalan:** S06 (Cloudflare Tunnel + docker-compose full stack + e2e)

## Hedef DB Şeması

```sql
-- Mevcut; kolon isimlerine dokunma, parser bağlı
CREATE TABLE trendyol_orders (
  uuid           CHAR(36)     PRIMARY KEY DEFAULT (UUID()),
  order_id       VARCHAR(255) NOT NULL,
  order_number   VARCHAR(255) NOT NULL,
  package_status VARCHAR(50)  NOT NULL,
  payload        JSON         NOT NULL,
  created_at     DATETIME     DEFAULT CURRENT_TIMESTAMP,
  updated_at     DATETIME     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uq_order_status (order_id, package_status)
);

CREATE TABLE users (
  id            INT          AUTO_INCREMENT PRIMARY KEY,
  username      VARCHAR(100) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  role          VARCHAR(50)  NOT NULL DEFAULT 'admin',
  created_at    DATETIME     DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE print_jobs (
  id           INT          AUTO_INCREMENT PRIMARY KEY,
  order_id     VARCHAR(255) NOT NULL,
  status       VARCHAR(50)  NOT NULL,  -- queued|success|failed
  error_msg    TEXT,
  attempted_at DATETIME     DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_order_id (order_id)
);

CREATE TABLE settings (
  `key`      VARCHAR(100) PRIMARY KEY,
  value      TEXT         NOT NULL,
  updated_at DATETIME     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

## Go Bağımlılıkları (hedef)

```
github.com/go-chi/chi/v5
github.com/jmoiron/sqlx
github.com/go-sql-driver/mysql
github.com/golang-jwt/jwt/v5
golang.org/x/crypto
```

## React Bağımlılıkları (hedef)

```
react@18, react-dom@18
react-router-dom@6
@tanstack/react-query@5
axios
tailwindcss@3, @tailwindcss/forms
```
