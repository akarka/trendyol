# CLAUDE.md — trendyol termal yazıcı modülü

## Proje Rolü

Bu repo, dükkan yönetim projesinin **termal yazıcı sınıfıdır**. Bağımsız bir Go servisidir; dış projeye printer interface olarak entegre edilmesi beklenir. Trendyol webhook'larını Supabase üzerinden dinler, siparişleri parse eder, termal yazıcıya (veya test dosyasına) basar.

## Mimari Akış

```
Trendyol → Supabase Edge Fn (trendyol-webhook) → trendyol_orders tablosu
→ listener (WebSocket/Realtime) → parser → printer → ESC/POS veya output.txt
```

## Paket Haritası

| Paket | Dosya | Sorumluluk |
|-------|-------|------------|
| `config` | config.go | Env var yükleme; SUPABASE_URL, SUPABASE_ANON_KEY zorunlu |
| `internal/listener` | supabase_client.go | WebSocket + Phoenix protocol; exponential backoff reconnect |
| `internal/parser` | payload_parser.go | DBRow → Order; `trendyol_orders` kolonları: order_id, order_number, package_status, payload(JSON) |
| `internal/printer` | txt_printer.go | **AKTİF**: `output.txt`'e append (TEST_MODE) |
| `internal/printer` | escpos_printer.go | **STUB**: ESC/POS impl yok, TODO |
| `internal/printer` | digital_printer.go | **UNUSED**: belirtilen dosyaya yazar, main'de çağrılmıyor |
| `internal/alerter` | system_alerter.go | log.Printf sarmalayıcı; OS bildirim TODO |

## Kritik Detaylar

- `main.go` şu an sadece `PrintToTXT` çağırıyor; `Print` (ESC/POS) ve `PrintToTextFile` (digital) devre dışı
- `TEST_MODE=true` → `output.txt`; `false` → `PRINTER_DEVICE` (örn. `/dev/usb/lp0`) zorunlu
- Supabase Realtime: `public:trendyol_orders` tablosuna `INSERT` eventi dinleniyor
- Docker volume: `./output.txt:/app/output.txt` — dosya host'ta tutulur
- `docker-compose.yml`'deki `devices:` bloğu gerçek yazıcı için yorumda

## Veri Modeli

```go
Order { OrderID, OrderNumber, PackageStatus, CreatedAt, Lines[]OrderLine, ShipmentInfo, CargoProvider }
OrderLine { ProductName, Barcode, Quantity, Price }
Shipment { FirstName, LastName, Address1, City, District, PostalCode }
```

PackageStatus değerleri: `Created`, `Cancelled`, `Delivered`, `UnSupplied`

## Geliştirme Notları

- ESC/POS implementasyonu için `escpos_printer.go` skeleton hazır; library seçilmedi
- `alerter` paketi OS ses/bildirim için genişletilmeyi bekliyor
- Gerçek yazıcıya geçiş: `TEST_MODE` kaldır, `PRINTER_DEVICE` set et, `docker-compose.yml`'de `devices:` bloğunu aç

---

## Asistan Davranış Kuralları

- Cevaplar kısa ve doğrudan. Gereksiz açıklama yok.
- Kod değişikliği: sadece ilgili satırları göster, tüm dosyayı tekrarlama.
- Övgü, teşekkür, "harika soru" gibi ifadeler yok.
- Yorum satırı yazma; yazmak gerekiyorsa tek satır, sadece "neden" açıklanır.
- Seçenekler sorulmadıkça sunma; doğrudan en iyi yaklaşımı uygula.
- Türkçe konuş.
