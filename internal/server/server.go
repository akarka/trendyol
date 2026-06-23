package server

import (
	"io/fs"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"

	"github.com/akarka/trendyol/config"
	"github.com/akarka/trendyol/internal/auth"
	"github.com/akarka/trendyol/internal/exporter"
	"github.com/akarka/trendyol/internal/parser"
)

// PrintTask, webhook handler'dan printer goroutine'ine geçen iş birimidir.
type PrintTask struct {
	Order *parser.Order
	JobID int64
}

type Server struct {
	cfg      *config.Config
	db       *sqlx.DB
	printCh  chan<- PrintTask
	exporter *exporter.Exporter
	router   *chi.Mux
}

func New(cfg *config.Config, db *sqlx.DB, printCh chan<- PrintTask, exp *exporter.Exporter, static fs.FS) *Server {
	s := &Server{cfg: cfg, db: db, printCh: printCh, exporter: exp}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Group(func(r chi.Router) {
		r.Use(basicAuth(cfg.WebhookUsername, cfg.WebhookPassword))
		r.Post("/webhook/trendyol", s.handleWebhook)
	})

	r.Post("/api/auth/login", s.handleLogin)
	r.Get("/api/config", s.handleConfig)

	r.Group(func(r chi.Router) {
		if cfg.RBACEnabled {
			r.Use(auth.JWTMiddleware(cfg.JWTSecret))
		}
		r.Get("/api/orders", s.handleListOrders)
		r.Post("/api/orders/manual", s.handleManualOrder)
		r.Get("/api/orders/{orderID}", s.handleGetOrder)
		r.Post("/api/orders/{orderID}/print", s.handleReprint)
		r.Get("/api/products", s.handleListProducts)
		r.Get("/api/printer/status", s.handlePrinterStatus)
		r.Get("/api/logs", s.handleLogs)
		r.Post("/api/export/print-jobs", s.handleExportPrintJobs)
		r.Get("/api/settings", s.handleGetSettings)
		r.Put("/api/settings/{key}", s.handlePutSetting)
	})

	r.Handle("/*", spaHandler(static))

	s.router = r
	return s
}

// spaHandler, embed edilmiş statik dosyaları sunar; bilinmeyen rotalarda
// React Router'ın devralması için index.html'e düşer.
func spaHandler(static fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(static))
	return func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if p == "" {
			p = "index.html"
		}
		if _, err := fs.Stat(static, p); err != nil {
			b, rerr := fs.ReadFile(static, "index.html")
			if rerr != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(b)
			return
		}
		fileServer.ServeHTTP(w, r)
	}
}

func (s *Server) Start(addr string) error {
	log.Printf("HTTP sunucusu %s adresinde dinleniyor", addr)
	return http.ListenAndServe(addr, s.router)
}
