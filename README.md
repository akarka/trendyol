# Trendyol Sipariş Yazdırma Sistemi 🚀

Bu yazılım, Trendyol mağazanıza gelen siparişleri otomatik olarak yakalar ve bilgisayarınıza bağlı termal yazıcıdan (veya dijital olarak dosya halinde) çıktı almanızı sağlar.

---

## 🛠️ Kurulum Adımları (Sıfırdan Başlayanlar İçin)

### 1. Docker Kurulumu
Bu sistemin çalışması için bilgisayarınızda **Docker Desktop** yüklü olmalıdır.
1. [Docker Desktop İndir](https://www.docker.com/products/docker-desktop/) adresine gidin.
2. "Download for Windows" butonuna tıklayın ve inen dosyayı kurun.
3. Kurulum bitince bilgisayarınızı yeniden başlatın.
4. Docker Desktop uygulamasını açın ve sağ üstte yeşil bir kutucuk (Running) görene kadar bekleyin.

### 2. Sistemi Başlatma
1. Bu proje klasörünü bilgisayarınızda bir yere koyun.
2. Klasörün içindeyken klavyenizden **Shift** tuşuna basılı tutun ve boş bir yere **sağ tıklayın**.
3. "PowerShell penceresini buradan açın" (veya "Terminali burada aç") seçeneğine tıklayın.
4. Açılan mavi/siyah ekrana şu komutu yapıştırın ve **Enter**'a basın:
   ```powershell
   docker-compose up -d --build
   ```
   *Bu işlem ilk seferde birkaç dakika sürebilir. Bitince ekran duracaktır.*

---

## 🧪 Sistemin Çalıştığını Test Etme

Sistemi kurdunuz, şimdi her şeyin yolunda olup olmadığını anlamak için sahte bir sipariş gönderelim:

1. Açık olan terminal penceresine şu komutu yazın ve **Enter**'a basın:
   ```powershell
   powershell.exe -ExecutionPolicy Bypass -File .	est_webhook.ps1
   ```
2. Ekranda **"OK - Tamamlandi!"** yazısını görmelisiniz.
3. Şimdi proje klasörünüze bakın. `order_123456789_... .txt` isimli yeni bir dosya oluşmuş olmalı.
4. Bu dosyayı açtığınızda Trendyol sipariş bilgilerini düzenli bir şekilde görebilirsiniz.

---

## 📖 Sıkça Sorulan Sorular

**Soru: Bilgisayarı kapatıp açınca ne yapmalıyım?**  
Cevap: Docker Desktop'ın açık olduğundan emin olun. Sistem otomatik olarak kaldığı yerden devam edecektir. Eğer çalışmazsa yukarıdaki "Sistemi Başlatma" adımındaki komutu tekrar yazmanız yeterlidir.

**Soru: Çıktıları nerede bulabilirim?**  
Cevap: Tüm sipariş çıktıları bu klasörün içinde `.txt` dosyası olarak anlık oluşur.

**Soru: Gerçek yazıcıya nasıl bağlarım?**  
Cevap: Teknik destek ekibimizle iletişime geçerek `docker-compose.yml` dosyasındaki yazıcı yolunu (`/dev/usb/lp0` gibi) kendi yazıcınıza göre güncelletmeniz yeterlidir.

---
*Hazırlayan: Trendyol Print Relay Team*
