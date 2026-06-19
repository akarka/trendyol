package db

import (
	"github.com/jmoiron/sqlx"

	"github.com/akarka/trendyol/internal/parser"
)

// InsertOrder, ham Trendyol payload'ını trendyol_orders'a yazar.
// UNIQUE(order_id, package_status) sayesinde duplicate'ler yok sayılır;
// yeni kayıt eklendiyse inserted=true döner.
func InsertOrder(db *sqlx.DB, o *parser.Order, payload []byte) (bool, error) {
	res, err := db.Exec(
		`INSERT IGNORE INTO trendyol_orders (order_id, order_number, package_status, payload)
		 VALUES (?, ?, ?, ?)`,
		o.OrderID, o.OrderNumber, o.PackageStatus, string(payload),
	)
	if err != nil {
		return false, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func InsertPrintJob(db *sqlx.DB, orderID, status, errMsg string) (int64, error) {
	res, err := db.Exec(
		`INSERT INTO print_jobs (order_id, status, error_msg) VALUES (?, ?, ?)`,
		orderID, status, nullable(errMsg),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdatePrintJob(db *sqlx.DB, id int64, status, errMsg string) error {
	_, err := db.Exec(
		`UPDATE print_jobs SET status = ?, error_msg = ?, attempted_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, nullable(errMsg), id,
	)
	return err
}

func nullable(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
