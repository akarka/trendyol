# Trendyol Sipariş Yazdırma Sistemi

Trendyol mağazanıza gelen siparişleri otomatik yakalar, veritabanına kaydeder ve bilgisayarınıza bağlı termal yazıcıdan (veya test modunda `output.txt` dosyasına) bastırır. Siparişleri ve baskıları yönetebileceğiniz bir **web arayüzü** ile gelir.

Mimari: `Trendyol → Cloudflare Tunnel → Go Sunucu (:8080) → MySQL + Yazıcı`. Web arayüzü aynı sunucudan sunulur.

Geliştirme ilerlemesi, mimari detaylar ve session geçmişi için bkz. [docs/PLAN.md](docs/PLAN.md).

---

## Kurulum Adımları (Sıfırdan Başlayanlar İçin)

### 1. Docker Kurulumu
1. [Docker Desktop İndir](https://www.docker.com/products/docker-desktop/) → "Download for Windows" → kurun.
2. Kurulum bitince bilgisayarı yeniden başlatın.
3. Docker Desktop'ı açın, sağ üstte yeşil **Running** kutucuğunu görene kadar bekleyin.

### 2. Git Kurulumu
1. [Git for Windows İndir](https://git-scm.com/download/win) → "64-bit Git for Windows Setup".
2. Kurulumda tüm seçenekleri varsayılan bırakarak tamamlayın.

### 3. Sistemi İndirme
1. Projeyi kurmak istediğiniz klasöre gidin (örn. `C:\Projeler`).
2. **Shift** + boş alana **sağ tık** → "PowerShell penceresini buradan açın".
3. Şu komutları sırayla çalıştırın:
   ```powershell
   git clone https://github.com/akarka/trendyol.git
   cd trendyol
   ```

### 4. Ayar Dosyasını (.env) Oluşturma
1. `.env.example`'ı `.env` olarak kopyalayın:
   ```powershell
   cp .env.example .env
   ```
   Varsayılan değerlerle **olduğu gibi çalışır**; lokal denemek için doldurmanız gereken bir şey yok.
2. Canlıya çıkarken şu değerleri kendi değerlerinizle değiştirin:
   - `MYSQL_ROOT_PASSWORD`, `MYSQL_PASSWORD` → MySQL şifreleri.
   - `MYSQL_DSN` içindeki şifre, `MYSQL_PASSWORD` ile **aynı** olmalı; host `mysql` kalmalı.
   - `JWT_SECRET` → uzun rastgele bir metin (web girişini güvene alır).
   - `WEBHOOK_USERNAME`, `WEBHOOK_PASSWORD` → Trendyol webhook'unun kullanacağı kullanıcı/şifre.
   - `CF_TUNNEL_TOKEN` → Cloudflare Tunnel kullanacaksanız (bkz. Adım 7); şimdilik boş bırakabilirsiniz.
   - `TEST_MODE=true` → çıktı `output.txt`'e gider. Gerçek yazıcı için bkz. SSS.

### 5. Sistemi Başlatma
```powershell
docker compose up -d --build
```
İlk seferde birkaç dakika sürer. MySQL sağlıklı olduğunda sunucu otomatik bağlanır.

### 6. Yönetici Kullanıcısı Oluşturma (web girişi için)
Web arayüzüne girebilmek için bir yönetici hesabı oluşturun:
```powershell
docker compose exec print-relay /app/seed --username admin --password Admin123
```
Artık `http://localhost:8080` adresinden `admin` / `Admin123` ile giriş yapabilirsiniz.

### 7. (Opsiyonel) Cloudflare Tunnel — Dışarıdan Erişim
Trendyol'un webhook gönderebilmesi ve arayüze internetten erişmek için:
1. cloudflare.com → **Zero Trust → Networks → Tunnels → Create a tunnel** (tip: Cloudflared).
2. Tunnel adı: `trendyol-printer`. Verilen **token**'ı kopyalayın → `.env`'de `CF_TUNNEL_TOKEN=...`.
3. **Public hostname** ekleyin: `printer.<alanadınız>` → Service `http://print-relay:8080`.
4. Tüneli başlatın:
   ```powershell
   docker compose --profile tunnel up -d --build
   ```
   Cloudflared loglarında "Registered tunnel connection" görmelisiniz.
5. Trendyol paneline webhook URL'i olarak şunu girin (BasicAuth ile `WEBHOOK_USERNAME`/`WEBHOOK_PASSWORD`):
   ```
   https://printer.<alanadınız>/webhook/trendyol
   ```

---

## Sistemin Çalıştığını Test Etme

Sahte sipariş gönderin (lokal):
```powershell
powershell.exe -ExecutionPolicy Bypass -File .\testing\test_webhook_status.ps1
```
- Ekranda **"Sunucu Yaniti: OK"** benzeri yazı görmelisiniz.
- `output.txt` dosyası oluşur ve sipariş bilgisi içine yazılır.
- Web arayüzünde sipariş listede görünür.

Farklı durumlar için `-Status` ekleyin:
```powershell
.\testing\test_webhook_status.ps1 -Status "Cancelled"   # iptal
.\testing\test_webhook_status.ps1 -Status "Delivered"   # teslim
.\testing\test_webhook_status.ps1 -Status "UnSupplied"  # tedarik edilemeyen
```

---

## Web Arayüzü Sayfaları

| Sayfa | Yol | Ne işe yarar |
|---|---|---|
| Siparişler | `/orders` | Gelen/elle girilen siparişleri listele, durum filtrele, yeniden yazdır, **Excel'e Aktar**. |
| Manuel Sipariş | `/manual` | Trendyol panelinden gelen siparişi elle girip etiket bas. |
| Ürünler | `/products` | Ürün kataloğunu listele/ara, **yeni ürün ekle**, **düzenle**, **sil**. Kategori/marka alanları mevcut değerleri önerir, yeni yazılan otomatik oluşturulur. |
| Yazıcı | `/printer` | Yazıcı durumu ve son baskı işleri. |
| Ayarlar | `/settings` | Etiket yerleşimi (A4 sticker) ve diğer ayarlar. |
| Yedekleme | `/backup` | Veritabanı tam yedeği indir/geri yükle, ürün kataloğu içe aktar (bkz. aşağıdaki bölüm). |

## Veri Yönetimi (Yedekleme / İçe-Dışa Aktarma)

**Yedekleme** sayfası (`/backup`) üç işlevi tek yerde toplar:

1. **Veritabanı Yedeği** — tüm tabloların (`trendyol_orders`, `products`, `users`, `print_jobs`, `settings`) tam yedeğini `.sql.gz` olarak indirir (`mysqldump` tabanlı, `GET /api/admin/backup`).
2. **Geri Yükleme** — yüklenen `.sql`/`.sql.gz` dosyasıyla veritabanını **tamamen değiştirir** (`mysql` CLI tabanlı, `POST /api/admin/restore`). Geri alınamaz; mevcut tüm veri silinir, dosyadakiyle değişir.
3. **Ürün İçe Aktarma** — katalog şablonundaki (bkz. `cmd/export-products` çıktısı) CSV/XLSX'i yükler; **SKU'su veritabanında zaten olan satırlar atlanır**, sadece yeni ürünler eklenir (`POST /api/products/import`, arka planda `import-products --skip-duplicates`).

Siparişler sayfasındaki **Excel'e Aktar** butonu (`GET /api/orders/export`) ise seçili durum filtresine göre sipariş listesini `.xlsx` olarak indirir.

> Ürün kataloğunun **tam round-trip** (işletmenin Excel'de toplu düzeltip geri yüklediği) akışı hâlâ CLI üzerinden çalışır — bkz. [docs/sessions/S08-manuel-siparis-katalog.md](docs/sessions/S08-manuel-siparis-katalog.md):
> ```powershell
> docker compose run --rm -v "${PWD}:/out" print-relay ./export-products --out /out/urunler.xlsx
> # işletme düzeltir →
> docker compose run --rm -v "${PWD}:/out" print-relay ./import-products --file /out/urunler.xlsx
> ```
> Bu CLI akışı **upsert**'tir (mevcut SKU güncellenir); web'deki içe aktarma ise **atlar** (mevcut SKU dokunulmaz). İhtiyacına göre seçin.

---

## Sıkça Sorulan Sorular

**Bilgisayarı kapatıp açınca?**
Docker Desktop açık olsun; sistem `restart: unless-stopped` ile kaldığı yerden devam eder. Çalışmazsa klasörde terminal açıp `docker compose up -d` yazın.

**Çıktıları nerede bulurum?**
Test modunda tüm çıktılar proje klasöründeki `output.txt` dosyasına alt alta eklenir. Geçmiş siparişler ve baskı kayıtları web arayüzünde de görünür.

**Web arayüzüne nasıl girerim?**
`http://localhost:8080` (veya Cloudflare hostname'iniz). Giriş bilgileri Adım 6'da oluşturduğunuz kullanıcı. Yeni kullanıcı için `seed` komutunu farklı `--username` ile tekrar çalıştırın.

**Gerçek yazıcıya nasıl geçerim?**
`.env`'de `TEST_MODE=false` yapın ve `PRINTER_DEVICE=/dev/usb/lp0` ayarlayın. `docker-compose.yml` içindeki `print-relay` servisinde `devices:` bloğunu (yazıcı yolu) açın. Sonra `docker compose up -d --build`.

**Şifremi/şemayı değiştirdim, MySQL eski haliyle kalıyor?**
`docs/schema.sql` yalnızca **boş veritabanında** çalışır. Şemayı sıfırlamak için (DİKKAT: tüm veri silinir): `docker compose down -v` → tekrar `up`.

**Güncelleme nasıl yüklenir?**
Proje klasöründe terminal açıp sırayla:
1. `git pull`
2. `docker compose up -d --build`

**Düzenli yedek nasıl alırım?**
Web arayüzünde **Yedekleme** (`/backup`) sayfasından "Yedek İndir" butonuyla `.sql.gz` indirin. Geri yüklemek için aynı sayfada dosyayı seçip "Geri Yükle" — bu işlem mevcut tüm veriyi siler ve dosyadakiyle değiştirir, dikkatli kullanın.

**`docker compose down -v` sonrası yedekleme/geri yükleme "Plugin caching_sha2_password" hatası veriyor?**
Volume sıfırdan oluşturulurken `docs/init-auth.sh` otomatik çalışıp DB kullanıcısını `mysql_native_password`'e çevirir (alpine tabanlı `mysqldump`/`mysql` CLI, MySQL 8'in varsayılan `caching_sha2_password` eklentisini desteklemiyor). Eğer eski bir volume'den geliyorsanız ve hata alıyorsanız: `docker compose down -v` ile volume'ü sıfırlayıp tekrar `up` edin (DİKKAT: tüm veri silinir) ya da root ile elle `ALTER USER '<MYSQL_USER>'@'%' IDENTIFIED WITH mysql_native_password BY '<MYSQL_PASSWORD>';` çalıştırın.

---
*Hazırlayan: Zze*
