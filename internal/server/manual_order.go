package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/akarka/trendyol/internal/db"
	"github.com/akarka/trendyol/internal/parser"
)

func (s *Server) handleListProducts(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "1"
	products, err := db.GetProducts(s.db, activeOnly)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ürünler okunamadı")
		return
	}
	writeJSON(w, http.StatusOK, products)
}

type manualOrderReq struct {
	CustomerName string `json:"customer_name"`
	Lines        []struct {
		SKU      string `json:"sku"`
		Quantity int    `json:"quantity"`
	} `json:"lines"`
}

// handleManualOrder, web arayüzünden elle girilen siparişi trendyol_orders'a yazar ve baskı kuyruğuna alır.
// Payload, parser.Order şemasıyla aynıdır; böylece yeniden baskı (reprint) aynı yoldan çalışır.
func (s *Server) handleManualOrder(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req manualOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz istek gövdesi")
		return
	}
	if len(req.Lines) == 0 {
		writeError(w, http.StatusBadRequest, "en az bir ürün gerekli")
		return
	}

	now := time.Now()
	order := &parser.Order{
		OrderID:       "MAN-" + strconv.FormatInt(now.UnixNano(), 10),
		OrderNumber:   "MAN-" + now.Format("20060102-150405"),
		PackageStatus: "Created",
		CreatedAt:     now,
		ShipmentInfo:  parser.Shipment{FirstName: req.CustomerName},
	}

	for _, l := range req.Lines {
		if l.Quantity <= 0 {
			writeError(w, http.StatusBadRequest, "adet 0'dan büyük olmalı")
			return
		}
		p, err := db.GetProductBySKU(s.db, l.SKU)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ürün bulunamadı: "+l.SKU)
			return
		}
		order.Lines = append(order.Lines, parser.OrderLine{
			ProductName: p.Name,
			Barcode:     p.Barcode,
			Quantity:    l.Quantity,
			Price:       p.Price,
		})
	}

	payload, err := json.Marshal(order)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "sipariş serileştirilemedi")
		return
	}
	if _, err := db.InsertOrder(s.db, order, payload); err != nil {
		writeError(w, http.StatusInternalServerError, "sipariş kaydedilemedi")
		return
	}

	jobID, err := db.InsertPrintJob(s.db, order.OrderID, "queued", "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "print job oluşturulamadı")
		return
	}
	select {
	case s.printCh <- PrintTask{Order: order, JobID: jobID}:
	default:
		_ = db.UpdatePrintJob(s.db, jobID, "failed", "printer kuyruğu dolu")
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"order_id":     order.OrderID,
		"order_number": order.OrderNumber,
		"job_id":       jobID,
	})
}
