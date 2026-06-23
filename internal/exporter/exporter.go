package exporter

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/akarka/trendyol/internal/db"
)

// Uploader, üretilen CSV'yi uzak bir hedefe yükler. Şimdilik sadece local disk
// kullanılıyor; R2/S3 entegrasyonu bu interface'in arkasına eklenir.
type Uploader interface {
	Upload(localPath, remoteName string) error
}

// NoopUploader local-disk-only mod; hiçbir yere yüklemez.
type NoopUploader struct{}

func (NoopUploader) Upload(localPath, remoteName string) error { return nil }

type Exporter struct {
	db       *sqlx.DB
	dir      string
	loc      *time.Location
	uploader Uploader
}

func New(database *sqlx.DB, dir, tz string, uploader Uploader) *Exporter {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Printf("[EXPORT_UYARI] saat dilimi '%s' yüklenemedi, lokal kullanılıyor: %v", tz, err)
		loc = time.Local
	}
	if uploader == nil {
		uploader = NoopUploader{}
	}
	return &Exporter{db: database, dir: dir, loc: loc, uploader: uploader}
}

// ExportDay, verilen güne ait tüm print job'ları DB'den çekip o günün CSV'sini
// bütünüyle yeniden yazar. Idempotent: aynı gün defalarca çağrılabilir, sonuç
// aynıdır. Yazım geçici dosya + atomik rename ile yapılır; crash anında yarım
// dosya kalmaz. Kaynak doğruluk DB'de olduğundan CSV her zaman yeniden üretilebilir.
func (e *Exporter) ExportDay(day time.Time) (string, int, error) {
	d := day.In(e.loc)
	start := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, e.loc)
	end := start.AddDate(0, 0, 1)

	jobs, err := db.GetPrintJobsBetween(e.db, start, end)
	if err != nil {
		return "", 0, fmt.Errorf("print job sorgusu: %w", err)
	}

	if err := os.MkdirAll(e.dir, 0o755); err != nil {
		return "", 0, fmt.Errorf("export dizini: %w", err)
	}

	name := start.Format("2006-01-02") + ".csv"
	finalPath := filepath.Join(e.dir, name)

	tmp, err := os.CreateTemp(e.dir, ".tmp-*.csv")
	if err != nil {
		return "", 0, fmt.Errorf("geçici dosya: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) // rename başarılıysa zaten taşınmış olur

	w := csv.NewWriter(tmp)
	_ = w.Write([]string{"id", "order_id", "status", "error_msg", "attempted_at"})
	for _, j := range jobs {
		errMsg := ""
		if j.ErrorMsg.Valid {
			errMsg = j.ErrorMsg.String
		}
		_ = w.Write([]string{
			strconv.FormatInt(j.ID, 10),
			j.OrderID,
			j.Status,
			errMsg,
			j.AttemptedAt.In(e.loc).Format(time.RFC3339),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		tmp.Close()
		return "", 0, fmt.Errorf("csv yazımı: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", 0, fmt.Errorf("geçici dosya kapatma: %w", err)
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return "", 0, fmt.Errorf("atomik rename: %w", err)
	}

	if err := e.uploader.Upload(finalPath, name); err != nil {
		return finalPath, len(jobs), fmt.Errorf("cloud upload: %w", err)
	}
	return finalPath, len(jobs), nil
}

// RunDaily, her gün belirtilen saatte (lokal TZ) bir önceki tam günü export eder.
// Bloke eder; goroutine olarak çağrılmalı.
func (e *Exporter) RunDaily(hour int) {
	for {
		now := time.Now().In(e.loc)
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, e.loc)
		if !next.After(now) {
			next = next.AddDate(0, 0, 1)
		}
		time.Sleep(time.Until(next))

		yesterday := time.Now().In(e.loc).AddDate(0, 0, -1)
		path, n, err := e.ExportDay(yesterday)
		if err != nil {
			log.Printf("[EXPORT_HATASI] %v", err)
			continue
		}
		log.Printf("[EXPORT] %s yazıldı (%d kayıt)", path, n)
	}
}
