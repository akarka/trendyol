# README_TECH — Teknik Genel Bakış

Dükkan yönetim sisteminin **termal yazıcı modülü**. Bağımsız Go servisi + gömülü React admin SPA.
Trendyol sipariş webhook'larını alır, MySQL'e yazar, etiket basar; ayrıca panelden **manuel sipariş** girişi ve **ürün kataloğu** yönetimi sağlar.

> Not: Proje başta Supabase (Edge Functions + PostgreSQL + Realtime) üzerine kuruluydu; **MySQL + Go (chi) + React** yığınına taşındı. Eski Supabase katmanı tamamen kaldırıldı.

## Mimari

```
Trendyol ──HTTPS POST──▶ Cloudflare Tunnel ──▶ Go HTTP Server (Docker, :8080)
                                                  ├─ POST /webhook/trendyol  → MySQL → print goroutine
                                                  ├─ POST /api/orders/manual → MySQL → print goroutine
                                                  ├─ /api/...                → REST (JWT)
                                                  └─ /                       → React SPA (go:embed)
                                                        │
                                MySQL 8 (Docker) ◀──────┘
```

Baskı görseli (etiket bitmap'i) şimdilik **tarayıcıda canvas** ile üretilip `window.print()` ile laser yazıcıya gönderilir. Sunucu tarafı baskı hattı (`print_jobs` + goroutine) `TEST_MODE`'da `output.txt`'e kayıt düşer; gerçek termal yazıcı (ESC/POS raster) ileride aynı bitmap'ten beslenecek.

## Teknoloji Yığını

| Katman | Teknoloji |
|--------|-----------|
| Backend | Go 1.22, `go-chi/chi/v5`, `jmoiron/sqlx` + `go-sql-driver/mysql` |
| Auth | `golang-jwt/jwt/v5` (HS256, 24s), `bcrypt` |
| DB | MySQL 8 (Docker) |
| Frontend | React 18, Vite, Tailwind v3, TanStack Query, axios |
| Embed | `web/dist` → `//go:embed` ile Go binary'ye gömülür |
| xlsx | `internal/xlsxlite` (stdlib `archive/zip` + `encoding/xml`; harici lib yok) |
| Tunnel | Cloudflare Tunnel (cloudflared, Docker) |
| Deploy | Docker Compose |

Harici Go bağımlılıkları minimum tutulur; bitmap render bilinçli olarak tarayıcıya bırakıldığı için backend `golang.org/x/image` gibi bir font/raster bağımlılığı içermez.

## Dizin Yapısı

```
cmd/
  app/             # print-relay: HTTP server + print goroutine
  seed/            # bcrypt admin kullanıcı
  import-products/ # CSV/xlsx → products (idempotent)
  export-products/ # products → .xlsx (işletme bakımı)
config/            # env yükleme
internal/
  server/          # chi router, webhook, auth, api, manuel sipariş, SPA handler
  db/              # sqlx bağlantı, query'ler, migrate (EnsureCatalogTables)
  auth/            # JWT üret/doğrula + middleware
  parser/          # Trendyol payload → Order
  printer/         # txt printer (test) + escpos stub
  xlsxlite/        # stdlib xlsx oku/yaz
web/               # React + Tailwind SPA (dist embed edilir)
data/              # urun_listesi.utf8.csv (katalog bootstrap kaynağı)
docs/              # PLAN, schema.sql, session handoff'ları
```

## Veri Akışları

**Webhook siparişi** (`internal/server/webhook.go`)
Basic Auth → ham gövde `parser.ParseOrder` → `trendyol_orders` `INSERT IGNORE` (UNIQUE `order_id,package_status` ile idempotent) → `print_jobs` queued → `printCh`. Trendyol retry'larını bastırmak için **daima 200** döner.

**Manuel sipariş** (`internal/server/manual_order.go`)
`{customer_name, lines:[{sku,quantity}]}` → her satır `products`'tan zenginleştirilir → `parser.Order` (payload Trendyol şemasıyla **birebir aynı**, böylece yeniden baskı aynı yoldan çalışır) → `trendyol_orders` (order_id `MAN-<unixnano>`, status `Created`) → `print_jobs` → `printCh`.

**Baskı hattı** (`cmd/app/main.go: runPrinter`)
`printCh` (buffer 64) tüketilir; `TEST_MODE=true` → `printer.PrintToTXT` (`output.txt`), aksi halde `printer.Print` (ESC/POS, şu an stub). Sonuç `print_jobs.status` (success/failed) olarak yazılır.

## Veritabanı Şeması

Kanonik tanım: [`docs/schema.sql`](docs/schema.sql). Tablolar:

- `trendyol_orders` (uuid PK, order_id, order_number, package_status, payload JSON, ts) — `UNIQUE(order_id, package_status)`
- `users` (bcrypt hash, role)
- `print_jobs` (order_id, status, error_msg, attempted_at)
- `settings` (key PK, value) — ör. `label_layout` (A4 etiket yerleşimi JSON)
- `categories`, `brands`, `products` (normalize katalog; FK'li)

`products`: sku PK, barcode UNIQUE, name (iç ad), marketplace_name, category_id/brand_id FK, net_weight+unit, price, vat_rate, is_active, **needs_fix**, description.

> `mysql_data` volume kalıcı olduğundan `schema.sql` init script'i mevcut DB'de yeniden çalışmaz. Katalog tabloları `internal/db/migrate.go: EnsureCatalogTables` ile app/import/export açılışında garanti edilir.

## HTTP Endpoint'leri

| Method | Path | Koruma | Açıklama |
|--------|------|--------|----------|
| POST | `/webhook/trendyol` | Basic Auth | Trendyol sipariş webhook'u (daima 200) |
| POST | `/api/auth/login` | — | bcrypt → JWT |
| GET | `/api/orders` | JWT | sipariş listesi (filtre/sayfalama) |
| POST | `/api/orders/manual` | JWT | manuel sipariş girişi |
| GET | `/api/orders/{id}` | JWT | sipariş detayı |
| POST | `/api/orders/{id}/print` | JWT | yeniden baskı |
| GET | `/api/products` | JWT | katalog (`?active=1`) |
| GET | `/api/printer/status` | JWT | test_mode/device + son job'lar |
| GET | `/api/logs` | JWT | print_jobs |
| GET/PUT | `/api/settings[/{key}]` | JWT | ayarlar |
| GET | `/*` | — | React SPA (bilinmeyen rota → index.html) |

## Ürün Kataloğu

- **Bootstrap:** `data/urun_listesi.utf8.csv` (cp1254 kaynaktan UTF-8'e çevrilmiş). `cmd/import-products` normalize eder (fiyat/%/gramaj), kategori+marka tablolarına dedup eder, `products`'a `ON DUPLICATE KEY UPDATE` (SKU) ile idempotent yazar.
- **Veri sorunları:** tüm 67 ürün içe alınır; eksik gramaj ve çift SKU `needs_fix=1` ile işaretlenir (import log'una düşer).
- **Round-trip:** `cmd/export-products` → `.xlsx` (import ile uyumlu başlıklar). İşletme Excel'de düzeltir → aynı dosya `--file urunler.xlsx` ile re-import edilir.

```bash
docker compose run --rm print-relay ./import-products
docker compose run --rm -v "$PWD:/out" print-relay ./export-products --out /out/urunler.xlsx
docker compose run --rm -v "$PWD:/out" print-relay ./import-products --file /out/urunler.xlsx
```

## Etiket Baskısı

- Etiket, `output.txt` düzeninde **monokrom canvas bitmap** olarak üretilir (`web/src/lib/label.ts`). Tarayıcı Türkçe fontu native render eder → encoding sorunu yok.
- **A4 şablonlu sticker yerleşimi** ayarlanabilir ve `settings.label_layout` (JSON) altında saklanır: sayfa ebadı, sütun/satır, etiket en/boy, kenar ve ara boşluklar, hücre iç padding (`web/src/lib/labelLayout.ts`). Ayarlar sayfasından düzenlenir; Manuel Sipariş'te hangi hücreye basılacağı seçilir.
- Baskı: `@page size:A4; margin:0` ile etiket mm cinsinden tam hücreye yerleştirilir; yazıcı **%100 / gerçek boyut** ile basmalıdır.
- **Termal geçiş (gelecek):** aynı canvas PNG'si sunucuya POST edilip `GS v 0` ESC/POS raster'a çevrilecek (1-bit, 384/576 dot genişlik).

## Binary'ler

| Binary | Komut | İş |
|--------|-------|----|
| `print-relay` | `./print-relay` (CMD) | HTTP server + baskı goroutine |
| `seed` | `./seed --username u --password p` | bcrypt admin kullanıcı |
| `import-products` | `./import-products [--file ...]` | katalog import |
| `export-products` | `./export-products [--out ...]` | katalog → xlsx |

## Env Değişkenleri

| Değişken | Zorunlu | Açıklama |
|----------|---------|----------|
| `MYSQL_DSN` | ✅ | `user:pass@tcp(mysql:3306)/trendyol?parseTime=true&charset=utf8mb4` |
| `JWT_SECRET` | ✅ | JWT imzalama |
| `WEBHOOK_USERNAME` / `WEBHOOK_PASSWORD` | ✅ | webhook Basic Auth |
| `TEST_MODE` | — | `true` → baskı `output.txt`'e |
| `PRINTER_DEVICE` | `TEST_MODE=false` iken ✅ | ör. `/dev/usb/lp0` |
| `LOG_LEVEL` | — | varsayılan `info` |
| `CF_TUNNEL_TOKEN` | tunnel profilinde | Cloudflare Tunnel |

## Çalıştırma

```bash
docker compose up -d --build                      # mysql + print-relay
docker compose run --rm print-relay ./seed --username admin --password <pass>
docker compose run --rm print-relay ./import-products
# http://localhost:8080  (admin / <pass>)
```

`cloudflared` servisi `tunnel` profilindedir; canlı için: `docker compose --profile tunnel up -d`.

## Notlar / Sınırlar

- Webhook akışı henüz canlı (Trendyol) ortamda test edilmedi; manuel akış ondan bağımsız çalışır.
- ESC/POS gerçek baskı (`internal/printer/escpos_printer.go`) stub; termal yazıcı alınınca tamamlanacak.
- `Zeytuni_Ops/` işletmenin ayrı git reposudur; bu projeye dahil değildir (`.gitignore`/`.dockerignore` ile hariç). Katalog kaynağı projeye `data/` altında kopyalanmıştır.
