package config

import (
	"os"
	"strconv"
)

type Config struct {
	MySQLDSN        string
	JWTSecret       string
	WebhookUsername string
	WebhookPassword string
	PrinterDevice   string
	LogLevel        string
	TestMode        bool
	RBACEnabled     bool

	ExportEnabled bool
	ExportDir     string
	ExportHour    int    // günlük export'un tetiklendiği saat (0-23, lokal TZ)
	ExportTZ      string // gün sınırlarının hesaplandığı saat dilimi
}

func Load() *Config {
	cfg := &Config{
		MySQLDSN:        os.Getenv("MYSQL_DSN"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		WebhookUsername: os.Getenv("WEBHOOK_USERNAME"),
		WebhookPassword: os.Getenv("WEBHOOK_PASSWORD"),
		PrinterDevice:   os.Getenv("PRINTER_DEVICE"),
		LogLevel:        os.Getenv("LOG_LEVEL"),
		TestMode:        os.Getenv("TEST_MODE") == "true",
		RBACEnabled:     os.Getenv("RBAC_ENABLED") != "false", // set değilse açık; sadece "false" kapatır

		ExportEnabled: os.Getenv("EXPORT_ENABLED") != "false", // set değilse açık
		ExportDir:     getenvDefault("EXPORT_DIR", "exports"),
		ExportHour:    atoiDefault(os.Getenv("EXPORT_HOUR"), 0), // varsayılan gece yarısı: bir önceki tam günü export eder
		ExportTZ:      getenvDefault("EXPORT_TZ", "Europe/Istanbul"),
	}

	if cfg.MySQLDSN == "" || cfg.JWTSecret == "" {
		panic("Zorunlu ortam değişkenleri eksik: MYSQL_DSN, JWT_SECRET")
	}

	if cfg.WebhookUsername == "" || cfg.WebhookPassword == "" {
		panic("Zorunlu ortam değişkenleri eksik: WEBHOOK_USERNAME, WEBHOOK_PASSWORD")
	}

	if !cfg.TestMode && cfg.PrinterDevice == "" {
		panic("TEST_MODE aktif değilken PRINTER_DEVICE zorunludur.")
	}

	if cfg.TestMode && cfg.PrinterDevice == "" {
		cfg.PrinterDevice = "output.txt"
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	if cfg.ExportHour < 0 || cfg.ExportHour > 23 {
		cfg.ExportHour = 0
	}

	return cfg
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func atoiDefault(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}
