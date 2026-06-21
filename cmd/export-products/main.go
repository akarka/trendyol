// export-products, products tablosunu işletmenin düzelteceği bir .xlsx olarak yazar.
// Başlıklar import-products ile uyumludur: düzeltilen dosya doğrudan re-import edilebilir.
//
//	go run ./cmd/export-products --out urunler.xlsx
package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/akarka/trendyol/internal/db"
	"github.com/akarka/trendyol/internal/xlsxlite"
)

var headers = []string{
	"SKU", "Barkod", "Kategori", "Ürün", "Marka", "Ty Ürün Adi",
	"Fiziki Magaza Fiyati", "KDV", "Aktif/Pasif", "Net Gramaj", "Birim", "Needs Fix", "Marketplace Açiklama",
}

func main() {
	out := flag.String("out", "urunler.xlsx", "çıktı .xlsx yolu")
	flag.Parse()

	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("MYSQL_DSN ortam değişkeni gerekli")
	}

	conn, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("DB bağlantısı kurulamadı: %v", err)
	}
	defer conn.Close()

	if err := db.EnsureCatalogTables(conn); err != nil {
		log.Fatalf("katalog tabloları oluşturulamadı: %v", err)
	}

	products, err := db.GetProducts(conn, false)
	if err != nil {
		log.Fatalf("ürünler okunamadı: %v", err)
	}

	rows := [][]string{headers}
	for _, p := range products {
		rows = append(rows, []string{
			p.SKU,
			p.Barcode,
			p.Category,
			p.Name,
			deref(p.Brand),
			deref(p.MarketplaceName),
			strconv.FormatFloat(p.Price, 'f', 2, 64),
			fnum(p.VATRate),
			activeStr(p.IsActive),
			fnum(p.NetWeight),
			deref(p.Unit),
			fixStr(p.NeedsFix),
			deref(p.Description),
		})
	}

	if err := xlsxlite.Write(*out, "Ürünler", rows); err != nil {
		log.Fatalf("xlsx yazılamadı: %v", err)
	}
	log.Printf("export tamam: %d ürün → %s", len(products), *out)
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func fnum(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', -1, 64)
}

func activeStr(b bool) string {
	if b {
		return "AKTIF"
	}
	return "PASIF"
}

func fixStr(b bool) string {
	if b {
		return "EVET"
	}
	return ""
}
