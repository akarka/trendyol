# Trendyol API Entegrasyonu: Webhook Kurulum Talebi

Merhaba,

Trendyol Go mağazamız (Satıcı paneli) için sipariş durum güncellemelerini anlık olarak alabilmek amacıyla webhook entegrasyonu yapmak istiyoruz. 

Teknik dokümantasyonunuzu (Özellikle "Webhook Integration" ve "Create Test Orders" bölümlerini) inceledik ve sistemimizi buna göre hazırladık. Dokümanda belirtilen "Stage ortamı endpoint kaydı" sürecini başlatmak istiyoruz.

---

## 1. Stage Ortamı Webhook Endpoint Bilgileri

Sipariş durumu güncellemeleri ("Created", "Cancelled", "Delivered", "UnSupplied") için hazırladığımız `HTTPS POST` endpoint'imiz aşağıdadır:

- **Endpoint URL:** `https://ruyxkthhvhbygwhwswri.supabase.co/functions/v1/trendyol-webhook`

*(Not: Opsiyonel path parametreleri kullanmayı düşünmüyoruz, tüm statüler için doğrudan bu kök URL'e gönderim yapılabilir.)*

### Kimlik Doğrulama (Authentication)

Endpoint'imiz, dokümanlarınızda zorunlu tutulduğu üzere **HTTP Basic Authentication** yöntemi ile korunmaktadır ve başarılı işlemlerde sadece `2xx` dönmektedir.

- İsteklerinizi gönderirken HTTP `Authorization` başlığında bu kimlik bilgilerini kullanmanızı rica ediyoruz.
- Gerekli olan `Kullanıcı Adı` ve `Şifre` bilgilerini sizinle **güvenli bir şekilde nasıl paylaşabileceğimizi** (özel bir form, şifreli bir e-posta vb.) iletir misiniz?

---

## 2. Sonraki Adımlar ve Test

Kimlik bilgilerimizi size iletip Stage kaydımızı tamamladıktan sonra:

1. Kendi sistemimizde `https://stageapi.tgoapis.com/integrator/grocery-test-order/orders/instant` üzerinden test siparişleri oluşturacağız.
2. Webhook'un sorunsuz tetiklendiğini ve verinin sistemimize aktığını doğrulayacağız.
3. Testleri başarıyla bitirdikten sonra, Production (Canlı) ortamı kaydı için sizinle tekrar iletişime geçeceğiz.

Stage ortamı webhook tanımlamamızı yapabilmeniz için izlememiz gereken adımları ve yetkili iletişim kanalını paylaşmanızı rica ederiz.

Yardımlarınız için teşekkürler.

**Teknik Yetkili:**  
[Adınız / Firma Adı]  
[Satıcı ID / Store ID]  
[Teknik Yetkili E-posta Adresi]
