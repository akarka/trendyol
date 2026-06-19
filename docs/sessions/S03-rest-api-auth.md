# S03 — REST API + JWT Auth

**Durum:** 🔲 Bekliyor
**Bağımlılık:** S02 tamamlanmış olmalı
**Sonraki:** S04-react-setup.md

---

## Bu Session Yapılacaklar

### 1. JWT Auth (`internal/auth/`)

```
internal/auth/
  jwt.go     # GenerateToken(userID, role) string, ValidateToken(token) (*Claims, error)
  middleware.go  # chi middleware: Authorization: Bearer <token> zorunlu, role kontrol
```

`Claims` struct:
```go
type Claims struct {
  UserID   int    `json:"user_id"`
  Username string `json:"username"`
  Role     string `json:"role"`
  jwt.RegisteredClaims
}
```

Token süresi: 24 saat. Refresh yok (şimdilik).

### 2. Auth endpoint (`internal/server/auth.go`)

```
POST /api/auth/login
  Body: { "username": "...", "password": "..." }
  → users tablosunda bak → bcrypt karşılaştır → JWT dön
  Response: { "token": "...", "role": "admin" }
```

### 3. `internal/db/queries.go` genişlet

```go
GetUserByUsername(username string) (*User, error)
GetOrders(limit, offset int, status string) ([]Order, error)
GetOrderByID(orderID string) (*Order, error)
GetPrintJobs(limit int) ([]PrintJob, error)
GetSettings() (map[string]string, error)
UpsertSetting(key, value string) error
```

### 4. API endpoint'leri (`internal/server/api.go`)

Tüm `/api/*` route'ları JWT middleware arkasında.

```
GET  /api/orders?limit=50&offset=0&status=Created
GET  /api/orders/:order_id
POST /api/orders/:order_id/print   → print_jobs'a yeni kayıt + printer channel'a gönder
GET  /api/printer/status           → son 20 print_job + TEST_MODE bilgisi
GET  /api/logs                     → son 100 print_job (tüm durumlar)
GET  /api/settings                 → tüm settings
PUT  /api/settings/:key            → Body: { "value": "..." }
```

### 5. İlk admin kullanıcısı

`docs/schema.sql`'e ekle:
```sql
-- Şifreyi seed sırasında değiştir
INSERT IGNORE INTO users (username, password_hash, role)
VALUES ('admin', '$2a$10$PLACEHOLDER', 'admin');
```

`cmd/seed/main.go` oluştur: `go run ./cmd/seed --username admin --password <pass>` → bcrypt hash üretir, DB'ye yazar.

### 6. Route yapısı (`internal/server/server.go`)

```
/webhook/trendyol        → BasicAuth
/api/auth/login          → public
/api/*                   → JWTMiddleware → handler'lar
/                        → static SPA (S04'te eklenecek, şimdi placeholder)
```

---

## Giriş Noktası

`internal/auth/jwt.go` → `internal/auth/middleware.go` → `internal/server/auth.go` → `internal/db/queries.go` → `internal/server/api.go` → `internal/server/server.go` route güncellemesi

---

## Çıkış Kriteri

- [ ] `POST /api/auth/login` geçerli credential ile JWT dönüyor
- [ ] JWT olmadan `/api/*` → 401
- [ ] `GET /api/orders` sipariş listesi dönüyor
- [ ] `POST /api/orders/:id/print` baskıyı tetikliyor
- [ ] `GET /api/printer/status` son job'ları dönüyor
- [ ] `GET/PUT /api/settings` çalışıyor
- [ ] `go run ./cmd/seed` admin kullanıcısı oluşturuyor
- [ ] `go build ./...` hatasız

---

## Dikkat Edilecekler

- JWT secret `.env`'den gelir: `JWT_SECRET=...` — hard-code etme
- Tüm API response'ları `application/json`
- Hata response formatı tutarlı: `{ "error": "mesaj" }`
- `POST /api/orders/:id/print` idempotent değil — her çağrı yeni print_job açar
