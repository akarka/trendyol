package db

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/akarka/trendyol/internal/parser"
)

type User struct {
	ID           int    `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
	Role         string `db:"role"`
}

type OrderRow struct {
	UUID          string          `db:"uuid" json:"uuid"`
	OrderID       string          `db:"order_id" json:"order_id"`
	OrderNumber   string          `db:"order_number" json:"order_number"`
	PackageStatus string          `db:"package_status" json:"package_status"`
	Payload       json.RawMessage `db:"payload" json:"payload"`
	CreatedAt     time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at" json:"updated_at"`
}

type PrintJob struct {
	ID          int64          `db:"id" json:"id"`
	OrderID     string         `db:"order_id" json:"order_id"`
	Status      string         `db:"status" json:"status"`
	ErrorMsg    sql.NullString `db:"error_msg" json:"-"`
	AttemptedAt time.Time      `db:"attempted_at" json:"attempted_at"`
}

func (p PrintJob) MarshalJSON() ([]byte, error) {
	type alias PrintJob
	var errMsg *string
	if p.ErrorMsg.Valid {
		errMsg = &p.ErrorMsg.String
	}
	return json.Marshal(struct {
		alias
		ErrorMsg *string `json:"error_msg"`
	}{alias(p), errMsg})
}

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

func GetUserByUsername(db *sqlx.DB, username string) (*User, error) {
	var u User
	err := db.Get(&u, `SELECT id, username, password_hash, role FROM users WHERE username = ?`, username)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetOrders(db *sqlx.DB, limit, offset int, status string) ([]OrderRow, error) {
	orders := []OrderRow{}
	var err error
	if status != "" {
		err = db.Select(&orders,
			`SELECT uuid, order_id, order_number, package_status, payload, created_at, updated_at
			 FROM trendyol_orders WHERE package_status = ?
			 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
			status, limit, offset)
	} else {
		err = db.Select(&orders,
			`SELECT uuid, order_id, order_number, package_status, payload, created_at, updated_at
			 FROM trendyol_orders
			 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
			limit, offset)
	}
	return orders, err
}

func GetOrderByID(db *sqlx.DB, orderID string) (*OrderRow, error) {
	var o OrderRow
	err := db.Get(&o,
		`SELECT uuid, order_id, order_number, package_status, payload, created_at, updated_at
		 FROM trendyol_orders WHERE order_id = ?
		 ORDER BY created_at DESC LIMIT 1`,
		orderID)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func GetPrintJobs(db *sqlx.DB, limit int) ([]PrintJob, error) {
	jobs := []PrintJob{}
	err := db.Select(&jobs,
		`SELECT id, order_id, status, error_msg, attempted_at
		 FROM print_jobs ORDER BY attempted_at DESC LIMIT ?`,
		limit)
	return jobs, err
}

// GetPrintJobsBetween, [start, end) yarı-açık aralığındaki print job'ları kronolojik döner.
func GetPrintJobsBetween(db *sqlx.DB, start, end time.Time) ([]PrintJob, error) {
	jobs := []PrintJob{}
	err := db.Select(&jobs,
		`SELECT id, order_id, status, error_msg, attempted_at
		 FROM print_jobs WHERE attempted_at >= ? AND attempted_at < ?
		 ORDER BY attempted_at ASC`,
		start, end)
	return jobs, err
}

func GetSettings(db *sqlx.DB) (map[string]string, error) {
	rows, err := db.Queryx("SELECT `key`, value FROM settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, rows.Err()
}

func UpsertSetting(db *sqlx.DB, key, value string) error {
	_, err := db.Exec(
		"INSERT INTO settings (`key`, value) VALUES (?, ?) ON DUPLICATE KEY UPDATE value = VALUES(value)",
		key, value,
	)
	return err
}

func nullable(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
