trendyol-print-relay
Teknik Tasarım Dökümanı
v1.0  ·  Mimari, Şema, Güvenlik & Implementasyon Rehberi
1. Sistem Mimarisi ve Proje Yapısı
Proje iki katmanlı bir mimariden oluşur: Trendyol webhook'larını güvenli biçimde alan bulut katmanı (Supabase Edge Function) ve yerel ağdaki termal yazıcıyı yöneten yerel katman (Dockerize Go uygulaması). İki katman arasındaki tek iletişim kanalı Supabase Realtime'dır; yerel uygulama dışarıya port açmaz.

1.1 Katmanlar ve Sorumluluklar
Katman	Bileşen	Sorumluluk	Teknoloji
Bulut	Edge Function	Webhook alımı, auth, DB yazma	Deno / TypeScript
Bulut	PostgreSQL	Sipariş depolama, idempotency	Supabase (PostgreSQL 15)
Bulut	Realtime	Yeni kayıt bildirimi	Supabase Realtime (WebSocket)
Yerel	listener	Realtime aboneliği, channel yönetimi	Go
Yerel	parser	JSON → Go struct dönüşümü	Go
Yerel	printer	Termal baskı komutları	Go / ESC-POS
Yerel	alerter	Ses ve OS bildirimi	Go

1.2 Veri Akışı
Trendyol API
    │
    │  POST /webhook  (Basic Auth)
    ▼
Edge Function (Supabase)
    │
    │  INSERT → trendyol_orders
    ▼
PostgreSQL (Supabase)
    │
    │  Realtime CDC (INSERT event)
    ▼
Go Listener (Yerel Docker)
    │
    ├──► Parser  →  Go Struct
    │
    ├──► Printer  →  ESC/POS  →  Termal Yazıcı
    │
    └──► Alerter  →  Ses / OS Bildirimi

1.3 Dizin Yapısı (Go Uygulaması)
trendyol-print-relay/
├── cmd/
│   └── app/
│       └── main.go                 # Uygulama giriş noktası; goroutine yönetimi
├── internal/
│   ├── listener/
│   │   └── supabase_client.go      # Supabase Realtime aboneliği
│   ├── parser/
│   │   └── payload_parser.go       # JSON ayrıştırma ve struct dönüşümleri
│   ├── printer/
│   │   └── escpos_printer.go       # Termal yazıcı ESC/POS komutları
│   └── alerter/
│       └── system_alerter.go       # OS seviyesi ses/bildirim tetikleyicileri
├── config/
│   └── config.go                   # ENV okuma ve doğrulama — YENİ
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
ℹ  NOT
config/ paketi tüm environment variable okuma ve doğrulama işlemlerini merkezi bir yerde toplar. main.go veya diğer paketler doğrudan os.Getenv çağırmamalıdır.

1.4 Environment Değişkenleri
Değişken	Açıklama	Örnek / Format	Zorunlu
SUPABASE_URL	Supabase proje URL'i	https://xyz.supabase.co	Evet
SUPABASE_ANON_KEY	Realtime dinleme için anon key	eyJhb...	Evet
WEBHOOK_USERNAME	Edge Function Basic Auth kullanıcı adı	trendyol-hook	Evet
WEBHOOK_PASSWORD	Edge Function Basic Auth şifresi (min 32 karakter)	random-uuid-benzeri	Evet
PRINTER_DEVICE	Yazıcı cihaz yolu	/dev/usb/lp0	Evet
LOG_LEVEL	Log seviyesi	info | debug | warn	Hayır
⚠  DİKKAT
Şifreler ve key'ler docker-compose.yml içine düz metin yazılmamalıdır. Docker secrets veya .env dosyası (git-ignored) kullanılmalıdır.
2. Veritabanı Şeması
Trendyol sipariş webhook'larını depolayan ve mükerrer kayıt oluşumunu engelleyen PostgreSQL tablo yapısı aşağıda tanımlanmıştır.
2.1 Tablo Tanımı
CREATE TABLE trendyol_orders (
    uuid            UUID                     PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        VARCHAR(255)             NOT NULL,
    order_number    VARCHAR(255)             NOT NULL,
    package_status  VARCHAR(50)              NOT NULL,
    payload         JSONB                    NOT NULL,
    created_at      TIMESTAMPTZ              DEFAULT NOW(),
    updated_at      TIMESTAMPTZ              DEFAULT NOW(),
 
    -- Idempotency: aynı paket + statü çifti yalnızca bir kez saklanır
    CONSTRAINT unique_order_status UNIQUE (order_id, package_status)
);
 
-- updated_at otomatik güncelleme
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
 
CREATE TRIGGER trendyol_orders_updated_at
    BEFORE UPDATE ON trendyol_orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
2.2 İndeksler
-- Sık sorgulanan alanlar için performans indeksleri
CREATE INDEX idx_trendyol_orders_order_id       ON trendyol_orders (order_id);
CREATE INDEX idx_trendyol_orders_created_at     ON trendyol_orders (created_at DESC);
CREATE INDEX idx_trendyol_orders_status         ON trendyol_orders (package_status);
 
-- Payload içindeki yüksek kardinaliteli alanlar için (opsiyonel, trafik hacmine göre)
CREATE INDEX idx_trendyol_payload_merchant
    ON trendyol_orders ((payload->>'merchantId'));
2.3 Row Level Security (RLS)
-- RLS aktif et
ALTER TABLE trendyol_orders ENABLE ROW LEVEL SECURITY;
 
-- Yalnızca Edge Function (service_role) INSERT yapabilir
CREATE POLICY "edge_function_insert"
    ON trendyol_orders FOR INSERT
    TO service_role
    WITH CHECK (true);
 
-- Anon key yalnızca SELECT yapabilir (Go listener Realtime için)
CREATE POLICY "listener_select"
    ON trendyol_orders FOR SELECT
    TO anon
    USING (true);
 
-- Hiçbir rol UPDATE veya DELETE yapamaz
-- (politika tanımlanmadıysa varsayılan DENY)
✓  ÖNERİ
uuid_generate_v4() yerine gen_random_uuid() kullanılmıştır. Bu fonksiyon pgcrypto extension'ına ihtiyaç duymaz ve PostgreSQL 13+ ile hazır gelir.
✕  KRİTİK
Go listener, Realtime aboneliği için service_role değil anon key kullanmalıdır. service_role key tüm RLS politikalarını bypass eder ve yalnızca Edge Function'a özeldir.
2.4 Kolon Referansı
Kolon	Tip	Kısıt	Açıklama
uuid	UUID	PK, NOT NULL	Supabase dahili satır kimliği
order_id	VARCHAR(255)	NOT NULL	Trendyol paket ID (payload.id)
order_number	VARCHAR(255)	NOT NULL	Müşteri siparişi numarası
package_status	VARCHAR(50)	NOT NULL	Paket durumu (CREATED, PICKING, vb.)
payload	JSONB	NOT NULL	Ham Trendyol webhook payload'ı
created_at	TIMESTAMPTZ	DEFAULT NOW()	Kaydın DB'ye ilk girdiği zaman
updated_at	TIMESTAMPTZ	Trigger ile güncellenir	Son güncelleme zamanı
3. Supabase Edge Function — Webhook Alıcısı
Trendyol'dan gelen POST isteklerini karşılayan, kimlik doğrulayan ve veritabanına yazan Deno/TypeScript fonksiyonu. Fonksiyonun birincil görevi hızlı 2xx yanıt dönerek Trendyol'un retry mekanizmasını tetiklememektir.
3.1 İstek Yaşam Döngüsü
Adım	İşlem	Başarı	Hata
1	Basic Auth doğrulama	Devam	401 Unauthorized
2	Payload JSON parse	Devam	400 Bad Request
3	Payload alanı doğrulama	Devam	400 Bad Request
4	DB INSERT	200 OK	Unique hatası → 200 OK (beklenen), Diğer → 500 loglanır
5	Yanıt dön	200 OK	—
3.2 Tam Implementasyon
import { serve } from "https://deno.land/std@0.168.0/http/server.ts"
import { createClient } from "https://esm.sh/@supabase/supabase-js@2"
 
// ── Sabitler ──────────────────────────────────────────────────────────────
const SUPABASE_URL        = Deno.env.get("SUPABASE_URL")!
const SUPABASE_SERVICE_KEY = Deno.env.get("SUPABASE_SERVICE_ROLE_KEY")!
const WEBHOOK_USERNAME    = Deno.env.get("WEBHOOK_USERNAME")!
const WEBHOOK_PASSWORD    = Deno.env.get("WEBHOOK_PASSWORD")!
 
// PostgreSQL unique constraint hata kodu
const PG_UNIQUE_VIOLATION = "23505"
 
// ── Yardımcılar ───────────────────────────────────────────────────────────
function isValidBasicAuth(header: string | null): boolean {
  if (!header?.startsWith("Basic ")) return false
  const decoded = atob(header.slice(6))
  const [user, pass] = decoded.split(":")
  return user === WEBHOOK_USERNAME && pass === WEBHOOK_PASSWORD
}
 
interface TrendyolPayload {
  id: string
  orderNumber: string
  packageStatus: string
  [key: string]: unknown
}
 
function validatePayload(data: unknown): data is TrendyolPayload {
  if (typeof data !== "object" || data === null) return false
  const d = data as Record<string, unknown>
  return (
    typeof d.id === "string" && d.id.length > 0 &&
    typeof d.orderNumber === "string" && d.orderNumber.length > 0 &&
    typeof d.packageStatus === "string" && d.packageStatus.length > 0
  )
}
 
// ── Ana Handler ───────────────────────────────────────────────────────────
serve(async (req: Request) => {
  // 1. Yalnızca POST kabul et
  if (req.method !== "POST") {
    return new Response("Method Not Allowed", { status: 405 })
  }
 
  // 2. Basic Auth kontrolü
  if (!isValidBasicAuth(req.headers.get("Authorization"))) {
    return new Response("Unauthorized", { status: 401 })
  }
 
  // 3. Payload parse
  let payload: unknown
  try {
    payload = await req.json()
  } catch {
    return new Response("Bad Request: invalid JSON", { status: 400 })
  }
 
  // 4. Payload doğrulama
  if (!validatePayload(payload)) {
    return new Response(
      "Bad Request: id, orderNumber ve packageStatus alanları zorunludur",
      { status: 400 }
    )
  }
 
  // 5. Supabase client
  const supabase = createClient(SUPABASE_URL, SUPABASE_SERVICE_KEY)
 
  // 6. Veritabanına yaz
  const { error } = await supabase
    .from("trendyol_orders")
    .insert({
      order_id:       payload.id,
      order_number:   payload.orderNumber,
      package_status: payload.packageStatus,
      payload:        payload,
    })
 
  if (error) {
    if (error.code === PG_UNIQUE_VIOLATION) {
      // Mükerrer kayıt — Trendyol'un retry'ını kesmek için 200 dön
      console.info(`[DUPLICATE] order_id=${payload.id} status=${payload.packageStatus}`)
    } else {
      // Beklenmedik hata — logla ama 200 dön (retry istemiyoruz)
      console.error(`[DB_ERROR] code=${error.code} message=${error.message}`, payload)
    }
  }
 
  // 7. Her durumda 2xx dön
  return new Response("OK", { status: 200 })
})
3.3 Güvenlik Notları
ℹ  NOT
Basic Auth yeterince güçlüdür ancak şifre en az 32 karakter, rastgele üretilmiş olmalıdır. Trendyol webhook panelinden bu credential'lar girilir.
⚠  DİKKAT
Edge Function, service_role key ile çalışır çünkü INSERT için RLS bypass gerekmektedir. Bu key hiçbir zaman istemci tarafında kullanılmamalı, yalnızca bu server-side fonksiyona özeldir.
ℹ  NOT
Trendyol retry mekanizması: Trendyol, 2xx dışında bir yanıt aldığında isteği tekrar gönderir. Bu nedenle beklenmedik DB hatalarında bile 200 döndürülmektedir. Veri kaybı riski düşüktür çünkü Trendyol retry yapar; ancak loglar mutlaka izlenmelidir.
4. Go Dinleyici ve Ayrıştırıcı Modülleri
Supabase Realtime'ı dinleyen ve yeni sipariş kayıtlarını termal yazıcıya ileten yerel Go uygulaması. Uygulama Docker container içinde çalışır ve dış dünyaya herhangi bir port açmaz.
4.1 Uygulama Başlangıç Noktası (main.go)
package main
 
import (
    "log"
    "os"
    "os/signal"
    "syscall"
 
    "github.com/yourorg/trendyol-print-relay/config"
    "github.com/yourorg/trendyol-print-relay/internal/alerter"
    "github.com/yourorg/trendyol-print-relay/internal/listener"
    "github.com/yourorg/trendyol-print-relay/internal/parser"
    "github.com/yourorg/trendyol-print-relay/internal/printer"
)
 
func main() {
    cfg := config.Load() // ENV doğrulama; eksikse panic
 
    dbChannel := make(chan string, 64) // Buffer: ani sipariş yığılmasına karşı
 
    // Realtime dinleyiciyi ayrı goroutine'de başlat
    go listener.StartRealtimeSubscription(cfg, dbChannel)
 
    // Sipariş işleme döngüsü
    go func() {
        for rawPayload := range dbChannel {
            order, err := parser.ParseOrder(rawPayload)
            if err != nil {
                log.Printf("[PARSE_ERROR] %v | raw: %s", err, rawPayload)
                continue
            }
 
            if err := printer.Print(cfg.PrinterDevice, order); err != nil {
                log.Printf("[PRINT_ERROR] order=%s err=%v", order.OrderNumber, err)
                alerter.NotifyError(order.OrderNumber)
                continue
            }
 
            alerter.NotifySuccess(order.OrderNumber)
        }
    }()
 
    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Uygulama kapatılıyor...")
}
4.2 Listener (supabase_client.go)
package listener
 
import (
    "log"
    "time"
 
    "github.com/yourorg/trendyol-print-relay/config"
    // Önerilen kütüphane: github.com/supabase-community/realtime-go
)
 
const (
    reconnectDelay    = 5 * time.Second
    maxReconnectDelay = 60 * time.Second
)
 
func StartRealtimeSubscription(cfg *config.Config, dbChannel chan<- string) {
    delay := reconnectDelay
 
    for {
        err := subscribe(cfg, dbChannel)
        if err != nil {
            log.Printf("[REALTIME] Bağlantı hatası: %v. %v sonra yeniden denenecek.", err, delay)
            time.Sleep(delay)
 
            // Exponential backoff
            delay = min(delay*2, maxReconnectDelay)
        } else {
            delay = reconnectDelay // Başarılı bağlantıda sıfırla
        }
    }
}
 
func subscribe(cfg *config.Config, dbChannel chan<- string) error {
    client := realtime.NewClient(cfg.SupabaseURL, cfg.SupabaseAnonKey)
 
    err := client.Connect()
    if err != nil {
        return fmt.Errorf("WebSocket bağlantısı kurulamadı: %w", err)
    }
    defer client.Disconnect()
 
    channel := client.Channel("realtime:public:trendyol_orders")
    channel.OnPostgresChanges(realtime.PostgresChangesOptions{
        Event:  "INSERT",
        Schema: "public",
        Table:  "trendyol_orders",
    }, func(payload realtime.PostgresChangesPayload) {
        rawJSON, err := json.Marshal(payload.New)
        if err != nil {
            log.Printf("[REALTIME] Payload marshal hatası: %v", err)
            return
        }
        dbChannel <- string(rawJSON)
    })
 
    return channel.Subscribe() // Bağlantı kapanana kadar bloklar
}
4.3 Parser (payload_parser.go)
package parser
 
import (
    "encoding/json"
    "fmt"
    "time"
)
 
// Order, bir Trendyol sipariş paketini temsil eder
type Order struct {
    OrderID       string     `json:"id"`
    OrderNumber   string     `json:"orderNumber"`
    PackageStatus string     `json:"packageStatus"`
    CreatedAt     time.Time  `json:"orderDate"`
    Lines         []OrderLine `json:"lines"`
    ShipmentInfo  Shipment   `json:"shipmentAddress"`
    CargoProvider string     `json:"cargoProviderName"`
}
 
type OrderLine struct {
    ProductName string  `json:"productName"`
    Barcode      string  `json:"barcode"`
    Quantity     int     `json:"quantity"`
    Price        float64 `json:"amount"`
}
 
type Shipment struct {
    FirstName   string `json:"firstName"`
    LastName    string `json:"lastName"`
    Address1    string `json:"address1"`
    City        string `json:"city"`
    District    string `json:"district"`
    PostalCode  string `json:"postalCode"`
}
 
// ParseOrder, Realtime'dan gelen ham JSON'ı Order struct'ına çevirir
func ParseOrder(raw string) (*Order, error) {
    var order Order
    if err := json.Unmarshal([]byte(raw), &order); err != nil {
        return nil, fmt.Errorf("JSON parse hatası: %w", err)
    }
 
    if err := validateOrder(&order); err != nil {
        return nil, fmt.Errorf("geçersiz sipariş verisi: %w", err)
    }
 
    return &order, nil
}
 
func validateOrder(o *Order) error {
    if o.OrderID == "" {
        return fmt.Errorf("order_id boş olamaz")
    }
    if o.OrderNumber == "" {
        return fmt.Errorf("order_number boş olamaz")
    }
    if len(o.Lines) == 0 {
        return fmt.Errorf("sipariş en az bir ürün içermelidir")
    }
    return nil
}
4.4 Hata Senaryoları ve Davranışlar
Senaryo	Davranış	Bildirim
Realtime bağlantısı koptu	Exponential backoff ile yeniden bağlanma (maks 60s)	Log
JSON parse hatası	Satır atlanır, hata loglanır, uygulama devam eder	Log
Yazıcı bağlantısı yok	Hata loglanır, OS bildirimi tetiklenir	Log + Alerter
Channel tampon doldu (64)	Yeni kayıtlar beklemeye alınır; listener yavaşlar	Log
Uygulama SIGTERM aldı	Graceful shutdown; aktif baskı tamamlanır	—
ℹ  NOT
Channel buffer boyutu (64) konfigürasyon değişkeni yapılabilir. Çok hızlı sipariş akışı varsa bu değer artırılmalı veya worker pool mimarisi tercih edilmelidir.
4.5 Docker Konfigürasyonu
# docker-compose.yml
version: "3.9"
 
services:
  print-relay:
    build: .
    restart: unless-stopped
    devices:
      - "/dev/usb/lp0:/dev/usb/lp0"   # Termal yazıcı erişimi
    env_file:
      - .env                            # Git'e eklenmez
    environment:
      LOG_LEVEL: "info"
    volumes:
      - /etc/localtime:/etc/localtime:ro  # Doğru timezone
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
✓  ÖNERİ
Dockerfile'da distroless veya scratch base image kullanılması önerilir. Bu, container boyutunu küçültür ve saldırı yüzeyini azaltır.
5. Güvenlik Özeti
Risk	Kontrol Mekanizması	Durum
Yetkisiz webhook isteği	Basic Auth (min 32 karakter şifre)	Zorunlu
Mükerrer veri	DB UNIQUE constraint (order_id, package_status)	Yerleşik
RLS bypass	anon key sadece SELECT; service_role sadece INSERT	Zorunlu
Credential sızıntısı	.env dosyası, Docker secrets, git-ignored	Zorunlu
Payload enjeksiyonu	validatePayload() fonksiyonu ile alan doğrulama	Zorunlu
Yazıcı yetkisiz erişimi	Docker devices — yalnızca ilgili container erişir	Yerleşik
Container ayrıcalıkları	Dockerfile'da USER direktifi ile root-dışı çalışma	Önerilir

Bu döküman projenin canlı implementasyonunu yansıtacak şekilde güncel tutulmalıdır.