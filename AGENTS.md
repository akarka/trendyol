# AGENTS.md — Agentic Handoff

Bu dosya codebase'i devralan AI agent'lar içindir.

---

## Proje Kimliği

**Rol:** Dükkan yönetim sisteminin termal yazıcı modülü + admin web arayüzü.
**Repo:** `github.com/akarka/trendyol` — Go 1.22, React 18, MySQL 8, Docker
**Mimari özeti:** Trendyol → Cloudflare Tunnel → Go HTTP Server → MySQL + Printer

---

## Aktif Session

`docs/sessions/` klasörünü oku. `status: active` olan session'ı bul ve oradan devam et.

---

## Dosya Haritası

```
CLAUDE.md                      # mimari + davranış kuralları
AGENTS.md                      # bu dosya
docs/PLAN.md                   # tüm implementasyon planı + ilerleme tablosu
docs/sessions/S0X-*.md         # session handoff'ları

cmd/app/main.go                # entry point: config→DB(retry)→printer goroutine→server
cmd/seed/main.go               # bcrypt admin kullanıcı: --username --password [--role]
config/config.go               # env var yükleme (MySQL+JWT+webhook+printer)
internal/server/server.go      # chi router, route'lar, spaHandler (SPA embed)
internal/server/webhook.go     # POST /webhook/trendyol (BasicAuth, daima 200)
internal/server/auth.go        # POST /api/auth/login (bcrypt → JWT)
internal/server/api.go         # /api/orders, reprint, printer/status, logs, settings
internal/server/middleware.go  # webhook BasicAuth
internal/db/db.go              # sqlx bağlantı
internal/db/queries.go         # orders/print_jobs/settings/users query'leri
internal/auth/jwt.go           # HS256 token üret/doğrula (24s)
internal/auth/middleware.go    # /api/* için Bearer JWT zorunlu
internal/printer/txt_printer.go        # AKTİF (TEST_MODE)
internal/printer/escpos_printer.go     # ESC/POS
internal/printer/digital_printer.go    # UNUSED
internal/parser/payload_parser.go      # DBRow → Order (dokunulmadı)
internal/alerter/system_alerter.go     # log sarmalayıcı
web/embed.go                   # //go:embed all:dist → web.Dist()
web/src/                       # React 18 + TS + Tailwind v3 (api/context/router/components/pages)
web/dist/                      # build çıktısı (git'te yok; Docker build üretir)
```

> `internal/listener/` (Supabase WebSocket) silindi; Phoenix protokolü tamamen kaldırıldı.

---

## Kritik Kısıtlar

1. `internal/listener/` Supabase'e bağlı — yeni mimaride silinecek, doğrudan HTTP webhook alıyor
2. `internal/parser/payload_parser.go` değişmiyor — Trendyol payload yapısı aynı
3. `internal/printer/` değişmiyor — printer abstraction aynı
4. Phoenix/WebSocket protokolü tamamen kalkıyor
5. JWT `role` claim zorunlu — middleware her zaman kontrol eder, şimdilik sadece `admin` geçer
6. Trendyol'a her zaman HTTP 200 dön (retry bastırma) — webhook handler değiştirme
7. React `dist/` → Go binary'ye `embed.FS` ile gömülü — ayrı statik sunucu yok

---

## Tamamlanan / Bekleyen

| # | Session | Durum |
|---|---------|-------|
| S01 | Infra + DB | ✅ Tamamlandı |
| S02 | Go HTTP Server + Webhook | ✅ Tamamlandı |
| S03 | REST API + JWT Auth | ✅ Tamamlandı |
| S04 | React Setup + Auth UI | ✅ Tamamlandı |
| S05 | React Pages | ✅ Tamamlandı |
| S06 | Cloudflare Tunnel + Docker Compose | 🔲 Bekliyor |

---

## Görev Başlangıç Protokolü

1. `docs/PLAN.md` → ilerleme tablosunu oku
2. `docs/sessions/` → aktif session dosyasını oku
3. Session dosyasındaki **Giriş Noktası** ve **Bu Session Yapılacaklar** bölümünden başla
4. Session bitince: session dosyasındaki **Durum** → `✅ Tamamlandı` yap, `docs/PLAN.md` tablosunu güncelle
