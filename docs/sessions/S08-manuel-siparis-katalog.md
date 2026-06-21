# S08 — Manuel Sipariş + Ürün Kataloğu + Bitmap Baskı

**Durum:** ✅ Tamamlandı (Docker'da uçtan uca doğrulandı)
**Kapsam:** Elle sipariş girişi → DB → laser printer'a A4 sticker baskı; Zeytuni ürün kataloğu (normalize şema + import/export).

## Hedef

Henüz hiç test edilmemiş webhook akışı yerine, Trendyol panelinden gelen siparişi **elle** girip (ürün dropdown + müşteri adı) DB'ye yazmak ve etiketi basmak. Thermal printer alınana kadar **laser**'a A4 sticker'a basılır; sonra aynı bitmap thermal'a (ESC/POS raster) yönlendirilir.

## Kararlar (2026-06-20)

- **Baskı = bitmap.** Hem laser hem thermal aynı görüntüyü basar; output.txt'deki Türkçe encoding sorunu (gerçek font ile) ortadan kalkar.
- **Bitmap tarayıcıda (canvas) üretilir.** Sebep: server-side Türkçe PNG render'ı `golang.org/x/image` (yeni bağımlılık + go.sum) gerektirir; yerel Go toolchain olmadığı için derlenemeyen bağımlılık riski alınmadı. Backend stdlib-only kaldı. Thermal'a geçişte canvas PNG'si server'a POST edilip ESC/POS raster'a çevrilecek (gelecek session).
- **Ürün kaynağı:** normalize `products` tablosu + Zeytuni CSV import. Excel round-trip ile işletme bakımı.
- **Excel formatı:** `.xlsx` (stdlib `archive/zip`+`encoding/xml`, harici lib yok).
- **Veri blocker'ları:** tüm 67 ürün içe alınır; gramaj eksik/0 ve çift SKU `needs_fix=1` ile işaretlenir, işletme Excel'de düzeltip re-import eder.

## Yapılanlar

**Şema (`docs/schema.sql` + runtime migrate)**
- `categories`, `brands`, `products` (normalize, FK'li). `products`: sku PK, barcode UNIQUE, name (iç ad), marketplace_name, category_id/brand_id FK, net_weight+unit, price (Fiziki Mağaza Fiyatı), vat_rate (KDV % — türetilemez, saklanır), is_active, **needs_fix**, description.
- `internal/db/migrate.go` `EnsureCatalogTables` — `mysql_data` volume kalıcı olduğundan init script mevcut DB'de çalışmaz; app/import/export açılışta tabloları garanti eder.

**Araçlar**
- `cmd/import-products` — CSV (cp1254→utf8 bootstrap) **veya** .xlsx oku → fiyat/%/gramaj normalize → kategori/marka dedup → `products` idempotent `ON DUPLICATE KEY UPDATE` (SKU). Çift SKU/barkod → suffix `-DUP2` + needs_fix. Gramaj ürün adından parse; 0/yok → NULL + needs_fix.
- `cmd/export-products` — `products` → işletmenin düzelteceği `.xlsx`. Başlıklar import ile uyumlu (round-trip). Ek kolonlar: `Net Gramaj`, `Birim`, `Needs Fix`.
- `internal/xlsxlite` — stdlib-only xlsx oku (sharedStrings+inlineStr) / yaz (inlineStr).

**Backend**
- `GET /api/products[?active=1]` — dropdown kaynağı.
- `POST /api/orders/manual` — `{customer_name, lines:[{sku,quantity}]}` → `parser.Order` kur (payload Trendyol şemasıyla aynı → reprint aynı yoldan çalışır) → `trendyol_orders` insert (order_id `MAN-<unixnano>`, status `Created`) → print_job + kuyruk.

**Frontend**
- `web/src/lib/label.ts` — output.txt düzeninde monokrom canvas bitmap; `printLabel(lines, layout, cell)` A4 sayfasında seçilen hücreye, etiket ebatında konumlandırıp `window.print()` çağırır (laser).
- `web/src/pages/ManualOrderPage.tsx` — müşteri adı + ürün dropdown satırları + adet + canlı etiket önizleme + **hücre seçici** + "Kaydet ve Yazdır". Baskı sonrası sıradaki sticker hücresine otomatik ilerler. Route `/manual`, nav'a eklendi.
- `web/src/api/products.ts`, `orders.ts` (createManualOrder).

**A4 sticker yerleşimi (ayarlanabilir/saklanan)**
- `web/src/lib/labelLayout.ts` — `LabelLayout` (sayfa ebadı, sütun/satır, etiket en/boy, kenar boşlukları, ara boşluklar, hücre iç padding) + hücre konum hesabı. Varsayılan: A4 tam dolduran 3×8.
- `settings` tablosunda **`label_layout`** anahtarı altında JSON olarak saklanır (yeni backend gerekmedi; mevcut GET/PUT settings kullanılır).
- `web/src/components/SheetPreview.tsx` — ölçekli A4 + hücre ızgarası; SettingsPage'de önizleme, ManualOrderPage'de tıklanabilir hücre seçici.
- `web/src/pages/SettingsPage.tsx` — yapısal "Etiket Yerleşimi" editörü (canlı önizleme + varsayılana dön); ham `label_layout` anahtarı genel listeden gizlenir.

**Docker / build**
- `Dockerfile` — `import-products`, `export-products` binary'leri + bootstrap CSV final image'e kopyalanır.
- `.dockerignore` eklendi (.git, node_modules, görseller).
- **`web/package-lock.json` yeniden üretildi** — mevcut lockfile `package.json` ile uyumsuzdu (picomatch), web build zaten kırıktı; düzeltildi.

## Doğrulama (Docker)

- `go build ./...` ✅ (sıfır yeni bağımlılık)
- Full image build ✅; web-builder (vite) ✅
- import → MySQL: **67 ürün, 14 kategori, 16 marka, 15 aktif, 13 needs_fix** (1 çift SKU + 11 gramaj + Nar Ekşisi). İdempotent (tekrar = 67). UTF-8 doğru (`Foça` = `46 6F C3 A7 61`).
- export → `.xlsx` (Excel açtı) → re-import ✅ round-trip.

## Çalıştırma

```bash
# katalog import (bootstrap CSV image'de gömülü)
docker compose run --rm print-relay ./import-products
# işletme için Excel çıktısı (host'a al)
docker compose run --rm -v "$PWD:/out" print-relay ./export-products --out /out/urunler.xlsx
# işletme düzeltir → re-import
docker compose run --rm -v "$PWD:/out" print-relay ./import-products --file /out/urunler.xlsx
```

Manuel sipariş: SPA → **Manuel Sipariş** → ürün seç + müşteri → Kaydet ve Yazdır → tarayıcı yazdırma diyaloğu (laser seç).

## Açık / Sonraki

- **Çift SKU `ZYT-TER-012`** — işletmeden doğru SKU (muhtemelen `-013`); şu an `-DUP2`.
- **Eksik gramaj** (turşu ×7, sirke ×4, Nar Ekşisi) — işletme Excel'de doldurur.
- **Baskıda ad:** şu an iç ad (`products.name`). `Ty Ürün Adı` istenirse handler'da tek satır.
- **Thermal geçişi:** canvas PNG → `POST` → ESC/POS raster (`internal/printer/escpos_printer.go` hâlâ stub).
- **Webhook akışı** hâlâ canlı test edilmedi (manuel akış ondan bağımsız çalışır).
