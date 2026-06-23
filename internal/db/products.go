package db

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// Product, import sırasında products tablosuna yazılan kayıt.
type Product struct {
	SKU             string
	Barcode         string
	Name            string
	MarketplaceName sql.NullString
	CategoryID      int
	BrandID         sql.NullInt64
	NetWeight       sql.NullFloat64
	Unit            sql.NullString
	Price           float64
	VATRate         sql.NullFloat64
	IsActive        bool
	NeedsFix        bool
	Description     sql.NullString
}

// ProductView, okuma (API dropdown + export) için kategori/marka adıyla join'li görünüm.
type ProductView struct {
	SKU             string   `db:"sku" json:"sku"`
	Barcode         string   `db:"barcode" json:"barcode"`
	Name            string   `db:"name" json:"name"`
	MarketplaceName *string  `db:"marketplace_name" json:"marketplace_name"`
	Category        string   `db:"category" json:"category"`
	Brand           *string  `db:"brand" json:"brand"`
	NetWeight       *float64 `db:"net_weight" json:"net_weight"`
	Unit            *string  `db:"unit" json:"unit"`
	Price           float64  `db:"price" json:"price"`
	VATRate         *float64 `db:"vat_rate" json:"vat_rate"`
	IsActive        bool     `db:"is_active" json:"is_active"`
	NeedsFix        bool     `db:"needs_fix" json:"needs_fix"`
	Description     *string  `db:"description" json:"description"`
}

// UpsertLookup, categories/brands gibi (id, name UNIQUE) tablolara idempotent insert eder, id döner.
func UpsertLookup(db *sqlx.DB, table, name string) (int64, error) {
	res, err := db.Exec(
		"INSERT INTO "+table+" (name) VALUES (?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id)",
		name,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ProductExists, sku zaten products tablosunda var mı diye bakar (mükerrer atlama importu için).
func ProductExists(db *sqlx.DB, sku string) (bool, error) {
	var exists bool
	err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM products WHERE sku = ?)", sku)
	return exists, err
}

func UpsertProduct(db *sqlx.DB, p Product) error {
	_, err := db.Exec(
		`INSERT INTO products
		   (sku, barcode, name, marketplace_name, category_id, brand_id, net_weight, unit, price, vat_rate, is_active, needs_fix, description)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   barcode=VALUES(barcode), name=VALUES(name), marketplace_name=VALUES(marketplace_name),
		   category_id=VALUES(category_id), brand_id=VALUES(brand_id), net_weight=VALUES(net_weight),
		   unit=VALUES(unit), price=VALUES(price), vat_rate=VALUES(vat_rate),
		   is_active=VALUES(is_active), needs_fix=VALUES(needs_fix), description=VALUES(description)`,
		p.SKU, p.Barcode, p.Name, p.MarketplaceName, p.CategoryID, p.BrandID, p.NetWeight, p.Unit,
		p.Price, p.VATRate, p.IsActive, p.NeedsFix, p.Description,
	)
	return err
}

// GetProducts, join'li ürün listesini döner. activeOnly=true ise sadece aktif ürünler.
func GetProducts(db *sqlx.DB, activeOnly bool) ([]ProductView, error) {
	q := `SELECT p.sku, p.barcode, p.name, p.marketplace_name, c.name AS category, b.name AS brand,
	             p.net_weight, p.unit, p.price, p.vat_rate, p.is_active, p.needs_fix, p.description
	      FROM products p
	      JOIN categories c ON c.id = p.category_id
	      LEFT JOIN brands b ON b.id = p.brand_id`
	if activeOnly {
		q += " WHERE p.is_active = 1"
	}
	q += " ORDER BY c.name, p.name"

	out := []ProductView{}
	err := db.Select(&out, q)
	return out, err
}

// GetProductBySKU, tek ürünü döner (manuel sipariş satırı zenginleştirme).
func GetProductBySKU(db *sqlx.DB, sku string) (*ProductView, error) {
	var p ProductView
	err := db.Get(&p,
		`SELECT p.sku, p.barcode, p.name, p.marketplace_name, c.name AS category, b.name AS brand,
		        p.net_weight, p.unit, p.price, p.vat_rate, p.is_active, p.needs_fix, p.description
		 FROM products p
		 JOIN categories c ON c.id = p.category_id
		 LEFT JOIN brands b ON b.id = p.brand_id
		 WHERE p.sku = ?`, sku)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
