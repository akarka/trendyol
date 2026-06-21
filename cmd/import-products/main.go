// import-products, Zeytuni ürün listesini (CSV veya .xlsx) normalize edip products tablosuna yazar.
// Idempotent: SKU üzerinden ON DUPLICATE KEY UPDATE. Tekrar çalıştırınca günceller.
//
//	go run ./cmd/import-products --file data/urun_listesi.utf8.csv
package main

import (
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/akarka/trendyol/internal/db"
	"github.com/akarka/trendyol/internal/xlsxlite"
)

func main() {
	file := flag.String("file", "data/urun_listesi.utf8.csv", "kaynak dosya (.csv veya .xlsx)")
	flag.Parse()

	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("MYSQL_DSN ortam değişkeni gerekli")
	}

	rows, err := readRows(*file)
	if err != nil {
		log.Fatalf("dosya okunamadı: %v", err)
	}
	if len(rows) < 2 {
		log.Fatal("dosyada veri yok")
	}

	conn, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("DB bağlantısı kurulamadı: %v", err)
	}
	defer conn.Close()

	if err := db.EnsureCatalogTables(conn); err != nil {
		log.Fatalf("katalog tabloları oluşturulamadı: %v", err)
	}

	col := headerIndex(rows[0])
	required := []string{"SKU", "Barkod", "Kategori", "Ürün", "Fiziki Magaza Fiyati", "Aktif/Pasif"}
	for _, h := range required {
		if _, ok := col[h]; !ok {
			log.Fatalf("zorunlu kolon eksik: %q (başlık: %v)", h, rows[0])
		}
	}

	catCache := map[string]int64{}
	brandCache := map[string]int64{}
	seenSKU := map[string]int{}
	seenBarcode := map[string]int{}

	var imported, flagged int
	var issues []string

	for i, row := range rows[1:] {
		get := func(h string) string {
			if idx, ok := col[h]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		sku := get("SKU")
		if sku == "" {
			continue
		}
		needsFix := false

		barcode := get("Barkod")
		name := get("Ürün")
		mpName := get("Ty Ürün Adi")
		category := get("Kategori")
		brand := get("Marka")
		desc := get("Marketplace Açiklama")

		// çift SKU/barkod → ilkini koru, sonrakileri suffix + needs_fix
		if n := seenSKU[sku]; n > 0 {
			needsFix = true
			issues = append(issues, fmt.Sprintf("satır %d: çift SKU %s → %s-DUP%d", i+2, sku, sku, n+1))
			sku = fmt.Sprintf("%s-DUP%d", sku, n+1)
		}
		seenSKU[get("SKU")]++
		if n := seenBarcode[barcode]; n > 0 {
			needsFix = true
			barcode = fmt.Sprintf("%s-DUP%d", barcode, n+1)
		}
		seenBarcode[get("Barkod")]++

		price, perr := parseNumber(get("Fiziki Magaza Fiyati"))
		if perr != nil {
			issues = append(issues, fmt.Sprintf("satır %d: fiyat okunamadı %q", i+2, get("Fiziki Magaza Fiyati")))
			needsFix = true
		}

		weight, unit, wok := resolveWeight(get, name, mpName)
		if !wok {
			needsFix = true
			issues = append(issues, fmt.Sprintf("satır %d: gramaj eksik/0 (%s)", i+2, name))
		}

		catID, err := lookup(conn, catCache, "categories", category)
		if err != nil {
			log.Fatalf("kategori upsert hatası (%s): %v", category, err)
		}
		var brandID sql.NullInt64
		if brand != "" {
			id, err := lookup(conn, brandCache, "brands", brand)
			if err != nil {
				log.Fatalf("marka upsert hatası (%s): %v", brand, err)
			}
			brandID = sql.NullInt64{Int64: id, Valid: true}
		}

		p := db.Product{
			SKU:             sku,
			Barcode:         barcode,
			Name:            name,
			MarketplaceName: nullStr(mpName),
			CategoryID:      int(catID),
			BrandID:         brandID,
			NetWeight:       nullFloat(weight, wok),
			Unit:            nullStr(unit),
			Price:           price,
			VATRate:         parseVAT(get("KDV")),
			IsActive:        strings.EqualFold(get("Aktif/Pasif"), "AKTIF"),
			NeedsFix:        needsFix,
			Description:     nullStr(desc),
		}
		if err := db.UpsertProduct(conn, p); err != nil {
			log.Fatalf("ürün yazılamadı (%s): %v", sku, err)
		}
		imported++
		if needsFix {
			flagged++
		}
	}

	log.Printf("import tamam: %d ürün (%d kategori, %d marka). needs_fix: %d", imported, len(catCache), len(brandCache), flagged)
	if len(issues) > 0 {
		log.Printf("--- işaretlenen sorunlar (%d) ---", len(issues))
		for _, s := range issues {
			log.Println(s)
		}
	}
}

func readRows(path string) ([][]string, error) {
	if strings.EqualFold(filepath.Ext(path), ".xlsx") {
		return xlsxlite.Read(path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	return r.ReadAll()
}

func headerIndex(header []string) map[string]int {
	m := map[string]int{}
	for i, h := range header {
		m[strings.TrimSpace(h)] = i
	}
	return m
}

func lookup(conn *sqlx.DB, cache map[string]int64, table, name string) (int64, error) {
	if id, ok := cache[name]; ok {
		return id, nil
	}
	id, err := db.UpsertLookup(conn, table, name)
	if err != nil {
		return 0, err
	}
	cache[name] = id
	return id, nil
}

var weightRe = regexp.MustCompile(`(?i)(\d+(?:[.,]\d+)?)\s*(kg|gr|g|ml|lt|l|adet|kase)\b`)

// resolveWeight, önce açık "Net Gramaj"/"Birim" kolonlarını, yoksa ürün adından çözer.
// ok=false → gramaj eksik veya 0 (needs_fix).
func resolveWeight(get func(string) string, names ...string) (float64, string, bool) {
	if g := get("Net Gramaj"); g != "" {
		v, err := parseNumber(g)
		if err == nil && v > 0 {
			unit := strings.ToLower(get("Birim"))
			if unit == "" {
				unit = "g"
			}
			return v, unit, true
		}
	}
	for _, n := range names {
		if v, u, ok := parseWeight(n); ok {
			return v, u, true
		}
	}
	return 0, "", false
}

func parseWeight(s string) (float64, string, bool) {
	matches := weightRe.FindAllStringSubmatch(s, -1)
	var fallback []string // adet/kase
	for _, m := range matches {
		v, err := parseNumber(m[1])
		if err != nil || v == 0 {
			continue
		}
		unit := strings.ToLower(m[2])
		switch unit {
		case "gr":
			unit = "g"
		case "lt":
			unit = "l"
		}
		if unit == "adet" || unit == "kase" {
			fallback = []string{strconv.FormatFloat(v, 'f', -1, 64), unit}
			continue
		}
		return v, unit, true
	}
	if fallback != nil {
		v, _ := strconv.ParseFloat(fallback[0], 64)
		return v, fallback[1], true
	}
	return 0, "", false
}

func parseNumber(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.NewReplacer(" ", "", "₺", "", "?", "", "TL", "").Replace(s)
	// binlik ayraç (virgül) temizle; ondalık nokta kalır
	if strings.Count(s, ",") > 0 && strings.Contains(s, ".") {
		s = strings.ReplaceAll(s, ",", "")
	} else {
		s = strings.ReplaceAll(s, ",", ".")
	}
	return strconv.ParseFloat(s, 64)
}

func parseVAT(s string) sql.NullFloat64 {
	s = strings.TrimSpace(strings.ReplaceAll(s, "%", ""))
	if s == "" {
		return sql.NullFloat64{}
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: v, Valid: true}
}

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullFloat(v float64, ok bool) sql.NullFloat64 {
	if !ok {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: v, Valid: true}
}
