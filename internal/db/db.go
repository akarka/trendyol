package db

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql"
)

func Connect(dsn string) (*sqlx.DB, error) {
	conn, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql bağlantısı kurulamadı: %w", err)
	}

	conn.SetConnMaxLifetime(5 * time.Minute)
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("mysql ping başarısız: %w", err)
	}

	return conn, nil
}
