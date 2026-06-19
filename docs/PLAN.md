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
| S01 | [Infra + DB](sessions/S01-infra-db.md) | MySQL Docker, şema, go.mod, eski Supabase kodu temizliği | 🔲 |
| S02 | [Go HTTP Server + Webhook](sessions/S02-go-server-webhook.md) | chi router, webhook endpoint, DB write, print goroutine | 🔲 |
| S03 | [REST API + JWT Auth](sessions/S03-rest-api-auth.md) | login, orders, print-jobs, settings, printer-status endpoint'leri | 🔲 |
| S04 | [React Setup + Auth UI](sessions/S04-react-setup.md) | Vite + React + Tailwind, login sayfası, API client, embed | 🔲 |
| S05 | [React Pages](sessions/S05-react-pages.md) | Order list, order detail, printer status/log, settings | 🔲 |
| S06 | [Cloudflare + Docker](sessions/S06-tunnel-docker.md) | cloudflared, docker-compose full stack, e2e test | 🔲 |

## Durum Kodları
- 🔲 Bekliyor
- 🔄 Devam ediyor
- ✅ Tamamlandı

## Bağımlılık Zinciri

S01 → S02 → S03 → S04 → S05 → S06

Her session bir öncekinin çıktısına bağlı.

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
