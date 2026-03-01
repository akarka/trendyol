1. Proje Başlatma ve Dizin Yapısı (High-Level)
Bu prompt, projenin iskeletini oluşturur .

"Create a new Go module named github.com/yourorg/trendyol-print-relay. Generate the following directory structure: cmd/app, internal/listener, internal/parser, internal/printer, internal/alerter, and config . Create empty main.go inside cmd/app and empty .go files inside each internal package according to standard Go project layout ."

2. Veritabanı Şeması ve RLS Kurulumu (Cloud Layer)
Bu prompt, Supabase üzerinde çalışacak SQL dosyasını hazırlatır .

"Create a file named supabase/migrations/20260301_init.sql. Write the PostgreSQL schema for the trendyol_orders table including columns: uuid (UUID PK), order_id (VARCHAR), order_number (VARCHAR), package_status (VARCHAR), payload (JSONB), created_at, and updated_at . Add a unique constraint for order_id and package_status to ensure idempotency . Add a trigger to auto-update updated_at . Finally, enable Row Level Security (RLS) and write policies allowing service_role to INSERT and anon role to SELECT ."
+4

3. Edge Function (Webhook Receiver)
Bu prompt, Trendyol API'sinden gelen veriyi karşılayan Deno fonksiyonunu yazdırır .

"Create a Deno Edge Function in supabase/functions/trendyol-webhook/index.ts. Implement an HTTP server using std/http/server.ts that only accepts POST requests . Validate Basic Auth headers using WEBHOOK_USERNAME and WEBHOOK_PASSWORD environment variables . Parse the JSON payload and validate that id, orderNumber, and packageStatus exist . Initialize the Supabase client using SUPABASE_SERVICE_ROLE_KEY and insert the payload into the trendyol_orders table . If a Postgres unique violation (code 23505) occurs, ignore it. Always return a 200 OK response quickly to prevent Trendyol retry mechanisms ."
+4

4. Go Uygulaması - Config Modülü
Bu prompt, uygulamanın çevre değişkenlerini (ENV) güvenli bir şekilde yönetmesini sağlar .

"Implement the config/config.go file for the Go application. Create a Config struct containing SupabaseURL, SupabaseAnonKey, PrinterDevice, and LogLevel. Write a Load() function that reads these from environment variables. If SUPABASE_URL, SUPABASE_ANON_KEY, or PRINTER_DEVICE are empty, trigger a panic since these are mandatory."

5. Go Uygulaması - Parser Modülü
Bu prompt, ham JSON verisini Go yapılarına dönüştürür .

"Implement internal/parser/payload_parser.go. Define the Go structs: Order, OrderLine, and Shipment exactly mapping the necessary JSON fields (orderNumber, packageStatus, lines, shipmentAddress, etc.) . Write a ParseOrder(raw string) (*Order, error) function that unmarshals the JSON . Write a validateOrder function that returns an error if OrderID, OrderNumber are empty, or if the Lines array has 0 length ."
+2

6. Go Uygulaması - Listener Modülü (Realtime)
Bu prompt, Supabase Realtime bağlantısını kurar .

"Implement internal/listener/supabase_client.go. Write a StartRealtimeSubscription(cfg *config.Config, dbChannel chan<- string) function. Use a Supabase Realtime Go client to connect to realtime:public:trendyol_orders . Listen for INSERT events on the trendyol_orders table . When a new row is inserted, marshal the payload.New data into a string and send it to dbChannel . Implement an exponential backoff mechanism (starting at 5 seconds, max 60 seconds) for reconnection if the websocket drops ."
+4

7. Go Uygulaması - Printer ve Alerter Modülleri
Bu prompt, donanım entegrasyonlarını hazırlar .

"Implement internal/printer/escpos_printer.go. Write a Print(devicePath string, order *parser.Order) error function. Use an ESC/POS library to connect to the printer via devicePath and print the OrderNumber, iterate over the Lines to print products, and send a cut command. Then, implement internal/alerter/system_alerter.go with NotifySuccess(orderNumber string) and NotifyError(orderNumber string) functions that log the status and trigger a basic OS sound or standard output notification."
+1

8. Go Uygulaması - Main Orkestrasyonu
Bu prompt, tüm Go modüllerini bir araya getirerek ana döngüyü kurar .

"Implement cmd/app/main.go. Call config.Load(). Create a string channel dbChannel with a buffer size of 64. Start listener.StartRealtimeSubscription in a goroutine. Create a for rawPayload := range dbChannel loop in another goroutine . Inside the loop, call parser.ParseOrder. If successful, call printer.Print . Use alerter to notify success or error . Finally, implement a graceful shutdown mechanism using os/signal to listen for SIGINT and SIGTERM before exiting ."
+4

9. Docker Dağıtımı (Low-Level)
Bu prompt, uygulamanın dükkan makinesinde çalışacak konteyner yapısını oluşturur .

"Create a Dockerfile for the Go application using a multi-stage build. Use golang:alpine as the builder and a minimal base image like alpine or scratch for the final runtime. Then create a docker-compose.yml file defining a print-relay service . Include restart: unless-stopped , define a device mapping for the USB printer /dev/usb/lp0 , and configure it to read from an .env file . Set up JSON file logging with size limits ."
+4

İlk promptu kopyalayıp CLI ajanınıza vererek dizin yapısını oluşturmakla başlayabilir misiniz?