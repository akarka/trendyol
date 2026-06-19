# S06 — Cloudflare Tunnel + Docker Compose

**Durum:** 🔲 Bekliyor
**Bağımlılık:** S05 tamamlanmış olmalı
**Sonraki:** Proje tamamlandı

---

## Bu Session Yapılacaklar

### 1. Cloudflare Tunnel kurulumu

**Ön koşul (kullanıcı yapacak, kodla değil):**
1. cloudflare.com → Zero Trust → Tunnels → "Create a tunnel"
2. Tunnel adı: `trendyol-printer`
3. Token al: `CF_TUNNEL_TOKEN=...`
4. Public hostname ekle: `webhook.senindomain.com` → `http://print-relay:8080`
5. Admin UI için ek hostname: `printer.senindomain.com` → `http://print-relay:8080`

**`docker-compose.yml` güncellemesi:**
```yaml
  cloudflared:
    image: cloudflare/cloudflared:latest
    command: tunnel --no-autoupdate run
    environment:
      - TUNNEL_TOKEN=${CF_TUNNEL_TOKEN}
    depends_on:
      - print-relay
    restart: unless-stopped
```

### 2. `docker-compose.yml` final hali

```yaml
version: "3.9"
services:
  print-relay:
    build: .
    restart: unless-stopped
    depends_on:
      mysql:
        condition: service_healthy
    env_file: [.env]
    volumes:
      - ./output.txt:/app/output.txt
      # devices: - "/dev/usb/lp0:/dev/usb/lp0"  # gerçek yazıcı için

  mysql:
    image: mysql:8
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: trendyol
      MYSQL_USER: ${MYSQL_USER}
      MYSQL_PASSWORD: ${MYSQL_PASSWORD}
    volumes:
      - mysql_data:/var/lib/mysql
      - ./docs/schema.sql:/docker-entrypoint-initdb.d/schema.sql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  cloudflared:
    image: cloudflare/cloudflared:latest
    command: tunnel --no-autoupdate run
    environment:
      - TUNNEL_TOKEN=${CF_TUNNEL_TOKEN}
    depends_on: [print-relay]
    restart: unless-stopped

volumes:
  mysql_data:
```

### 3. `.env.example` final hali

```env
# MySQL
MYSQL_DSN=trendyol_user:password@tcp(mysql:3306)/trendyol?parseTime=true&charset=utf8mb4
MYSQL_ROOT_PASSWORD=changeme
MYSQL_USER=trendyol_user
MYSQL_PASSWORD=changeme

# Auth
JWT_SECRET=changeme-at-least-32-chars

# Webhook
WEBHOOK_USERNAME=trendyol
WEBHOOK_PASSWORD=changeme

# Printer
TEST_MODE=true
PRINTER_DEVICE=output.txt

# Cloudflare
CF_TUNNEL_TOKEN=

# App
LOG_LEVEL=info
```

### 4. Dockerfile final

S04'te yapılan multi-stage build'i gözden geçir:
- Node build stage → `web/dist/`
- Go build stage → binary
- Alpine runtime stage → sadece binary + output.txt

### 5. E2E test

```powershell
# 1. Başlat
docker-compose up -d --build

# 2. Admin kullanıcısı oluştur
docker exec print-relay ./print-relay seed --username admin --password Admin123

# 3. Webhook testi
powershell.exe -ExecutionPolicy Bypass -File .\test_webhook_status.ps1

# 4. output.txt kontrolü
Get-Content output.txt -Tail 20

# 5. Web UI: https://printer.senindomain.com
#    Login: admin / Admin123
#    Sipariş listesinde yeni sipariş görünmeli
#    "Yazdır" → printer status sayfasında success görünmeli
```

### 6. `README.md` güncelle

Son kullanıcı (dükkan sahibi) için kurulum adımlarını yeni mimariye göre yeniden yaz.

---

## Giriş Noktası

`docker-compose.yml` → `.env.example` → `Dockerfile` → e2e test → `README.md`

---

## Çıkış Kriteri

- [ ] `docker-compose up -d --build` tüm servisler ayağa kalkıyor
- [ ] MySQL `service_healthy` beklenince `print-relay` başlıyor
- [ ] Cloudflare Tunnel bağlı (cloudflared log'unda "Registered tunnel connection")
- [ ] Dış URL'den webhook testi → `output.txt`'e baskı
- [ ] Web UI dışarıdan erişilebilir, login çalışıyor
- [ ] `README.md` güncel

---

## Dikkat Edilecekler

- `CF_TUNNEL_TOKEN` boşsa `cloudflared` crash eder — `docker-compose.yml`'de `profiles` ile opsiyonel yap veya token zorunluluğunu belgele
- MySQL `healthcheck` olmadan `print-relay` DB'ye bağlanamadan crash eder — `depends_on condition: service_healthy` zorunlu
- `docs/schema.sql` Docker init script olarak çalışır — sadece ilk başlatmada, volume varsa tekrar çalışmaz
- Gerçek yazıcı geçişi: `TEST_MODE` kaldır + `docker-compose.yml`'de `devices:` bloğunu aç
