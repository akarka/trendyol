package server

import (
	"io"
	"log"
	"net/http"

	"github.com/akarka/trendyol/internal/db"
	"github.com/akarka/trendyol/internal/parser"
)

// handleWebhook, Trendyol'dan gelen siparişi alır.
// Trendyol retry'larını bastırmak için her durumda 200 döner.
func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[WEBHOOK] gövde okunamadı: %v", err)
		ok200(w)
		return
	}

	// parser.ParseOrder bir DBRow ({payload: ...}) bekler; ham Trendyol gövdesini sarmalıyoruz.
	wrapped := append(append([]byte(`{"payload":`), body...), '}')
	order, err := parser.ParseOrder(string(wrapped))
	if err != nil {
		log.Printf("[WEBHOOK] parse hatası: %v | ham: %s", err, string(body))
		ok200(w)
		return
	}

	inserted, err := db.InsertOrder(s.db, order, body)
	if err != nil {
		log.Printf("[WEBHOOK] DB insert hatası: %v", err)
		ok200(w)
		return
	}
	if !inserted {
		log.Printf("[WEBHOOK] duplicate sipariş, atlanıyor: %s/%s", order.OrderID, order.PackageStatus)
		ok200(w)
		return
	}

	jobID, err := db.InsertPrintJob(s.db, order.OrderID, "queued", "")
	if err != nil {
		log.Printf("[WEBHOOK] print job kaydı hatası: %v", err)
		ok200(w)
		return
	}

	select {
	case s.printCh <- PrintTask{Order: order, JobID: jobID}:
	default:
		log.Printf("[WEBHOOK] printer kuyruğu dolu, sipariş kuyruğa alınamadı: %s", order.OrderNumber)
		_ = db.UpdatePrintJob(s.db, jobID, "failed", "printer kuyruğu dolu")
	}

	ok200(w)
}

func ok200(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
