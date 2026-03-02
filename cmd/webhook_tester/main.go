package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/akarka/trendyol/internal/parser"
	"github.com/akarka/trendyol/internal/printer"
)

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Sadece POST istekleri kabul edilir", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Gövde okunamadı", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Println("Yeni webhook alindi, parse ediliyor...")
	
	order, err := parser.ParseOrder(string(body))
	if err != nil {
		log.Printf("[PARSE_ERROR] %v
Raw: %s", err, string(body))
		http.Error(w, fmt.Sprintf("Parse hatasi: %v", err), http.StatusBadRequest)
		return
	}

	outputFile := fmt.Sprintf("order_%s_%d.txt", order.OrderNumber, time.Now().Unix())
	log.Printf("Siparis gecerli. Digital print olusturuluyor: %s...", outputFile)

	if err := printer.PrintToText(outputFile, order); err != nil {
		log.Printf("[PRINT_ERROR] %v", err)
		http.Error(w, "Yazdirma hatasi", http.StatusInternalServerError)
		return
	}

	log.Printf("Basarili: %s olusturuldu.", outputFile)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Siparis basariyla alindi ve yazdirildi.
"))
}

func main() {
	http.HandleFunc("/webhook", webhookHandler)

	port := ":8080"
	log.Printf("Webhook test sunucusu baslatiliyor. Port%s dinleniyor...
", port)
	log.Printf("Test icin su adrese POST atabilirsiniz: http://localhost%s/webhook
", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Sunucu hatasi: %v", err)
	}
}
