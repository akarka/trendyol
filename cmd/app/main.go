// cmd/app/main.go
package main
 
import (
    "log"
    "os"
    "os/signal"
    "syscall"
 
    "github.com/akarka/trendyol-print-relay/config"
    "github.com/akarka/trendyol-print-relay/internal/alerter"
    "github.com/akarka/trendyol-print-relay/internal/listener"
    "github.com/akarka/trendyol-print-relay/internal/parser"
    "github.com/akarka/trendyol-print-relay/internal/printer"
)
 
func main() {
    log.Println("Uygulama baslatiliyor...")
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
    log.Println("Uygulama kapatiliyor...")
}
