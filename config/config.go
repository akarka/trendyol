package config

import (
	"os"
)

type Config struct {
	SupabaseURL     string
	SupabaseAnonKey string
	PrinterDevice   string
	LogLevel        string
}

func Load() *Config {
	cfg := &Config{
		SupabaseURL:     os.Getenv("SUPABASE_URL"),
		SupabaseAnonKey: os.Getenv("SUPABASE_ANON_KEY"),
		PrinterDevice:   os.Getenv("PRINTER_DEVICE"),
		LogLevel:        os.Getenv("LOG_LEVEL"),
	}

	if cfg.SupabaseURL == "" || cfg.SupabaseAnonKey == "" || cfg.PrinterDevice == "" {
		panic("Zorunlu ortam değişkenleri eksik: SUPABASE_URL, SUPABASE_ANON_KEY veya PRINTER_DEVICE")
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	return cfg
}
