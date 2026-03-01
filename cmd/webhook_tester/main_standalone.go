package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Order struct {
	OrderNumber   string `json:"orderNumber"`
	PackageStatus string `json:"packageStatus"`
	OrderDate     string `json:"orderDate"`
	ShipmentInfo  struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Address1  string `json:"address1"`
		City      string `json:"city"`
		District  string `json:"district"`
	} `json:"shipmentAddress"`
	Lines []struct {
		ProductName string  `json:"productName"`
		Barcode     string  `json:"barcode"`
		Quantity    int     `json:"quantity"`
		Amount      float64 `json:"amount"`
	} `json:"lines"`
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	body, _ := io.ReadAll(r.Body)
	var order Order
	if err := json.Unmarshal(body, &order); err != nil {
		log.Printf("Parse error: %v", err)
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	var sb strings.Builder
	sb.WriteString("==========================================\n")
	sb.WriteString("            TRENDYOL SIPARIS             \n")
	sb.WriteString("==========================================\n")
	sb.WriteString(fmt.Sprintf("Siparis No: %s\n", order.OrderNumber))
	sb.WriteString(fmt.Sprintf("Tarih:      %s\n", order.OrderDate))
	sb.WriteString("------------------------------------------\n")
	for _, l := range order.Lines {
		sb.WriteString(fmt.Sprintf("%dx %s\n", l.Quantity, l.ProductName))
	}
	sb.WriteString("------------------------------------------\n")
	sb.WriteString(fmt.Sprintf("Adres: %s %s, %s/%s\n", 
		order.ShipmentInfo.FirstName, order.ShipmentInfo.LastName,
		order.ShipmentInfo.District, order.ShipmentInfo.City))
	sb.WriteString("==========================================\n")

	filename := fmt.Sprintf("order_%s_%d.txt", order.OrderNumber, time.Now().Unix())
	os.WriteFile(filename, []byte(sb.String()), 0644)

	log.Printf("Order %s saved to %s", order.OrderNumber, filename)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	http.HandleFunc("/webhook", webhookHandler)
	log.Println("Standalone Webhook Tester starting on :8080...")
	http.ListenAndServe(":8080", nil)
}
