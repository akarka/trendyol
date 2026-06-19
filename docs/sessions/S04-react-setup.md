# S04 — React Setup + Auth UI

**Durum:** 🔲 Bekliyor
**Bağımlılık:** S03 tamamlanmış olmalı
**Sonraki:** S05-react-pages.md

---

## Bu Session Yapılacaklar

### 1. Vite + React + Tailwind kurulumu (`web/`)

```bash
cd web
npm create vite@latest . -- --template react-ts
npm install
npm install -D tailwindcss @tailwindcss/forms autoprefixer postcss
npm install react-router-dom @tanstack/react-query axios
npx tailwindcss init -p
```

`tailwind.config.js`:
```js
content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"]
plugins: [require('@tailwindcss/forms')]
```

### 2. Proje yapısı (`web/src/`)

```
web/src/
  api/
    client.ts       # axios instance, baseURL, Bearer token interceptor
    orders.ts       # getOrders, getOrder, printOrder
    auth.ts         # login
    settings.ts     # getSettings, updateSetting
    printer.ts      # getStatus, getLogs
  context/
    AuthContext.tsx  # token, role, login(), logout()
  router/
    index.tsx        # React Router, ProtectedRoute wrapper
  components/
    Layout.tsx       # sidebar + mobile nav
    Spinner.tsx
    Badge.tsx        # sipariş durumu renk kodlaması
  pages/
    LoginPage.tsx
    OrdersPage.tsx   # S05'te doldurulacak — şimdi placeholder
    PrinterPage.tsx  # S05'te doldurulacak — şimdi placeholder
    SettingsPage.tsx # S05'te doldurulacak — şimdi placeholder
```

### 3. API client (`web/src/api/client.ts`)

```ts
// baseURL: production'da '' (same-origin), dev'de http://localhost:8080
const client = axios.create({ baseURL: import.meta.env.VITE_API_URL ?? '' })
client.interceptors.request.use(cfg => {
  const token = localStorage.getItem('token')
  if (token) cfg.headers.Authorization = `Bearer ${token}`
  return cfg
})
// 401 gelirse: localStorage temizle, /login'e yönlendir
```

### 4. AuthContext + Login sayfası

`LoginPage.tsx`: email/password form → `POST /api/auth/login` → token localStorage'a → `/orders`'a redirect.
Mobil öncelikli: full-screen centered form, Tailwind.

### 5. Layout (`components/Layout.tsx`)

Mobil: alt nav bar (sipariş ikonu, yazıcı ikonu, ayarlar ikonu)
Desktop: sol sidebar (aynı linkler + başlık)
Responsive: `sm:hidden` / `hidden sm:flex` pattern

### 6. Vite proxy (geliştirme için)

`vite.config.ts`:
```ts
server: {
  proxy: {
    '/api': 'http://localhost:8080',
    '/webhook': 'http://localhost:8080'
  }
}
```

### 7. Go embed güncellemesi

`cmd/app/main.go`:
```go
//go:embed web/dist
var staticFiles embed.FS

// server.New() içinde:
// r.Handle("/*", http.FileServer(http.FS(staticFiles)))
```

`Dockerfile`'a build adımı ekle:
```dockerfile
FROM node:20-alpine AS web-builder
WORKDIR /web
COPY web/package*.json ./
RUN npm ci
COPY web/ .
RUN npm run build

FROM golang:1.22-alpine AS builder
COPY --from=web-builder /web/dist ./web/dist
# ... go build
```

---

## Giriş Noktası

`web/` klasörü oluştur → npm setup → `src/api/client.ts` → `src/context/AuthContext.tsx` → `src/pages/LoginPage.tsx` → `src/components/Layout.tsx` → `src/router/index.tsx`

---

## Çıkış Kriteri

- [ ] `npm run dev` çalışıyor, login sayfası açılıyor
- [ ] Login → JWT alınıyor, localStorage'a yazılıyor
- [ ] Layout: mobilde alt nav, desktop'ta sidebar
- [ ] 401 → otomatik logout + login sayfasına yönlendirme
- [ ] `npm run build` → `web/dist/` oluşuyor
- [ ] `go build ./...` embed ile birlikte hatasız
- [ ] `localhost:8080/` → React SPA açılıyor

---

## Dikkat Edilecekler

- `web/dist/` `.gitignore`'a ekle
- `node_modules/` `.gitignore`'da olmalı
- Tailwind: dark mode şimdilik yok, gereksiz karmaşıklık
- TanStack Query: `QueryClient` ile wrap et, `staleTime: 30_000` default
