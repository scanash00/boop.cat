package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/nrednav/cuid2"

	"boop-cat/db"
	"boop-cat/deploy"
	"boop-cat/lib"
	"boop-cat/middleware"
)

type AuthHandler struct {
	DB     *sql.DB
	Engine *deploy.Engine
}

func NewAuthHandler(database *sql.DB, engine *deploy.Engine) *AuthHandler {
	return &AuthHandler{DB: database, Engine: engine}
}

func jsonError(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": error})
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/login", h.Login)
	r.Post("/register", h.Register)
	r.Post("/logout", h.Logout)
	r.Get("/me", h.Me)
	r.Get("/providers", h.GetProviders)

	r.Post("/verify-email", h.SendVerificationEmail)
	r.Get("/verify-email", h.VerifyEmail)
	r.Post("/verify-email/resend", h.ResendVerificationEmail)
	r.Post("/forgot-password", h.RequestPasswordReset)
	r.Post("/reset-password", h.ResetPasswordConfirm)

	return r
}

type loginRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	TurnstileToken string `json:"turnstileToken"`
}

type registerRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	TurnstileToken string `json:"turnstileToken"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid-json", http.StatusBadRequest)
		return
	}

	ts := lib.VerifyTurnstile(req.TurnstileToken, r.RemoteAddr)
	if !ts.OK {
		jsonError(w, "captcha-failed", http.StatusBadRequest)
		return
	}

	user, err := db.GetUserByEmail(h.DB, req.Email)
	if err != nil {

		jsonError(w, "invalid-credentials", http.StatusUnauthorized)
		return
	}

	if !user.VerifyPassword(req.Password) {
		jsonError(w, "invalid-credentials", http.StatusUnauthorized)
		return
	}

	if user.Banned {
		jsonError(w, "user-banned", http.StatusForbidden)
		return
	}

	_ = db.UpdateLastLogin(h.DB, user.ID)

	if err := middleware.LoginUser(w, r, user.ID); err != nil {
		jsonError(w, "session-error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": true,
		"user": map[string]interface{}{
			"id":            user.ID,
			"email":         user.Email,
			"username":      getNullString(user.Username),
			"emailVerified": user.EmailVerified,
		},
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid-json", http.StatusBadRequest)
		return
	}

	ts := lib.VerifyTurnstile(req.TurnstileToken, r.RemoteAddr)
	if !ts.OK {
		jsonError(w, "captcha-failed", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		jsonError(w, "missing-fields", http.StatusBadRequest)
		return
	}

	existing, _ := db.GetUserByEmail(h.DB, req.Email)
	if existing != nil {
		jsonError(w, "email-already-registered", http.StatusConflict)
		return
	}

	id := cuid2.Generate()
	user, err := db.CreateUser(h.DB, id, req.Email, req.Password)
	if err != nil {
		jsonError(w, "create-user-failed", http.StatusInternalServerError)
		return
	}

	if err := middleware.LoginUser(w, r, user.ID); err != nil {
		jsonError(w, "session-error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": true,
		"user": map[string]interface{}{
			"id":            user.ID,
			"email":         user.Email,
			"username":      nil,
			"emailVerified": false,
		},
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	middleware.LogoutUser(w, r)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"authenticated":false,"user":null}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"authenticated": true,
		"user": map[string]interface{}{
			"id":            user.ID,
			"email":         user.Email,
			"username":      getNullString(user.Username),
			"emailVerified": user.EmailVerified,
		},
	})
}

func (h *AuthHandler) GetProviders(w http.ResponseWriter, r *http.Request) {
	config := map[string]interface{}{
		"github":           os.Getenv("GITHUB_CLIENT_ID") != "" && os.Getenv("GITHUB_CLIENT_SECRET") != "",
		"google":           os.Getenv("GOOGLE_CLIENT_ID") != "" && os.Getenv("GOOGLE_CLIENT_SECRET") != "",
		"atproto":          os.Getenv("ATPROTO_PRIVATE_KEY_1") != "",
		"turnstileSiteKey": os.Getenv("TURNSTILE_SITE_KEY"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func getNullString(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
