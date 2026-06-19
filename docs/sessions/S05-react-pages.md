# S05 — React Pages

**Durum:** 🔲 Bekliyor
**Bağımlılık:** S04 tamamlanmış olmalı
**Sonraki:** S06-tunnel-docker.md

---

## Bu Session Yapılacaklar

### 1. OrdersPage (`src/pages/OrdersPage.tsx`)

**Özellikler:**
- Sipariş listesi; `GET /api/orders?status=&limit=50&offset=0`
- Filtre: PackageStatus (Created / Cancelled / Delivered / UnSupplied) — tab veya dropdown
- Her satırda: sipariş no, müşteri adı, durum badge'i, tarih, "Yazdır" butonu
- Satıra tıklayınca `OrderDetailModal` açılır (ayrı sayfa değil, modal)
- Sayfalama: Load More veya basit prev/next

**Badge renkleri (Tailwind):**
```
Created    → green
Cancelled  → red
Delivered  → blue
UnSupplied → yellow
```

**Manuel baskı:**
- "Yazdır" → `POST /api/orders/:id/print` → toast bildirim (success/error)
- Loading state butonda göster

### 2. OrderDetailModal (`src/components/OrderDetailModal.tsx`)

```
Sipariş No: ...    Durum: [badge]
Müşteri: ...       Tarih: ...
Kargo: ...

Ürünler:
  2x Ürün Adı    45.00 TL
  1x Başka Ürün  22.50 TL
  ─────────────────────
  Toplam:        112.50 TL

Adres: ...

[Kapat]  [Yeniden Yazdır]
```

### 3. PrinterPage (`src/pages/PrinterPage.tsx`)

- `GET /api/printer/status` → yazıcı modu (TEST / ESC/POS), device path
- Son 20 print_job tablosu: sipariş no, durum (queued/success/failed), hata mesajı, zaman
- Hata satırları kırmızı highlight
- Auto-refresh: her 10 saniyede TanStack Query refetch (`refetchInterval: 10_000`)

### 4. SettingsPage (`src/pages/SettingsPage.tsx`)

Basit key-value form. Her satır edit edilebilir:
```
PRINTER_DEVICE    [/dev/usb/lp0        ] [Kaydet]
TEST_MODE         [true                ] [Kaydet]
WEBHOOK_USERNAME  [***                 ] [Kaydet]
```
`PUT /api/settings/:key` body: `{ "value": "..." }`
Toast: "Kaydedildi" / "Hata"

### 5. Toast bileşeni (`src/components/Toast.tsx`)

Basit: sağ alt köşe, 3 saniye sonra kaybolur.
```tsx
type ToastType = 'success' | 'error'
```
Context veya basit state — overkill etme.

### 6. Mobil optimizasyon

- Sipariş listesi: kart görünümü (tablo değil) mobilde (`sm:hidden` tablo, `sm:grid` kart)
- Tüm butonlar minimum `h-10` touch target
- Modal: mobilde full-screen (`sm:max-w-lg sm:rounded`)

---

## Giriş Noktası

`src/pages/OrdersPage.tsx` → `src/components/OrderDetailModal.tsx` → `src/pages/PrinterPage.tsx` → `src/pages/SettingsPage.tsx` → `src/components/Toast.tsx`

---

## Çıkış Kriteri

- [ ] Sipariş listesi yükleniyor, filtre çalışıyor
- [ ] "Yazdır" → toast success + print_jobs'a kayıt
- [ ] Detail modal açılıp kapanıyor, tüm alanlar görünüyor
- [ ] Printer sayfası 10s'de bir güncelleniyor
- [ ] Settings kaydediliyor
- [ ] Mobil görünüm: alt nav + kart layout çalışıyor
- [ ] `npm run build` hatasız

---

## Dikkat Edilecekler

- React Query key'leri tutarlı: `['orders', filters]`, `['printer-status']`, `['settings']`
- Modal içinde scroll: uzun sipariş/adres için `overflow-y-auto max-h-[80vh]`
- Hata state'leri: API down → "Bağlantı hatası" mesajı, boş liste → "Henüz sipariş yok"
- Gereksiz abstraction ekleme — 5 sayfa için büyük state management (Redux vb.) yok
