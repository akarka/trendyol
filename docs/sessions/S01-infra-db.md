# S01 — Infra + DB

**Durum:** 🔲 Bekliyor
**Bağımlılık:** Yok (ilk session)
**Sonraki:** S02-go-server-webhook.md

---

## Bu Session Yapılacaklar

### 1. Eski Supabase kodunu temizle
- `internal/listener/` klasörünü sil (Supabase WebSocket, Phoenix protocol)
- `go.mod` / `go.sum`'dan `github.com/gorilla/websocket` kaldır
- `cmd/app/main.go`'dan listener import ve `StartRealtimeSubscription` çağrısını kaldır
- `config/config.go`'dan `SupabaseURL`, `SupabaseAnonKey` alanlarını kaldır; `MYSQL_DSN`, `JWTSecret` ekle
- `.env.example`'dan Supabase değişkenlerini kaldır; MySQL + JWT değişkenlerini ekle

### 2. Go bağımlılıklarını ekle
```bash
go get github.com/go-chi/chi/v5
go get github.com/jmoiron/sqlx
go get github.com/go-sql-driver/mysql
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto
go mod tidy
```

### 3. MySQL Docker setup
`docker-compose.yml` yeniden yaz:
```yaml
services:
  print-relay:
    build: .
    depends_on: [mysql]
    env_file: [.env]
    volumes:
      - ./output.txt:/app/output.txt
    ports:
      - "8080:8080"

  mysql:
    image: mysql:8
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: trendyol
      MYSQL_USER: ${MYSQL_USER}
      MYSQL_PASSWORD: ${MYSQL_PASSWORD}
    volumes:
      - mysql_data:/var/lib/mysql
      - ./docs/schema.sql:/docker-entrypoint-initdb.d/schema.sql

volumes:
  mysql_data:
```

### 4. DB şemasını dosyaya çıkar
`docs/schema.sql` oluştur — `docs/PLAN.md`'deki hedef şemayı yaz.

### 5. `internal/db/` paketi oluştur (sadece bağlantı)
```
internal/db/db.go   # sqlx.Connect, ping, DB nesnesini döner
```

### 6. `main.go`'yu minimal hale getir
Sadece: config yükle → DB bağlan → (S02'de server başlatılacak).

---

## Giriş Noktası

`config/config.go` → temizlik → `go.mod` → bağımlılıklar → `docker-compose.yml` → `docs/schema.sql` → `internal/db/db.go` → `cmd/app/main.go`

---

## Çıkış Kriteri

- [ ] `go build ./...` hatasız geçiyor
- [ ] `docker-compose up mysql` → container ayağa kalkıyor
- [ ] `docs/schema.sql` çalıştırılınca 4 tablo oluşuyor
- [ ] `internal/listener/` klasörü yok
- [ ] `gorilla/websocket` `go.mod`'da yok

---

## Dikkat Edilecekler

- `internal/parser/payload_parser.go` **dokunma** — `Order` struct değişmemeli
- `internal/printer/` **dokunma** — printer paketleri aynı kalacak
- `main.go`'daki `printer.PrintToTXT` çağrısını şimdilik stub olarak bırak; S02'de webhook handler'a taşınacak
- MySQL DSN formatı: `user:pass@tcp(mysql:3306)/trendyol?parseTime=true&charset=utf8mb4`
