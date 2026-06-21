package db

import "github.com/jmoiron/sqlx"

// EnsureCatalogTables, ürün kataloğu tablolarını yoksa oluşturur.
// docs/schema.sql ile birebir aynı; mysql_data volume'ü kalıcı olduğundan init script
// mevcut DB'de yeniden çalışmaz — bu fonksiyon tabloların varlığını her açılışta garanti eder.
func EnsureCatalogTables(db *sqlx.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS categories (
		  id   INT AUTO_INCREMENT PRIMARY KEY,
		  name VARCHAR(64) NOT NULL UNIQUE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS brands (
		  id   INT AUTO_INCREMENT PRIMARY KEY,
		  name VARCHAR(64) NOT NULL UNIQUE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS products (
		  sku              VARCHAR(40)   PRIMARY KEY,
		  barcode          VARCHAR(40)   NOT NULL,
		  name             VARCHAR(255)  NOT NULL,
		  marketplace_name VARCHAR(255),
		  category_id      INT           NOT NULL,
		  brand_id         INT,
		  net_weight       DECIMAL(10,3),
		  unit             VARCHAR(8),
		  price            DECIMAL(10,2) NOT NULL,
		  vat_rate         DECIMAL(5,2),
		  is_active        BOOLEAN       NOT NULL DEFAULT 0,
		  needs_fix        BOOLEAN       NOT NULL DEFAULT 0,
		  description      TEXT,
		  created_at       DATETIME      DEFAULT CURRENT_TIMESTAMP,
		  updated_at       DATETIME      DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		  UNIQUE KEY uq_barcode (barcode),
		  CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES categories(id),
		  CONSTRAINT fk_products_brand    FOREIGN KEY (brand_id)    REFERENCES brands(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
