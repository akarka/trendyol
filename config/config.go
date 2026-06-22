package config

import (
	"os"
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

	return cfg
}
