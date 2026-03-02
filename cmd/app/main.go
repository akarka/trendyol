package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/akarka/trendyol/config"
	"github.com/akarka/trendyol/internal/alerter"
	"github.com/akarka/trendyol/internal/listener"
	"github.com/akarka/trendyol/internal/parser"
	"github.com/akarka/trendyol/internal/printer"
)

func main() {
	cfg := config.Load()

	dbChannel := make(chan string, 64)

	go listener.StartRealtimeSubscription(cfg, dbChannel)

	go func() {
		for rawPayload := range dbChannel {
			order, err := parser.ParseOrder(rawPayload)
			if err != nil {
				log.Printf("[PARSE_ERROR] %v | raw: %s", err, rawPayload)
				continue
			}

			// Force writing to TXT file as per directives
			if err := printer.PrintToTXT(order); err != nil {
				log.Printf("[PRINT_ERROR] order=%s err=%v", order.OrderNumber, err)
				alerter.NotifyError(order.OrderNumber)
				continue
			}

			alerter.NotifySuccess(order.OrderNumber)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Uygulama kapatiliyor...")
}
