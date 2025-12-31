// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package handlers

import (
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"boop-cat/db"
	"boop-cat/lib"
	"boop-cat/middleware"
)

func (h *AuthHandler) SendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if user.EmailVerified {
		jsonError(w, "already-verified", http.StatusBadRequest)
		return
	}

	token := generateToken()
	expiresAt := time.Now().Add(24 * time.Hour).Unix()
	id := generateToken()

	err := db.CreateVerificationToken(h.DB, id, user.ID, token, expiresAt)
	if err != nil {
		jsonError(w, "db-error", http.StatusInternalServerError)
		return
	}

	username := ""
	if user.Username.Valid {
		username = user.Username.String
	}
	go lib.SendVerificationEmail(user.Email, token, username)

	w.Write([]byte(`{"ok":true}`))
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing-token", http.StatusBadRequest)
		return
	}

	ev, err := db.GetVerificationToken(h.DB, token)
	if err != nil || ev == nil {
		http.Error(w, "invalid-token", http.StatusBadRequest)
		return
	}

	if time.Now().Unix() > ev.ExpiresAt {
		http.Error(w, "expired-token", http.StatusBadRequest)
		return
	}

	err = db.UpdateUserEmailVerified(h.DB, ev.UserID)
	if err != nil {
		http.Error(w, "db-error", http.StatusInternalServerError)
		return
	}

	db.MarkTokenUsed(h.DB, ev.ID)

	http.Redirect(w, r, "/dashboard?verified=true", http.StatusFound)
}

func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid-json", http.StatusBadRequest)
		return
	}

	user, err := db.GetUserByEmail(h.DB, req.Email)
	if err != nil || user == nil {

		w.Write([]byte(`{"ok":true}`))
		return
	}

	token := generateToken()
	expiresAt := time.Now().Add(1 * time.Hour).Unix()
	id := generateToken()

	err = db.CreateVerificationToken(h.DB, id, user.ID, token, expiresAt)
	if err != nil {
		jsonError(w, "db-error", http.StatusInternalServerError)
		return
	}

	username := ""
	if user.Username.Valid {
		username = user.Username.String
	}
	go lib.SendPasswordResetEmail(user.Email, token, username)

	w.Write([]byte(`{"ok":true}`))
}

func (h *AuthHandler) ResetPasswordConfirm(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid-json", http.StatusBadRequest)
		return
	}

	ev, err := db.GetVerificationToken(h.DB, req.Token)
	if err != nil || ev == nil {
		jsonError(w, "invalid-token", http.StatusBadRequest)
		return
	}

	if time.Now().Unix() > ev.ExpiresAt {
		jsonError(w, "expired-token", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		jsonError(w, "hash-failed", http.StatusInternalServerError)
		return
	}

	err = db.UpdateUserPassword(h.DB, ev.UserID, string(hash))
	if err != nil {
		jsonError(w, "db-error", http.StatusInternalServerError)
		return
	}

	db.MarkTokenUsed(h.DB, ev.ID)

	w.Write([]byte(`{"ok":true}`))
}

func (h *AuthHandler) ResendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid-json", http.StatusBadRequest)
		return
	}

	user, err := db.GetUserByEmail(h.DB, req.Email)
	if err != nil || user == nil {

		w.Write([]byte(`{"ok":true}`))
		return
	}

	if user.EmailVerified {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"alreadyVerified":true}`))
		return
	}

	token := generateToken()
	expiresAt := time.Now().Add(24 * time.Hour).Unix()
	id := generateToken()

	err = db.CreateVerificationToken(h.DB, id, user.ID, token, expiresAt)
	if err != nil {
		jsonError(w, "db-error", http.StatusInternalServerError)
		return
	}

	username := ""
	if user.Username.Valid {
		username = user.Username.String
	}
	go lib.SendVerificationEmail(user.Email, token, username)

	w.Write([]byte(`{"ok":true}`))
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
