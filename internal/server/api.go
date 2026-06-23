package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/akarka/trendyol/internal/db"
	"github.com/akarka/trendyol/internal/parser"
)

func (s *Server) handleListOrders(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 50)
	offset := queryInt(r, "offset", 0)
	status := r.URL.Query().Get("status")

	orders, err := db.GetOrders(s.db, limit, offset, status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "siparişler okunamadı")
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

func (s *Server) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderID")
	order, err := db.GetOrderByID(s.db, orderID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "sipariş bulunamadı")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "sipariş okunamadı")
		return
	}
	writeJSON(w, http.StatusOK, order)
}

// handleReprint, kayıtlı siparişi yeniden baskı kuyruğuna alır. Her çağrı yeni print_job açar.
func (s *Server) handleReprint(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderID")
	row, err := db.GetOrderByID(s.db, orderID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "sipariş bulunamadı")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "sipariş okunamadı")
		return
	}

	wrapped := append(append([]byte(`{"payload":`), row.Payload...), '}')
	order, err := parser.ParseOrder(string(wrapped))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "sipariş verisi ayrıştırılamadı")
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
		writeError(w, http.StatusServiceUnavailable, "printer kuyruğu dolu")
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]interface{}{"job_id": jobID, "status": "queued"})
}

func (s *Server) handlePrinterStatus(w http.ResponseWriter, r *http.Request) {
	jobs, err := db.GetPrintJobs(s.db, 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "print job'lar okunamadı")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"test_mode": s.cfg.TestMode,
		"device":    s.cfg.PrinterDevice,
		"jobs":      jobs,
	})
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	jobs, err := db.GetPrintJobs(s.db, 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "loglar okunamadı")
		return
	}
	writeJSON(w, http.StatusOK, jobs)
}

// handleExportPrintJobs, ?date=YYYY-MM-DD (varsayılan: bugün) gününün print
// job'larını CSV'ye export eder. DB'den yeniden üretir; tekrar tekrar çağrılabilir.
func (s *Server) handleExportPrintJobs(w http.ResponseWriter, r *http.Request) {
	day := time.Now()
	if ds := r.URL.Query().Get("date"); ds != "" {
		d, err := time.Parse("2006-01-02", ds)
		if err != nil {
			writeError(w, http.StatusBadRequest, "geçersiz tarih, format: YYYY-MM-DD")
			return
		}
		day = d
	}

	path, n, err := s.exporter.ExportDay(day)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "export başarısız: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"file": path, "count": n})
}

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := db.GetSettings(s.db)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ayarlar okunamadı")
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) handlePutSetting(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var body struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz istek gövdesi")
		return
	}
	if err := db.UpsertSetting(s.db, key, body.Value); err != nil {
		writeError(w, http.StatusInternalServerError, "ayar kaydedilemedi")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"key": key, "value": body.Value})
}

func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
