package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"

	"github.com/akarka/trendyol/config"
	"github.com/akarka/trendyol/internal/parser"
)

// PrintTask, webhook handler'dan printer goroutine'ine geçen iş birimidir.
type PrintTask struct {
	Order *parser.Order
	JobID int64
}

type Server struct {
	cfg     *config.Config
	db      *sqlx.DB
	printCh chan<- PrintTask
	router  *chi.Mux
}

func New(cfg *config.Config, db *sqlx.DB, printCh chan<- PrintTask) *Server {
	s := &Server{cfg: cfg, db: db, printCh: printCh}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Group(func(r chi.Router) {
		r.Use(basicAuth(cfg.WebhookUsername, cfg.WebhookPassword))
		r.Post("/webhook/trendyol", s.handleWebhook)
	})

	s.router = r
	return s
}

func (s *Server) Start(addr string) error {
	log.Printf("HTTP sunucusu %s adresinde dinleniyor", addr)
	return http.ListenAndServe(addr, s.router)
}
