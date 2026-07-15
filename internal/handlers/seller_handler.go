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

type updateSellerRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Password *string `json:"password"`
	Status   *string `json:"status"`
}

func (h *SellerHandler) GetSeller(w http.ResponseWriter, r *http.Request) {

	sellerId := r.PathValue("id")
	if sellerId == "" {
		httputil.WriteError(w, http.StatusBadRequest, "Missing seller ID")
		return
	}

	var id int
	var name, email, phone, password, status string

	err := h.DB.QueryRow("SELECT id, name, email, phone, password, status FROM SELLERS WHERE ID = $1", sellerId).Scan(&id, &name, &email, &phone, &password, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			httputil.WriteError(w, http.StatusNotFound, "Seller not found")
		} else {
			httputil.WriteError(w, http.StatusInternalServerError, "Failed to retrieve seller "+err.Error())
		}
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":     id,
		"name":   name,
		"email":  email,
		"phone":  phone,
		"status": status,
	})
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

func (h *SellerHandler) UpdateSeller(w http.ResponseWriter, r *http.Request) {
	sellerId := r.PathValue("id")

	if sellerId == "" {
		httputil.WriteError(w, http.StatusBadRequest, "Missing seller ID")
		return
	}

	var req updateSellerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Name == nil && req.Email == nil && req.Phone == nil && req.Password == nil && req.Status == nil {
		httputil.WriteError(w, http.StatusBadRequest, "No fields to update")
		return
	}

	var exist bool
	err := h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM sellers WHERE id = $1)", sellerId).Scan(&exist)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to check existing seller "+err.Error())
		return
	}

	if !exist {
		httputil.WriteError(w, http.StatusNotFound, "Seller not found")
		return
	}

	if req.Email != nil || req.Phone != nil {
		email := ""
		phone := ""
		if req.Email != nil {
			email = *req.Email
		}
		if req.Phone != nil {
			phone = *req.Phone
		}

		var existMail, existPhone string
		err := h.DB.QueryRow("SELECT email, phone FROM sellers WHERE (email = $1 OR phone = $2) AND id != $3", email, phone, sellerId).Scan(&existMail, &existPhone)

		switch {
		case err == nil:
			if req.Email != nil && existMail == *req.Email {
				httputil.WriteError(w, http.StatusConflict, "Email already exists")
				return
			}
			if req.Phone != nil && existPhone == *req.Phone {
				httputil.WriteError(w, http.StatusConflict, "Phone already exists")
				return
			}
		case err == sql.ErrNoRows:
		default:
			httputil.WriteError(w, http.StatusInternalServerError, "Failed to check existing seller "+err.Error())
			return
		}
	}

	if req.Password != nil {
		passwordHash, err := auth.HashPassword(*req.Password)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "Failed to hash password "+err.Error())
			return
		}
		req.Password = &passwordHash
	}

	_, err = h.DB.Exec(`
		UPDATE sellers 
		SET 
			name = COALESCE($1, name), 
			email = COALESCE($2, email), 
			phone = COALESCE($3, phone), 
			password = COALESCE($4, password), 
			status = COALESCE($5, status),
			updated_at = NOW()
		WHERE id = $6`, req.Name, req.Email, req.Phone, req.Password, req.Status, sellerId,
	)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to update seller "+err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Seller updated successfully",
	})

}

func (h *SellerHandler) DeleteSeller(w http.ResponseWriter, r *http.Request) {

	sellerId := r.PathValue("id")
	if sellerId == "" {
		httputil.WriteError(w, http.StatusBadRequest, "Missing seller ID")
		return
	}

	result, err := h.DB.Exec("DELETE FROM SELLERS WHERE ID = $1", sellerId)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to delete seller"+err.Error())
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		httputil.WriteError(w, http.StatusNotFound, "Seller not found")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Seller deleted successfully",
	})

}
