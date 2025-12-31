// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"boop-cat/db"
	"boop-cat/middleware"
)

type AccountHandler struct {
	DB *sql.DB
}

func NewAccountHandler(database *sql.DB) *AccountHandler {
	return &AccountHandler{DB: database}
}

func (h *AccountHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireLogin)

	r.Get("/linked-accounts", h.ListLinkedAccounts)
	r.Delete("/linked-accounts/{id}", h.UnlinkAccount)
	r.Post("/email", h.ChangeEmail)
	r.Post("/password", h.ChangePassword)

	return r
}

func (h *AccountHandler) ListLinkedAccounts(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	rows, err := h.DB.Query(`
		SELECT id, provider, displayName, createdAt
		FROM oauthAccounts WHERE userId = ?
	`, userID)
	if err != nil {
		jsonError(w, "db-error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var accounts []map[string]interface{}
	for rows.Next() {
		var id, provider string
		var displayName, createdAt sql.NullString
		if err := rows.Scan(&id, &provider, &displayName, &createdAt); err != nil {
			continue
		}
		acc := map[string]interface{}{
			"id":       id,
			"provider": provider,
		}
		if displayName.Valid {
			acc["displayName"] = displayName.String
		}
		if createdAt.Valid {
			acc["createdAt"] = createdAt.String
		}
		accounts = append(accounts, acc)
	}

	if accounts == nil {
		accounts = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"accounts": accounts,
	})
}

func (h *AccountHandler) UnlinkAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	accountID := chi.URLParam(r, "id")

	user, err := db.GetUserByID(h.DB, userID)
	if err != nil {
		jsonError(w, "user-not-found", http.StatusNotFound)
		return
	}

	var count int
	h.DB.QueryRow(`SELECT COUNT(*) FROM oauthAccounts WHERE userId = ?`, userID).Scan(&count)

	hasPassword := user.PasswordHash.Valid && user.PasswordHash.String != ""

	if count <= 1 && !hasPassword {
		jsonError(w, "cannot-unlink-only-auth", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec(`DELETE FROM oauthAccounts WHERE id = ? AND userId = ?`, accountID, userID)
	if err != nil {
		jsonError(w, "db-error", http.StatusInternalServerError)
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		jsonError(w, "account-not-found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func (h *AccountHandler) ChangeEmail(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req struct {
		NewEmail        string `json:"newEmail"`
		CurrentPassword string `json:"currentPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid-json", http.StatusBadRequest)
		return
	}

	if req.NewEmail == "" {
		jsonError(w, "email-required", http.StatusBadRequest)
		return
	}

	user, err := db.GetUserByID(h.DB, userID)
	if err != nil {
		jsonError(w, "user-not-found", http.StatusNotFound)
		return
	}

	if !user.PasswordHash.Valid || !user.VerifyPassword(req.CurrentPassword) {
		jsonError(w, "invalid-password", http.StatusUnauthorized)
		return
	}

	existing, _ := db.GetUserByEmail(h.DB, req.NewEmail)
	if existing != nil && existing.ID != userID {
		jsonError(w, "email-already-registered", http.StatusConflict)
		return
	}

	_, err = h.DB.Exec(`UPDATE users SET email = ? WHERE id = ?`, req.NewEmail, userID)
	if err != nil {
		jsonError(w, "db-error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func (h *AccountHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid-json", http.StatusBadRequest)
		return
	}

	if req.NewPassword == "" || len(req.NewPassword) < 8 {
		jsonError(w, "password-too-short", http.StatusBadRequest)
		return
	}

	user, err := db.GetUserByID(h.DB, userID)
	if err != nil {
		jsonError(w, "user-not-found", http.StatusNotFound)
		return
	}

	if user.PasswordHash.Valid && user.PasswordHash.String != "" {
		if !user.VerifyPassword(req.CurrentPassword) {
			jsonError(w, "invalid-password", http.StatusUnauthorized)
			return
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		jsonError(w, "hash-failed", http.StatusInternalServerError)
		return
	}

	_, err = h.DB.Exec(`UPDATE users SET passwordHash = ? WHERE id = ?`, string(hash), userID)
	if err != nil {
		jsonError(w, "db-error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}
