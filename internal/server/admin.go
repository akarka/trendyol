package server

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/akarka/trendyol/internal/db"
	"github.com/akarka/trendyol/internal/parser"
	"github.com/akarka/trendyol/internal/xlsxlite"
)

// handleDBBackup, tüm veritabanını mysqldump ile gzip'lenmiş .sql.gz olarak indirir.
func (s *Server) handleDBBackup(w http.ResponseWriter, r *http.Request) {
	info, err := db.ParseConnInfo(s.cfg.MySQLDSN)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DSN ayrıştırılamadı")
		return
	}

	cmd := exec.Command("mysqldump",
		"-h", info.Host, "-P", info.Port, "-u", info.User,
		"--add-drop-table", "--single-transaction", "--routines",
		info.DBName,
	)
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+info.Password)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "yedekleme başlatılamadı")
		return
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		writeError(w, http.StatusInternalServerError, "mysqldump çalıştırılamadı")
		return
	}

	filename := fmt.Sprintf("trendyol-backup-%s.sql.gz", time.Now().Format("20060102-150405"))
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)

	gz := gzip.NewWriter(w)
	if _, err := io.Copy(gz, stdout); err != nil {
		log.Printf("yedekleme akışı kesildi: %v", err)
	}
	_ = gz.Close()

	if err := cmd.Wait(); err != nil {
		log.Printf("mysqldump hata verdi: %v, stderr: %s", err, stderr.String())
	}
}

// handleDBRestore, yüklenen .sql veya .sql.gz dosyasını mysql CLI ile geri yükler.
// Yıkıcı işlemdir: mevcut tablolar dump içindeki DROP TABLE ile değiştirilir.
func (s *Server) handleDBRestore(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 200<<20)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "dosya çok büyük veya geçersiz istek")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "dosya alanı (file) bulunamadı")
		return
	}
	defer file.Close()

	var reader io.Reader = file
	if strings.HasSuffix(strings.ToLower(header.Filename), ".gz") {
		gz, err := gzip.NewReader(file)
		if err != nil {
			writeError(w, http.StatusBadRequest, "gzip dosyası okunamadı")
			return
		}
		defer gz.Close()
		reader = gz
	}

	info, err := db.ParseConnInfo(s.cfg.MySQLDSN)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DSN ayrıştırılamadı")
		return
	}

	cmd := exec.Command("mysql", "-h", info.Host, "-P", info.Port, "-u", info.User, info.DBName)
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+info.Password)
	cmd.Stdin = reader
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		writeError(w, http.StatusInternalServerError, "geri yükleme başarısız: "+stderr.String())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleOrdersExport, siparişleri (opsiyonel status filtresi) .xlsx olarak indirir.
func (s *Server) handleOrdersExport(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	orders, err := db.GetOrders(s.db, 100000, 0, status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "siparişler okunamadı")
		return
	}

	rows := [][]string{{"Sipariş No", "Durum", "Tarih", "Müşteri", "Şehir", "Ürünler", "Kargo"}}
	for _, o := range orders {
		wrapped := append(append([]byte(`{"payload":`), o.Payload...), '}')
		order, parseErr := parser.ParseOrder(string(wrapped))

		var customer, city, cargo, items string
		if parseErr == nil {
			customer = strings.TrimSpace(order.ShipmentInfo.FirstName + " " + order.ShipmentInfo.LastName)
			city = order.ShipmentInfo.City
			cargo = order.CargoProvider
			parts := make([]string, 0, len(order.Lines))
			for _, l := range order.Lines {
				parts = append(parts, fmt.Sprintf("%s x%d", l.ProductName, l.Quantity))
			}
			items = strings.Join(parts, "; ")
		}

		rows = append(rows, []string{
			o.OrderNumber, o.PackageStatus, o.CreatedAt.Format("2006-01-02 15:04"),
			customer, city, items, cargo,
		})
	}

	tmp, err := os.CreateTemp("", "siparisler-*.xlsx")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "geçici dosya oluşturulamadı")
		return
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)

	if err := xlsxlite.Write(tmpPath, "Siparişler", rows); err != nil {
		writeError(w, http.StatusInternalServerError, "xlsx oluşturulamadı")
		return
	}

	filename := fmt.Sprintf("siparisler-%s.xlsx", time.Now().Format("20060102-150405"))
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	http.ServeFile(w, r, tmpPath)
}

// handleProductsImport, yüklenen .csv/.xlsx'i import-products binary'sine (--skip-duplicates)
// devreder: DB'de SKU'su zaten var olan satırlar güncellenmez, atlanır.
func (s *Server) handleProductsImport(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 20<<20)
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "dosya çok büyük veya geçersiz istek")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "dosya alanı (file) bulunamadı")
		return
	}
	defer file.Close()

	ext := ".csv"
	if strings.HasSuffix(strings.ToLower(header.Filename), ".xlsx") {
		ext = ".xlsx"
	}

	tmp, err := os.CreateTemp("", "products-import-*"+ext)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "geçici dosya oluşturulamadı")
		return
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmp, file); err != nil {
		tmp.Close()
		writeError(w, http.StatusInternalServerError, "dosya kaydedilemedi")
		return
	}
	tmp.Close()

	cmd := exec.Command("./import-products", "--file", tmpPath, "--skip-duplicates")
	out, err := cmd.CombinedOutput()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "içe aktarma başarısız",
			"log":   string(out),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "log": string(out)})
}
