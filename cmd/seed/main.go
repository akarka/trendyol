package main

import (
	"flag"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/akarka/trendyol/internal/db"
)

func main() {
	username := flag.String("username", "", "kullanıcı adı")
	password := flag.String("password", "", "şifre")
	role := flag.String("role", "admin", "rol")
	flag.Parse()

	if *username == "" || *password == "" {
		log.Fatal("kullanım: go run ./cmd/seed --username <user> --password <pass>")
	}

	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("MYSQL_DSN ortam değişkeni gerekli")
	}

	conn, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("DB bağlantısı kurulamadı: %v", err)
	}
	defer conn.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("bcrypt hash üretilemedi: %v", err)
	}

	_, err = conn.Exec(
		`INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE password_hash = VALUES(password_hash), role = VALUES(role)`,
		*username, string(hash), *role,
	)
	if err != nil {
		log.Fatalf("kullanıcı kaydedilemedi: %v", err)
	}

	log.Printf("kullanıcı oluşturuldu/güncellendi: %s (%s)", *username, *role)
}
