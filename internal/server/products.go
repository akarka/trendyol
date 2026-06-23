package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-sql-driver/mysql"

	"github.com/akarka/trendyol/internal/db"
)

func (s *Server) handleListCategories(w http.ResponseWriter, r *http.Request) {
	names, err := db.GetCategoryNames(s.db)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "kategoriler okunamadı")
		return
	}
	writeJSON(w, http.StatusOK, names)
}

func (s *Server) handleListBrands(w http.ResponseWriter, r *http.Request) {
	names, err := db.GetBrandNames(s.db)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "markalar okunamadı")
		return
	}
	writeJSON(w, http.StatusOK, names)
}

type productReq struct {
	SKU             string   `json:"sku"`
	Barcode         string   `json:"barcode"`
	Name            string   `json:"name"`
	MarketplaceName string   `json:"marketplace_name"`
	Category        string   `json:"category"`
	Brand           string   `json:"brand"`
	NetWeight       *float64 `json:"net_weight"`
	Unit            string   `json:"unit"`
	Price           float64  `json:"price"`
	VATRate         *float64 `json:"vat_rate"`
	IsActive        bool     `json:"is_active"`
	Description     string   `json:"description"`
}

// toProduct, category/brand adlarını upsert ederek db.Product'a çevirir.
func (s *Server) toProduct(req productReq) (db.Product, error) {
	catID, err := db.UpsertLookup(s.db, "categories", strings.TrimSpace(req.Category))
	if err != nil {
		return db.Product{}, err
	}
	var brandID sql.NullInt64
	if b := strings.TrimSpace(req.Brand); b != "" {
		id, err := db.UpsertLookup(s.db, "brands", b)
		if err != nil {
			return db.Product{}, err
		}
		brandID = sql.NullInt64{Int64: id, Valid: true}
	}

	p := db.Product{
		SKU:         strings.TrimSpace(req.SKU),
		Barcode:     strings.TrimSpace(req.Barcode),
		Name:        strings.TrimSpace(req.Name),
		CategoryID:  int(catID),
		BrandID:     brandID,
		Unit:        strOrNull(req.Unit),
		Price:       req.Price,
		IsActive:    req.IsActive,
		Description: strOrNull(req.Description),
	}
	if req.MarketplaceName != "" {
		p.MarketplaceName = sql.NullString{String: req.MarketplaceName, Valid: true}
	}
	if req.NetWeight != nil {
		p.NetWeight = sql.NullFloat64{Float64: *req.NetWeight, Valid: true}
	}
	if req.VATRate != nil {
		p.VATRate = sql.NullFloat64{Float64: *req.VATRate, Valid: true}
	}
	return p, nil
}

func strOrNull(s string) sql.NullString {
	s = strings.TrimSpace(s)
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func (s *Server) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	var req productReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz istek gövdesi")
		return
	}
	if strings.TrimSpace(req.SKU) == "" || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Category) == "" {
		writeError(w, http.StatusBadRequest, "sku, name ve category zorunlu")
		return
	}

	p, err := s.toProduct(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "kategori/marka oluşturulamadı")
		return
	}

	if err := db.CreateProduct(s.db, p); err != nil {
		if isDupEntry(err) {
			writeError(w, http.StatusConflict, "bu SKU veya barkod zaten kayıtlı")
			return
		}
		writeError(w, http.StatusInternalServerError, "ürün oluşturulamadı")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"sku": p.SKU})
}

func (s *Server) handleUpdateProduct(w http.ResponseWriter, r *http.Request) {
	sku := chi.URLParam(r, "sku")
	var req productReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz istek gövdesi")
		return
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Category) == "" {
		writeError(w, http.StatusBadRequest, "name ve category zorunlu")
		return
	}

	p, err := s.toProduct(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "kategori/marka oluşturulamadı")
		return
	}

	if err := db.UpdateProduct(s.db, sku, p); err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "ürün bulunamadı")
		return
	} else if err != nil {
		if isDupEntry(err) {
			writeError(w, http.StatusConflict, "bu barkod zaten kayıtlı")
			return
		}
		writeError(w, http.StatusInternalServerError, "ürün güncellenemedi")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"sku": sku})
}

func (s *Server) handleDeleteProduct(w http.ResponseWriter, r *http.Request) {
	sku := chi.URLParam(r, "sku")
	if err := db.DeleteProduct(s.db, sku); err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "ürün bulunamadı")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "ürün silinemedi")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func isDupEntry(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
