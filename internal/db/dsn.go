package db

import (
	"net"

	"github.com/go-sql-driver/mysql"
)

// ConnInfo, mysqldump/mysql CLI çağırmak için DSN'den çıkarılan bağlantı bilgisi.
type ConnInfo struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// ParseConnInfo, MYSQL_DSN'i (go-sql-driver formatı) CLI argümanlarına çevirir.
func ParseConnInfo(dsn string) (*ConnInfo, error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}

	host, port := cfg.Addr, "3306"
	if h, p, splitErr := net.SplitHostPort(cfg.Addr); splitErr == nil {
		host, port = h, p
	}

	return &ConnInfo{
		Host:     host,
		Port:     port,
		User:     cfg.User,
		Password: cfg.Passwd,
		DBName:   cfg.DBName,
	}, nil
}
