package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"rekber-fesnuk/internal/auth"
	"rekber-fesnuk/internal/httputil"
)

type SellerHandler struct {
	DB *sql.DB
}

func NewSellerHandler(db *sql.DB) *SellerHandler {
	return &SellerHandler{DB: db}
}

type createSellerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

func (h *SellerHandler) CreateSeller(w http.ResponseWriter, r *http.Request) {
	var req createSellerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Name == "" || req.Email == "" || req.Phone == "" || req.Password == "" {
		httputil.WriteError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	var existMail, existPhone string
	err := h.DB.QueryRow("SELECT email, phone FROM sellers where email = $1 or phone = $2", req.Email, req.Phone).Scan(&existMail, &existPhone)

	switch {
	case err == nil:
		if existMail == req.Email {
			httputil.WriteError(w, http.StatusConflict, "Email already exists")
			return
		}
		if existPhone == req.Phone {
			httputil.WriteError(w, http.StatusConflict, "Phone already exists")
			return
		}
	case err == sql.ErrNoRows:
	default:
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to check existing seller "+err.Error())
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to hash password "+err.Error())
		return
	}
	
	var id int
	err = h.DB.QueryRow("INSERT INTO sellers (name, email, phone, password, status) VALUES ($1, $2, $3, $4, $5) RETURNING id", req.Name, req.Email, req.Phone, passwordHash, "inactive").Scan(&id)

	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to create seller "+err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id": id,
	})
}
