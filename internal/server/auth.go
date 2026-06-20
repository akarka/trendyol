package server

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/akarka/trendyol/internal/auth"
	"github.com/akarka/trendyol/internal/db"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz istek gövdesi")
		return
	}

	user, err := db.GetUserByUsername(s.db, req.Username)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusUnauthorized, "kullanıcı adı veya şifre hatalı")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "sunucu hatası")
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "kullanıcı adı veya şifre hatalı")
		return
	}

	token, err := auth.GenerateToken(s.cfg.JWTSecret, user.ID, user.Username, user.Role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token üretilemedi")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token, "role": user.Role})
}
