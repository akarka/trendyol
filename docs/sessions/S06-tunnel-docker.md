# S06 — Cloudflare Tunnel + Docker Compose (full stack + e2e)

**Durum:** ✅ Tamamlandı
**Bağımlılık:** S01–S05 tamamlandı (✅)
**Sonraki:** Proje tamamlandı

> Yapılanlar: Dockerfile'a `seed` binary'si eklendi (runtime'a kopyalanıyor); compose'a mysql healthcheck + `print-relay` `service_healthy` bağımlılığı + `cloudflared` servisi (`profiles: ["tunnel"]`); `.env.example`'a `CF_TUNNEL_TOKEN`; README yeni mimariye (.env → compose → seed → Cloudflare → webhook) göre yeniden yazıldı.

> Bu dosya kendi kendine yeterlidir: yeni session yalnızca bunu okuyarak başlayabilir.
> Davranış kuralları için `CLAUDE.md` (kısa/doğrudan, Türkçe, yorum yok) geçerlidir.

---

## 0. Proje Özeti

Dükkan yönetim sisteminin termal yazıcı modülü + admin web arayüzü.
`Trendyol → Cloudflare Tunnel → Go HTTP Server (:8080) → MySQL 8 + Printer`, React SPA Go binary'ye embed.

Repo: `github.com/akarka/trendyol` — Go 1.22, React 18, MySQL 8, Docker.

---

## 1. Mevcut Gerçek Durum (S06 başlamadan önce hazır olanlar)

### Backend (Go) — hepsi çalışır
- `cmd/app/main.go` — entry point. Sıra: `config.Load()` → DB bağlantısı **30×2s retry** → `printCh := make(chan server.PrintTask, 64)` → `go runPrinter(...)` → `server.New(cfg, db, printCh, web.Dist())` → `srv.Start(":8080")`.
- `runPrinter`: `cfg.TestMode` ise `printer.PrintToTXT` (çıktı `output.txt`), değilse `printer.Print(cfg.PrinterDevice)`. Sonuç `print_jobs`'a `success`/`failed` yazılır.
- `config/config.go` — env'ler: `MYSQL_DSN`, `JWT_SECRET` (zorunlu), `WEBHOOK_USERNAME`, `WEBHOOK_PASSWORD` (zorunlu), `PRINTER_DEVICE`, `TEST_MODE` (`=="true"`), `LOG_LEVEL`. `TEST_MODE` true + `PRINTER_DEVICE` boş → `output.txt`. `TEST_MODE` false + device boş → panic.
- `internal/server/` — chi router. Route'lar (`server.go`):
  - `POST /webhook/trendyol` → **BasicAuth** (`WEBHOOK_USERNAME/PASSWORD`), `webhook.go`, **daima 200** döner (parse/DB hatası bile olsa loglar ve 200).
  - `POST /api/auth/login` → public, `auth.go`, body `{username,password}` → bcrypt → `{token, role}`.
  - JWT korumalı (`Authorization: Bearer`) grup, `api.go`:
    - `GET /api/orders?status=&limit=50&offset=0` → `OrderRow[]`
    - `GET /api/orders/{orderID}` → tek sipariş (en güncel `package_status`)
    - `POST /api/orders/{orderID}/print` → yeni `print_job` + kuyruğa it → `202 {job_id,status:"queued"}`
    - `GET /api/printer/status` → `{test_mode, device, jobs:[son 20]}`
    - `GET /api/logs` → son 100 `print_job`
    - `GET /api/settings` → `map[string]string`
    - `PUT /api/settings/{key}` → body `{value}`
  - `GET /*` → `spaHandler(web.Dist())`: dosya yoksa `index.html` (React Router fallback).
- `internal/auth/` — `jwt.go` (HS256, 24s, `Claims{user_id,username,role}`), `middleware.go` (Bearer zorunlu).
- `internal/db/` — `db.go` (sqlx), `queries.go`. `InsertOrder` → `INSERT IGNORE` (UNIQUE `order_id+package_status`, duplicate sessizce atlanır).
- `cmd/seed/main.go` — **ayrı binary**. `--username --password [--role admin]`, `MYSQL_DSN` env'inden bağlanır, bcrypt hash'leyip `users`'a upsert eder.
- `internal/parser/`, `internal/printer/` — dokunulmadı. `internal/listener/` (Supabase) silindi.

### Frontend (React) — build çalışır
- `web/` — Vite + React 18 + TS + Tailwind v3. `npm run build` → `web/dist/` (≈270 KB JS). `web/embed.go` `//go:embed all:dist` ile binary'ye gömer.
- Sayfalar: Login, Orders (filtre/tablo/kart/sayfalama/baskı), OrderDetailModal, Printer (10s refetch), Settings. Auth + Toast context, axios client (Bearer + 401 auto-logout).

### Şema (DB init)
- `docs/schema.sql` — 4 tablo (`trendyol_orders`, `users`, `print_jobs`, `settings`) + `INSERT IGNORE admin / $2a$10$PLACEHOLDER`. **PLACEHOLDER geçerli bir hash değil → seed çalıştırılana kadar login olmaz.**
- Compose'da `docker-entrypoint-initdb.d`'ye mount edilir; **yalnızca ilk başlatmada** (boş volume) çalışır.

### Mevcut `docker-compose.yml` (S06'da tamamlanacak)
- Servisler: `print-relay` (build `.`, ports `8080:8080`, env_file `.env`, volume `output.txt` + `/etc/localtime`, `depends_on: [mysql]` düz) ve `mysql:8` (ports `3306:3306`, schema.sql mount, `mysql_data` volume).
- **Eksik:** `cloudflared` servisi yok; mysql `healthcheck` yok (main.go retry'ladığı için kritik değil ama önerilir).

### Mevcut `Dockerfile`
- 3 stage: `node:20` web-builder (`npm ci` + `npm run build` → `/web/dist`) → `golang:1.22-alpine` builder (`COPY --from=web-builder /web/dist ./web/dist`, `go build -o print-relay ./cmd/app`) → `alpine` runtime (sadece `print-relay`).
- **Eksik:** `cmd/seed` runtime imajına dahil değil → container içinde seed çalıştırılamaz (alpine'de Go yok). Bkz. §3.

### `.env` / `.env.example`
- `.env` var (gerçek değerler, gitignore'da). `.env.example` var ama minimal — **`CF_TUNNEL_TOKEN` yok.**

### Test script'leri (hazır, Windows PowerShell)
- `test_webhook_status.ps1 [-Status Created|Cancelled|Delivered|UnSupplied]` — `.env`'i okur, `http://localhost:8080/webhook/trendyol`'a BasicAuth ile örnek sipariş POST'lar. Her çağrı benzersiz `id` üretir.
- `test_webhook_unique.ps1`, `test_webhook.ps1` — benzer.

---

## 2. Ön Koşul (kullanıcı yapar, kod değil)

1. cloudflare.com → Zero Trust → Networks → Tunnels → **Create a tunnel** (tipi: Cloudflared).
2. Tunnel adı: `trendyol-printer`. Token'ı kopyala → `.env`'e `CF_TUNNEL_TOKEN=...`.
3. **Public hostname** ekle: `printer.<domain>` → Service `http://print-relay:8080`.
   - Tek hostname yeterli: aynı servis hem `/webhook/trendyol`'u hem SPA/`/api`'yi sunar.
   - Trendyol webhook URL'i: `https://printer.<domain>/webhook/trendyol` (BasicAuth ile).

---

## 3. Yapılacaklar

### 3.1 Seed'i container'da çalıştırılabilir yap (KRİTİK)
Runtime alpine'de Go yok; `cmd/seed` imaja eklenmeli. **Önerilen:** `Dockerfile` builder stage'inde ikinci binary'yi de derle ve runtime'a kopyala:
```dockerfile
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/print-relay ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/seed ./cmd/seed
# runtime stage:
COPY --from=builder /app/seed .
```
Sonra: `docker compose exec print-relay /app/seed --username admin --password <pass>` (DSN container env'inden gelir).
Alternatif (imaja dokunmadan): `docker compose run --rm --entrypoint sh print-relay` yerine bcrypt hash'i offline üret + `users` tablosuna `UPDATE` — daha kırılgan, önerilmez.

### 3.2 `docker-compose.yml` — cloudflared + healthcheck
- `mysql`'e healthcheck ekle (`mysqladmin ping`), `print-relay` → `depends_on: mysql: condition: service_healthy`.
- `cloudflared` servisi ekle:
```yaml
  cloudflared:
    image: cloudflare/cloudflared:latest
    command: tunnel --no-autoupdate run
    environment:
      - TUNNEL_TOKEN=${CF_TUNNEL_TOKEN}
    depends_on: [print-relay]
    restart: unless-stopped
    profiles: ["tunnel"]   # token yokken crash etmesin; `--profile tunnel` ile aç
```
- Prod'da `3306:3306` portunu dışarı açma (yalnızca iç ağ). Geliştirmede kalabilir.

### 3.3 `.env.example` güncelle
`CF_TUNNEL_TOKEN=` satırını ekle. Tüm anahtarlar: `MYSQL_DSN, MYSQL_ROOT_PASSWORD, MYSQL_USER, MYSQL_PASSWORD, JWT_SECRET, WEBHOOK_USERNAME, WEBHOOK_PASSWORD, TEST_MODE, PRINTER_DEVICE, CF_TUNNEL_TOKEN, LOG_LEVEL`.
> `MYSQL_DSN` içindeki user/pass, `MYSQL_USER/MYSQL_PASSWORD` ile aynı olmalı; host `mysql` (servis adı).

### 3.4 README.md
Dükkan sahibi için kurulumu yeni mimariye göre yaz: `.env` doldur → `docker compose up -d --build` → seed → Cloudflare hostname → Trendyol'a webhook URL.

---

## 4. E2E Test (doğrulama)
```powershell
# 1. Tüm stack (tünelsiz)
docker compose up -d --build
# 1b. Tünel de istenirse:
docker compose --profile tunnel up -d --build

# 2. Admin kullanıcısı (§3.1 sonrası)
docker compose exec print-relay /app/seed --username admin --password Admin123

# 3. Webhook testi (lokal)
powershell.exe -ExecutionPolicy Bypass -File .\test_webhook_status.ps1 -Status Created

# 4. Baskı çıktısı
Get-Content output.txt -Tail 30

# 5. API duman testi
#    POST /api/auth/login {admin/Admin123} → token
#    GET /api/orders (Bearer) → yeni sipariş listede
#    POST /api/orders/{id}/print → printer/status'ta success

# 6. Dış erişim: https://printer.<domain>  → login admin/Admin123
#    Trendyol panelinden test webhook → sipariş listede + output.txt'te baskı
```

---

## 5. Çıkış Kriteri
- [ ] `docker compose up -d --build` → mysql + print-relay sağlıklı ayağa kalkıyor
- [ ] `docker compose exec print-relay /app/seed ...` admin kullanıcı oluşturuyor, login çalışıyor
- [ ] Lokal webhook testi → `trendyol_orders` + `print_jobs` + `output.txt` baskı
- [ ] `--profile tunnel` ile cloudflared bağlı (log: "Registered tunnel connection")
- [ ] Dış URL'den webhook → baskı; web UI dışarıdan login oluyor
- [ ] `.env.example` (CF_TUNNEL_TOKEN dahil) ve `README.md` güncel

---

## 6. Dikkat Edilecekler
- `CF_TUNNEL_TOKEN` boşsa cloudflared crash eder → `profiles: ["tunnel"]` ile opsiyonel tut.
- mysql healthcheck olmadan da main.go 30×2s retry'lar; yine de healthcheck + `service_healthy` daha temiz.
- `docs/schema.sql` yalnızca **boş volume**'da çalışır; şema değişirse `docker compose down -v` gerekir (veri silinir).
- Trendyol'a **daima 200** dönülür — webhook handler'a dokunma.
- Gerçek yazıcıya geçiş: `.env`'de `TEST_MODE=false` + `PRINTER_DEVICE=/dev/usb/lp0`, compose'da `devices:` bloğunu aç.
- Go toolchain geliştirme makinesinde **yok**; tüm Go derlemesi Docker builder stage'inde doğrulanır.
- `web/dist` ve `node_modules` git'te yok; imaj `web-builder` stage'inde üretir.
