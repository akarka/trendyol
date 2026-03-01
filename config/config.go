// config/config.go
package config

import (
	"log"
	"os"
)

// Config holds the application configuration
type Config struct {
	SupabaseURL     string
	SupabaseAnonKey string
	PrinterDevice   string
	LogLevel        string
}

// Load reads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		SupabaseURL:     os.Getenv("SUPABASE_URL"),
		SupabaseAnonKey: os.Getenv("SUPABASE_ANON_KEY"),
		PrinterDevice:   os.Getenv("PRINTER_DEVICE"),
		LogLevel:        os.Getenv("LOG_LEVEL"),
	}

	if cfg.SupabaseURL == "" {
		log.Fatal("SUPABASE_URL environment variable is mandatory")
	}
	if cfg.SupabaseAnonKey == "" {
		log.Fatal("SUPABASE_ANON_KEY environment variable is mandatory")
	}
	if cfg.PrinterDevice == "" {
		log.Fatal("PRINTER_DEVICE environment variable is mandatory")
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	return cfg
}
