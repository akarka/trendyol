# S07 — Zeytuni Ürün Verisi İnceleme + Karar

**Durum:** ✅ Tamamlandı (2026-06-20) → çıktı: `docs/zeytuni-urunler-rapor.md`
**Bağımlılık:** S06 (deploy) devam ediyor — bu session onu BLOKLAMAZ, paralel inceleme.
**Tip:** Discovery / review. **Kod yazılmaz**, sadece rapor + şema önerisi + import planı üretilir.
**Sonraki:** Karar onaylanırsa ayrı bir implementasyon session'ı.

> Kendi kendine yeterlidir. Davranış kuralları: `CLAUDE.md` (kısa/doğrudan, Türkçe, yorum yok).
> **Kapsam sınırı:** S06 Cloudflare/Docker deploy ana iş. Bu session ondan uzaklaşmaz; ürün tablosunu
> bu session'da İMPLEMENTE ETME. Çıktı = karar + plan.

---

## 0. Amaç

Dükkanın (Zeytuni) ürün listesi elde var (`Zeytuni_Ops/Ürün_Listesi.xlsx` / `.csv`). Bu veriyi printer
sisteminde bir `products` tablosuna taşımanın doğru yolunu belirlemek. Önce **operasyonel anlayış + veri
kalitesi**, sonra şema/import kararı.

---

## 1. Kaynaklar (oku)

- `Zeytuni_Ops/README.md`, `Zeytuni_Ops/AGENTS.md` — işin felsefesi. Özet: Türkiye'de 7-8 masalık küçük
  kahvaltıcı; perakende (Trendyol vb.) mevcut operasyona **sessiz uzantı**. ERP zihniyeti / aşırı mühendislik
  / ağır mimariden KAÇIN. "Ölçekten önce stabilite."
- `Zeytuni_Ops/operasyon/urunler/` — ürün notları + `Gorseller/` (3 png).
- `Zeytuni_Ops/Ürün_Listesi.xlsx` — **gerçek kaynak (source of truth)**.
- `Zeytuni_Ops/Ürün_Listesi.csv` — xlsx'ten export ama **encoding BOZUK** (cp1254→UTF-8 mojibake:
  `Ürün`→`�r�n`, `₺`→`?`). Bu CSV'yi parse etme; xlsx'i UTF-8 (`;`/`,` ayraç netleştirilerek) yeniden export et.
- Not: `Zeytuni_Ops/` ayrı bir git reposudur (nested). Orada değişiklik yaparsan AGENTS.md gereği
  her değişiklik için commit. **Bu session'da Zeytuni_Ops'a yazma** — yalnızca oku.

---

## 2. Verinin Yapısı (CSV'den gözlem)

15 kolon: `SKU, Barkod, Kategori, Ürün, Marka, Ty Ürün Adı, Fiziki Mağaza Fiyatı, Komisyon Oranı,
Komisyon Tutarı, Komisyon Sonrası, Aktif/Pasif, KDV, KDV Dahil Satış Fiyatı, Frictions, Marketplace Açıklama`.

~67 satır. Kategoriler: Soğuk, Zeytin, Yoğurt, Peynir, Tereyağı, Reçel, Marmelat, Turşu, Sirke, Salça,
Sos, Pekmez, Bal, Kuruyemiş.

---

## 3. Bilinen Veri Kalitesi Sorunları (doğrula + raporla)

1. **Encoding** — CSV mojibake; xlsx'ten temiz UTF-8 export gerekli.
2. **Çift SKU** — `ZYT-TER-012` iki satırda (Çanakkale 500g ve Çanakkale 1 kg). SKU unique olmalı → biri yanlış.
3. **Placeholder ağırlık** — bazı ürünlerde `0 g` / `0 L` (Turşu, Sirke, Sos, Pekmez). Gerçek gramaj eksik.
4. **Sahte barkod** — çoğu `299000000xxxx` iç barkod (gerçek EAN değil). Tek gerçek EAN: `8694415150010`
   (Foça yoğurt). Trendyol gerçek barkod ister → strateji kararı gerekir.
5. **Fiyat alanları** — `KDV Dahil Satış Fiyatı` metin + `₺` mojibake + binlik ayraç (`"1,049.18"`). Sayıya
   normalize edilmeli. Komisyon `23.75%` metin.
6. **Aktif/Pasif** — çoğu PASIF, az sayıda AKTIF. Tablo bunu boolean/enum tutmalı.
7. **`Frictions` kolonu boş** — Zeytuni_Ops'un sürtünme haritasıyla ilişkili olabilir; şimdilik yok say.

---

## 4. Yapılacaklar (çıktılar)

### 4.1 Veri kalitesi raporu
`docs/zeytuni-urunler-rapor.md` — §3'teki sorunları satır/SKU referanslı listele. Düzeltme önerileri ver
(çift SKU, eksik gramaj, barkod stratejisi). Karar gereken noktaları işaretle (kullanıcıya sor).

### 4.2 `products` şema önerisi (taslak, uygulanmaz)
Aynı dosyada `CREATE TABLE products (...)` taslağı öner. Hizala:
- `internal/parser/` `Order.Lines[].Barcode` ile eşleşme: webhook siparişindeki barkod → `products.barcode`
  join → baskıda ürün adı/kategori zenginleştirme. **Asıl değer bu** (printer'a operasyonel fayda).
- Kolonlar minimal tut (ERP'leşme): sku, barcode, name, category, brand, marketplace_name, price,
  is_active. Komisyon/KDV türevleri hesaplanabilir → saklamak şart değil; gerekçeyi yaz.
- Mevcut şema konvansiyonuna uy (`docs/schema.sql` stili, MySQL 8, utf8mb4).

### 4.3 Import planı (taslak, uygulanmaz)
xlsx → temiz CSV/JSON → `cmd/seed` benzeri ayrı `cmd/import-products` binary fikri. Idempotent
(`INSERT ... ON DUPLICATE KEY UPDATE` SKU üzerinden). Görseller şimdilik kapsam dışı.

---

## 5. Kapsam DIŞI (bu session'da yapma)

- `docs/schema.sql` veya herhangi bir Go koduna **dokunma**.
- Migration / import aracı **yazma**.
- Zeytuni_Ops reposuna **yazma**.
- S06 Cloudflare/Docker işini **değiştirme**.

---

## 6. Çıkış Kriteri

- [x] `docs/zeytuni-urunler-rapor.md` üretildi: veri kalitesi bulguları + karar soruları
- [x] `products` şema taslağı + parser join gerekçesi yazıldı
- [x] Import yaklaşımı (idempotent, ayrı binary) taslağı yazıldı
- [x] Kullanıcıya sorulacak açık kararlar listelendi → barkod/kapsam karara bağlandı; gramaj + çift SKU dükkana açık
- [x] Hiçbir şema/kod değişmedi; S06 etkilenmedi

---

## 7. Dikkat

- Önce operasyonel anlayış, sonra yazılım — Zeytuni felsefesi. Aşırı mühendislikten kaçın.
- xlsx okumak için PowerShell `Import-Excel` yoksa, kullanıcıdan xlsx'i UTF-8 CSV (`,` ayraç) export
  etmesini iste; mevcut bozuk CSV'yi kaynak alma.
- Gerçek karar (barkod, gramaj) kullanıcıya ait — varsayma, sor.
