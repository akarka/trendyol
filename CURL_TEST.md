# Webhook Test Komutları

En kolay test yöntemi için oluşturulan PowerShell scriptini kullanabilirsiniz.

## PowerShell ile Test (Önerilen)

Terminalinizde aşağıdaki komutu çalıştırın:

```powershell
powershell.exe -ExecutionPolicy Bypass -File .\test_webhook.ps1
```

## Manuel PowerShell Komutu (Script kullanmak istemezseniz)

Script dosyasının içeriğini (test_webhook.ps1) kopyalayıp satır numaraları olmadan yapıştırabilirsiniz.

## Beklenen Sonuç
İstek başarılı olduğunda terminalde `OK` yanıtını almalısınız.
Ayrıca proje dizininde `order_123456789_[timestamp].txt` isimli bir dosya oluşacaktır.
