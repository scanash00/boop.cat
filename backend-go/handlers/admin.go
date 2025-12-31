package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	"boop-cat/db"
	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	DB *sql.DB
}

func NewAdminHandler(database *sql.DB) *AdminHandler {
	return &AdminHandler{DB: database}
}

func (h *AdminHandler) RequireAdminKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("x-admin-api-key")
		if key == "" || key != os.Getenv("ADMIN_API_KEY") {
			jsonError(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *AdminHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(h.RequireAdminKey)

	r.Post("/poll-dmca", h.PollDMCA)
	r.Post("/ban", h.BanUser)
	r.Get("/lookup", h.LookupDomain)
	r.Get("/sites", h.ListSites)

	return r
}

type banRequest struct {
	UserID string `json:"userId"`
	Ban    bool   `json:"ban"`
	IP     string `json:"ip,omitempty"`
}

func (h *AdminHandler) BanUser(w http.ResponseWriter, r *http.Request) {
	var req banRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid-params", http.StatusBadRequest)
		return
	}
	if req.UserID == "" {
		jsonError(w, "invalid-params", http.StatusBadRequest)
		return
	}

	user, err := db.GetUserByID(h.DB, req.UserID)
	if err != nil {
		jsonError(w, "user-not-found", http.StatusNotFound)
		return
	}

	deletedSites := []map[string]interface{}{}

	if req.Ban {

		sites, _ := db.GetSitesByUserID(h.DB, user.ID)

		for _, s := range sites {

			if err := db.DeleteSite(h.DB, user.ID, s.ID); err == nil {
				deletedSites = append(deletedSites, map[string]interface{}{
					"id":   s.ID,
					"name": s.Name,
				})
			}
		}

		if req.IP != "" {
			_ = db.BanIP(h.DB, req.IP, user.ID, "Banned with user "+user.Email)
		}
	}

	_ = db.SetUserBanned(h.DB, user.ID, req.Ban)

	user.Banned = req.Ban

	response := map[string]interface{}{
		"ok": true,
		"user": map[string]interface{}{
			"id":     user.ID,
			"email":  user.Email,
			"banned": user.Banned,
		},
		"deletedSites": deletedSites,
		"bannedIP":     nil,
	}
	if req.Ban && req.IP != "" {
		response["bannedIP"] = req.IP
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AdminHandler) LookupDomain(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		jsonError(w, "missing-domain", http.StatusBadRequest)
		return
	}

	site, err := db.GetSiteByDomain(h.DB, domain)
	if err != nil {

		cd, err := db.GetCustomDomainByHostname(h.DB, domain)
		if err != nil || cd == nil {
			jsonError(w, "not-found", http.StatusNotFound)
			return
		}

		site, err = db.GetSiteByIDAdmin(h.DB, cd.SiteID)
		if err != nil {
			jsonError(w, "site-not-found", http.StatusNotFound)
			return
		}
	}

	user, _ := db.GetUserByID(h.DB, site.UserID)

	response := map[string]interface{}{
		"ok": true,
		"site": map[string]interface{}{
			"id":     site.ID,
			"name":   site.Name,
			"domain": site.Domain,
		},
		"user": nil,
	}

	if user != nil {
		response["user"] = map[string]interface{}{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
			"banned":   user.Banned,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AdminHandler) ListSites(w http.ResponseWriter, r *http.Request) {

	sites, err := db.GetAllSites(h.DB)
	if err != nil {
		jsonError(w, "db-error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":    true,
		"sites": sites,
		"total": len(sites),
	})
}

func (h *AdminHandler) PollDMCA(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true,"message":"Polling initiated (stub)."}`))
}
