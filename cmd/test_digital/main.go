package main

import (
	"log"

	"github.com/akarka/trendyol-print-relay/internal/parser"
	"github.com/akarka/trendyol-print-relay/internal/printer"
)

const dummyJSON = `{
  "id": "test-123",
  "orderNumber": "123456789",
  "packageStatus": "Created",
  "orderDate": "2026-03-01T14:30:00Z",
  "cargoProviderName": "Trendyol Express",
  "shipmentAddress": {
    "firstName": "Hasan",
    "lastName": "Demir",
    "address1": "Kadıköy Caddesi No:15 Daire:4",
    "city": "İstanbul",
    "district": "Kadıköy",
    "postalCode": "34710"
  },
  "lines": [
    {
      "productName": "Siyah T-Shirt (M)",
      "barcode": "868000123456",
      "quantity": 2,
      "amount": 250.00
    },
    {
      "productName": "Mavi Jean Pantolon",
      "barcode": "868000654321",
      "quantity": 1,
      "amount": 750.00
    }
  ]
}`

func main() {
	log.Println("Digital test siparisi parse ediliyor...")
	order, err := parser.ParseOrder(dummyJSON)
	if err != nil {
		log.Fatalf("[PARSE_ERROR] %v", err)
	}

	outputFile := "test_order_output.txt"
	log.Printf("Digital print olusturuluyor: %s...", outputFile)
	
	if err := printer.PrintToText(outputFile, order); err != nil {
		log.Fatalf("[DIGITAL_PRINT_ERROR] %v", err)
	}

	log.Println("Digital print basariyla olusturuldu. 'test_order_output.txt' dosyasini kontrol edebilirsiniz.")
}
