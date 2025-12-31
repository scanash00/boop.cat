package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nrednav/cuid2"

	"boop-cat/db"
	"boop-cat/middleware"
)

type APIKeysHandler struct {
	DB *sql.DB
}

func NewAPIKeysHandler(database *sql.DB) *APIKeysHandler {
	return &APIKeysHandler{DB: database}
}

func (h *APIKeysHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireLogin)

	r.Get("/", h.ListAPIKeys)
	r.Post("/", h.CreateAPIKey)
	r.Delete("/{id}", h.DeleteAPIKey)

	return r
}

func (h *APIKeysHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	keys, err := db.ListAPIKeys(h.DB, userID)
	if err != nil {
		jsonError(w, "list-keys-failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

func (h *APIKeysHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid-json", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		jsonError(w, "name-required", http.StatusBadRequest)
		return
	}

	id := cuid2.Generate()
	rawKey := "sk_" + cuid2.Generate() + cuid2.Generate()
	prefix := rawKey[:6]

	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	err := db.CreateAPIKey(h.DB, id, userID, req.Name, keyHash, prefix)
	if err != nil {
		jsonError(w, "create-key-failed", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"id":        id,
		"name":      req.Name,
		"keyPrefix": prefix,
		"key":       rawKey,
		"createdAt": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *APIKeysHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	keyID := chi.URLParam(r, "id")
	if keyID == "" {
		jsonError(w, "id-required", http.StatusBadRequest)
		return
	}

	err := db.DeleteAPIKey(h.DB, userID, keyID)
	if err != nil {
		jsonError(w, "delete-key-failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}
