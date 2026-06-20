package main

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/akarka/trendyol/config"
	"github.com/akarka/trendyol/internal/db"
	"github.com/akarka/trendyol/internal/printer"
	"github.com/akarka/trendyol/internal/server"
	"github.com/akarka/trendyol/web"
)

func main() {
	cfg := config.Load()

	// depends_on MySQL'in hazır olmasını garanti etmez; bir süre yeniden dene.
	var database *sqlx.DB
	var err error
	for i := 0; i < 30; i++ {
		database, err = db.Connect(cfg.MySQLDSN)
		if err == nil {
			break
		}
		log.Printf("[DB_BEKLE] MySQL'e bağlanılamadı (deneme %d): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("[DB_HATASI] MySQL'e bağlanılamadı: %v", err)
	}
	defer database.Close()

	printCh := make(chan server.PrintTask, 64)
	go runPrinter(printCh, database, cfg)

	srv := server.New(cfg, database, printCh, web.Dist())
	if err := srv.Start(":8080"); err != nil {
		log.Fatalf("[SUNUCU_HATASI] %v", err)
	}
}

func runPrinter(printCh <-chan server.PrintTask, database *sqlx.DB, cfg *config.Config) {
	for task := range printCh {
		var perr error
		if cfg.TestMode {
			perr = printer.PrintToTXT(task.Order)
		} else {
			perr = printer.Print(cfg.PrinterDevice, task.Order)
		}

		if perr != nil {
			log.Printf("[YAZDIRMA_HATASI] sipariş=%s hata=%v", task.Order.OrderNumber, perr)
			_ = db.UpdatePrintJob(database, task.JobID, "failed", perr.Error())
			continue
		}

		if err := db.UpdatePrintJob(database, task.JobID, "success", ""); err != nil {
			log.Printf("[DB_HATASI] print job güncellenemedi id=%d: %v", task.JobID, err)
		}
	}
}
