package main

import (
	"log"
	"os"

	"github.com/akarka/trendyol-print-relay/internal/alerter"
	"github.com/akarka/trendyol-print-relay/internal/parser"
	"github.com/akarka/trendyol-print-relay/internal/printer"
)

const dummyJSON = `{
  "id": "test-999",
  "orderNumber": "999888777",
  "packageStatus": "Created",
  "orderDate": "2026-03-01T10:00:00Z",
  "shipmentAddress": {
    "firstName": "Test",
    "lastName": "Kullanıcı",
    "address1": "Örnek Mah. Test Sok.",
    "city": "İstanbul",
    "district": "Kadıköy"
  },
  "lines": [
    {
      "productName": "Test Ürünü 1",
      "barcode": "123456",
      "quantity": 2,
      "amount": 45.50
    }
  ]
}`

func main() {
	device := os.Getenv("PRINTER_DEVICE")
	if device == "" {
		device = "/dev/usb/lp0"
	}

	log.Println("Dummy JSON parse ediliyor...")
	order, err := parser.ParseOrder(dummyJSON)
	if err != nil {
		log.Fatalf("[PARSE_ERROR] %v", err)
	}

	log.Printf("Yazıcıya gönderiliyor (Cihaz: %s)...", device)
	if err := printer.Print(device, order); err != nil {
		alerter.NotifyError(order.OrderNumber)
		log.Fatalf("[PRINT_ERROR] %v", err)
	}

	alerter.NotifySuccess(order.OrderNumber)
	log.Println("Test baskısı başarılı.")
}
