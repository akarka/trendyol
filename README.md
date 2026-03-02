# Trendyol Sipariş Yazdırma Sistemi 

Bu yazılım, Trendyol mağazanıza gelen siparişleri otomatik olarak yakalar ve bilgisayarınıza bağlı termal yazıcıdan (veya test modunda dijital olarak dosya halinde) çıktı almanızı sağlar.

---

## Kurulum Adımları (Sıfırdan Başlayanlar İçin)

### 1. Docker Kurulumu
Bu sistemin çalışması için bilgisayarınızda **Docker Desktop** yüklü olmalıdır.
1. [Docker Desktop İndir](https://www.docker.com/products/docker-desktop/) adresine gidin.
2. "Download for Windows" butonuna tıklayın ve inen dosyayı kurun.
3. Kurulum bitince bilgisayarınızı yeniden başlatın.
4. Docker Desktop uygulamasını açın ve sağ üstte yeşil bir kutucuk (Running) görene kadar bekleyin.

### 2. Git Kurulumu
Bu sistemi bilgisayarınıza indirmek ve güncellemeleri almak için Git gereklidir.
1. [Git for Windows İndir](https://git-scm.com/download/win) adresine gidin.
2. "64-bit Git for Windows Setup" seçeneğine tıklayarak indirin.
3. Kurulum sırasında karşınıza çıkan tüm seçenekleri varsayılan (Default) haliyle bırakarak "Next" diyerek tamamlayın.

### 3. Sistemi Bilgisayara İndirme ve Başlatma
1. Bilgisayarınızda projeyi kurmak istediğiniz klasöre gidin (örneğin: `C:\Projeler`).
2. Klasörün içindeyken klavyenizden **Shift** tuşuna basılı tutun ve boş bir yere **sağ tıklayın**.
3. "PowerShell penceresini buradan açın" (veya "Terminali burada aç") seçeneğine tıklayın.
4. Açılan siyah ekrana şu komutu yapıştırın ve **Enter**'a basın (bu, dosyaları internetten çeker):
   ```powershell
   git clone https://github.com/akarka/trendyol.git
   ```
5. İndirme bitince şu komutla klasörün içine girin:
   ```powershell
   cd trendyol
   ```
6. **Önemli:** Sistemi başlatmadan önce klasör içindeki `.env.example` dosyasının adını `.env` olarak değiştirin ve içindeki Supabase/Webhook bilgilerini kendi projenize göre doldurun.
7. Son olarak sistemi başlatmak için şu komutu yazın:
   ```powershell
   docker-compose up -d --build
   ```
   *Bu işlem ilk seferde birkaç dakika sürebilir. Bitince ekran duracaktır.*

---

## Sistemin Çalıştığını Test Etme

Sistemi kurdunuz, şimdi her şeyin yolunda olup olmadığını anlamak için sahte bir sipariş gönderelim:

1. Açık olan terminal penceresine şu komutu yazın ve **Enter**'a basın:
   ```powershell
   powershell.exe -ExecutionPolicy Bypass -File .\test_webhook_unique.ps1
   ```
2. Ekranda **"OK (Inserted) - Tamamlandi!"** benzeri bir yazı görmelisiniz.
3. Şimdi proje klasörünüze bakın. `output.txt` isimli bir dosya oluşmuş olmalı.
4. Bu dosyayı açtığınızda Trendyol sipariş bilgilerini düzenli bir şekilde görebilirsiniz. Yeni testler yaptıkça siparişler bu dosyanın sonuna eklenecektir.

---

## Sıkça Sorulan Sorular

**Soru: Bilgisayarı kapatıp açınca ne yapmalıyım?**  
Cevap: Docker Desktop'ın açık olduğundan emin olun. Sistem otomatik olarak kaldığı yerden devam edecektir. Eğer çalışmazsa klasör içinde terminal açıp `docker-compose up -d` komutunu tekrar yazmanız yeterlidir.

**Soru: Çıktıları nerede bulabilirim?**  
Cevap: Sistem şu an test modunda çalıştığı için tüm sipariş çıktıları proje klasörünün içindeki `output.txt` dosyasına anlık olarak alt alta eklenir.

**Soru: Gerçek yazıcıya nasıl bağlarım?**  
Cevap: Yazılımın içinde termal yazıcı altyapısı mevcuttur. `.env` dosyasındaki `TEST_MODE=true` ayarını kaldırıp, `docker-compose.yml` dosyasındaki yazıcı yolunu (`/dev/usb/lp0` gibi) kendi yazıcınıza göre güncelleyerek gerçek yazıcıya geçiş yapabilirsiniz.

**Soru: Sisteme yeni bir güncelleme gelirse nasıl yüklerim?**  
Cevap: Bilgisayarınızdaki proje klasörünün (`trendyol`) içinde bir terminal (veya PowerShell) açın. Ardından sırasıyla şu iki komutu yazın:
1. `git pull` (Bu komut internetteki en son yenilikleri bilgisayarınıza indirir.)
2. `docker-compose up -d --build` (Bu komut yeni indirdiğiniz kodlarla sistemi yeniden kurup başlatır.)

---
*Hazırlayan: Zze*