# Zeytuni Ürün Verisi — İnceleme Raporu (S07)

**Kaynak:** `Zeytuni_Ops/Ürün_Listesi.csv` (cp1254 → UTF-8 decode edilerek okundu)
**Kapsam:** Discovery / karar. Kod ve şema değiştirilmedi.

## 0. Özet

- 67 ürün, 14 kategori.
- Durum: **15 AKTIF, 52 PASIF**.
- Veri Trendyol'a doğrudan basılacak kalitede değil. 5 bloklayıcı sorun var (encoding, çift SKU, sahte barkod, eksik gramaj, fiyat formatı).
- Printer modülü için asıl değer: **webhook barkodu → `products` join → baskıda ürün adı/kategori zenginleştirme.** Ürün tablosu bu amaçla minimal tutulmalı; ERP'leşmeden.

## 1. Kategori dağılımı

| Kategori | Adet | AKTIF |
|----------|-----:|------:|
| Reçel | 15 | 4 |
| Peynir | 8 | 2 |
| Zeytin | 7 | 2 |
| Turşu | 7 | 0 |
| Soğuk | 6 | 2 |
| Yoğurt | 5 | 1 |
| Tereyağı | 5 | 2 |
| Sirke | 4 | 0 |
| Marmelat | 3 | 2 |
| Salça | 3 | 0 |
| Kuruyemiş | 3 | 0 |
| Sos | 1 | 0 |
| Pekmez | 1 | 0 |
| Bal | 1 | 0 |

## 2. Veri kalitesi bulguları

### 2.1 Encoding (bloklayıcı)
CSV cp1254 ve üstüne çift bozulma var:
- `Ürün`→`�r�n`, `₺`→`?` (kalıcı kayıp, byte yok).
- Dotless ı kaybı: kaynakta zaten `Yapımı`→`Yapimi`, `Fiyatı`→`Fiyati` düz `i` olmuş. cp1254 decode bunu kurtarmıyor.

**Öneri:** CSV'yi kaynak alma. `Ürün_Listesi.xlsx`'i UTF-8 + `,` ayraçla yeniden export et; ₺ ve ı/İ karakterlerini xlsx'te doğrula.

### 2.2 Çift SKU (bloklayıcı)
`ZYT-TER-012` iki satırda:
- satır 28 → Çanakkale **500 g** (Köy Tereyağı 500 g), AKTIF
- satır 32 → Çanakkale **1 kg**, AKTIF

TER dizisi 009,010,011,012,**012** — 013 atlanmış. Muhtemelen 500 g satırı `ZYT-TER-013` olmalı. SKU unique olacaksa biri düzeltilmeli. **Karar gerek.**

### 2.3 Sahte barkod (strateji kararı)
66/67 ürün `299000000xxxx` iç barkod (gerçek EAN değil). Tek gerçek EAN: `8694415150010` (Foça yoğurt, satır 15).
Trendyol gerçek/onaylı barkod ister. **Karar gerek** (bkz. §5).

### 2.4 Eksik gramaj
- Turşu 7 ürün → `0 g` placeholder (satır 49-55).
- Sirke 4 ürün → `0 L` placeholder (satır 56-59).
- Sos (Nar Ekşisi, satır 63) → gramaj/hacim yok.
- Pekmez (satır 64) → tutarsızlık: `Ürün`="Üzüm Pekmezi **850 g**" ama `Ty Ürün Adı`="Üzüm Pekmezi **0 g**".

Gerçek gramajlar eksik. **Karar gerek** (dükkandan gerçek değerler).

### 2.5 Fiyat formatı
- `KDV Dahil Satış Fiyatı`: metin + `?` (eski ₺) + binlik ayraç → `" ?1,049.18 "`. Sayıya normalize edilmeli.
- `Komisyon Oranı`: `23.75%` metin. `KDV`: `1%` / `10%` metin.
- Temiz alanlar: `Fiziki Mağaza Fiyatı` (düz integer) ve `Komisyon Sonrası` kullanılabilir; KDV dahil fiyat bunlardan türetilebilir → saklamak şart değil.

### 2.6 Diğer (yeni gözlemler)
- **SKU numaralandırma tutarsız:** çoğu kategori bazında sıfırdan (`001`), ama Sos/Pekmez/Bal global sayaç (`017,018,019`), Sirke `013`'ten başlıyor. Şema açısından önemli değil (SKU sadece unique anahtar) ama veri hijyeni notu.
- **Marka çeşitli:** sadece Zeytuni değil — Kıryatan, Teksen, Orfarm, Atalay, Foça, Mezre, Petek vb. (kendi üretim vs. tedarik ayrımı var). `brand` kolonu anlamlı.
- **`Frictions` kolonu tamamen boş** → tablo dışı bırak.
- **`Ürün` vs `Ty Ürün Adı`:** ikisi farklı (iç ad vs. pazaryeri başlığı). Baskıda hangisi gösterilecek? Muhtemelen kısa iç ad (`Ürün`). Saklamaya değer iki ayrı alan.

## 3. `products` şema önerisi (taslak — UYGULANMADI)

`docs/schema.sql` stiline hizalı (MySQL 8, utf8mb4). Minimal; komisyon/KDV türevleri saklanmıyor.

```sql
CREATE TABLE IF NOT EXISTS products (
  sku             VARCHAR(32)  PRIMARY KEY,
  barcode         VARCHAR(32)  NOT NULL,
  name            VARCHAR(255) NOT NULL,   -- iç ad (Ürün kolonu)
  marketplace_name VARCHAR(255),           -- Ty Ürün Adı (baskı/eşleşme yedek)
  category        VARCHAR(64)  NOT NULL,
  brand           VARCHAR(64),
  price           DECIMAL(10,2) NOT NULL,  -- Fiziki Mağaza Fiyatı
  is_active       BOOLEAN      NOT NULL DEFAULT 0,
  created_at      DATETIME     DEFAULT CURRENT_TIMESTAMP,
  updated_at      DATETIME     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uq_barcode (barcode)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**Saklanmayanlar (gerekçe):** `Komisyon Oranı/Tutarı/Sonrası`, `KDV Dahil Fiyat` → hepsi `price` + sabit oranlardan hesaplanabilir, değiştiğinde tek yerden. `Frictions` → boş. Görsel → şimdilik kapsam dışı.

### Parser join gerekçesi
Webhook siparişi geldiğinde `products` join → baskı fişine ürün adı/kategori/iç ad eklenir.
Bu, Trendyol payload'ının verdiğinden daha zengin/tutarlı baskı sağlar. **Tablonun asıl operasyonel faydası bu.**

**Join anahtarı kararı:** şimdilik **SKU** üzerinden eşleştir; Trendyol uygun endpoint sağladığında barkoda geçilir.
⚠️ Pratik not: mevcut `internal/parser/payload_parser.go` `OrderLine` yalnızca `Barcode` taşıyor, `SKU` yok. SKU join'i için ya
(a) webhook payload'ında SKU alanı bulunmalı, ya da (b) `products.barcode` (iç `299...` barkod) Trendyol sipariş barkoduyla
birebir aynıysa geçici köprü olarak barkod kullanılır. Bu, Trendyol endpoint'i gelene kadar implementasyon session'ında netleşir.

## 4. Import planı (taslak — UYGULANMADI)

- Kaynak: temiz UTF-8 export (xlsx → CSV/JSON).
- `cmd/seed` benzeri ayrı binary: `cmd/import-products`.
- Idempotent: `INSERT ... ON DUPLICATE KEY UPDATE` — **SKU üzerinden** (PK), tekrar çalıştırılınca günceller.
- Fiyat/oran metinleri import sırasında normalize edilir (`%`, `₺`, binlik ayraç temizle).
- Görseller (`operasyon/urunler/Gorseller/`) kapsam dışı.
- Bu, ayrı bir implementasyon session'ında yapılır (S07 kod yazmaz).

## 5. Kararlar

Kullanıcı kararları (2026-06-20):

1. **Join anahtarı** → ✅ **SKU** üzerinden eşleştir; Trendyol endpoint geldiğinde barkoda geçilecek. (bkz. §3 pratik not)
2. **Import kapsamı** → ✅ **Tüm 67 ürün** (PASIF dahil katalog); `is_active` flag ile ayrım.
3. **Çift SKU `ZYT-TER-012`** → ⏳ **Dükkana sorulacak** (doğru SKU teyit edilecek). Import öncesi çözülmeli, aksi halde PK çakışır.

Hâlâ açık (implementasyon session'ı öncesi netleşmeli):

4. **Eksik gramaj** — Turşu (0 g ×7), Sirke (0 L ×4), Nar Ekşisi, Pekmez (850 mı 0 mı?) gerçek değerleri dükkandan.
5. **Baskıda ad** — fiş için `Ürün` (iç ad) mı `Ty Ürün Adı` mı? (Şema ikisini de saklıyor; sadece görüntüleme tercihi.)
</content>
</invoke>
