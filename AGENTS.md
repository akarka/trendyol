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

cmd/app/main.go                # entry point
config/config.go               # env var yükleme
internal/server/               # chi router + webhook/middleware [AKTİF]
internal/db/                   # MySQL bağlantısı + query'ler [AKTİF]
internal/auth/                 # JWT [HENÜZ YOK — S03]
internal/printer/txt_printer.go        # AKTİF
internal/printer/escpos_printer.go     # STUB
internal/printer/digital_printer.go   # UNUSED
internal/parser/payload_parser.go      # DBRow → Order
internal/alerter/system_alerter.go     # log sarmalayıcı
web/                           # React + Tailwind [HENÜZ YOK]
```

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
| S05 | React Pages | 🔲 Bekliyor |
| S06 | Cloudflare Tunnel + Docker Compose | 🔲 Bekliyor |

---

## Görev Başlangıç Protokolü

1. `docs/PLAN.md` → ilerleme tablosunu oku
2. `docs/sessions/` → aktif session dosyasını oku
3. Session dosyasındaki **Giriş Noktası** ve **Bu Session Yapılacaklar** bölümünden başla
4. Session bitince: session dosyasındaki **Durum** → `✅ Tamamlandı` yap, `docs/PLAN.md` tablosunu güncelle
