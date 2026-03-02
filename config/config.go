package config

import (
	"os"
)

type Config struct {
	SupabaseURL     string
	SupabaseAnonKey string
	PrinterDevice   string
	LogLevel        string
	TestMode        bool
}

func Load() *Config {
	cfg := &Config{
		SupabaseURL:     os.Getenv("SUPABASE_URL"),
		SupabaseAnonKey: os.Getenv("SUPABASE_ANON_KEY"),
		PrinterDevice:   os.Getenv("PRINTER_DEVICE"),
		LogLevel:        os.Getenv("LOG_LEVEL"),
		TestMode:        os.Getenv("TEST_MODE") == "true",
	}

	// In TestMode, PrinterDevice is not mandatory as it's a file path.
	if cfg.SupabaseURL == "" || cfg.SupabaseAnonKey == "" {
		panic("Zorunlu ortam değişkenleri eksik: SUPABASE_URL, SUPABASE_ANON_KEY")
	}

	if !cfg.TestMode && cfg.PrinterDevice == "" {
		panic("TEST_MODE aktif değilken PRINTER_DEVICE zorunludur.")
	}
	
	if cfg.TestMode && cfg.PrinterDevice == "" {
		// Default file path for test mode if not provided
		cfg.PrinterDevice = "orders.txt"
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	return cfg
}
