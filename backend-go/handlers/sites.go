// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/nrednav/cuid2"

	"boop-cat/db"
	"boop-cat/deploy"
	"boop-cat/lib"
	"boop-cat/middleware"
)

type SitesHandler struct {
	DB     *sql.DB
	Engine *deploy.Engine
}

func NewSitesHandler(database *sql.DB, engine *deploy.Engine) *SitesHandler {
	return &SitesHandler{DB: database, Engine: engine}
}

func (h *SitesHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireLogin)

	r.Get("/", h.ListSites)
	r.Post("/", h.CreateSite)
	r.Patch("/{id}", h.UpdateSiteEnv)
	r.Patch("/{id}/settings", h.UpdateSiteSettings)
	r.Put("/{id}/settings", h.UpdateSiteSettings)
	r.Post("/{id}/settings", h.UpdateSiteSettings)
	r.Delete("/{id}", h.DeleteSite)

	return r
}

func (h *SitesHandler) ListSites(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	sites, err := db.GetSitesByUserID(h.DB, userID)
	if err != nil {
		jsonError(w, "list-sites-failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var resp []db.SiteResponse
	for _, s := range sites {
		r := s.ToResponse()
		if r.EnvText != "" {
			r.EnvText = lib.Decrypt(r.EnvText)
		}
		resp = append(resp, r)
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *SitesHandler) CreateSite(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req struct {
		Name         string `json:"name"`
		GitURL       string `json:"gitUrl"`
		Branch       string `json:"branch"`
		Subdir       string `json:"subdir"`
		Domain       string `json:"domain"`
		BuildCommand string `json:"buildCommand"`
		OutputDir    string `json:"outputDir"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid-json", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		jsonError(w, "name-required", http.StatusBadRequest)
		return
	}

	siteID := cuid2.Generate()

	if req.Branch == "" {
		req.Branch = "main"
	}

	edgeRoot := os.Getenv("FSD_EDGE_ROOT_DOMAIN")
	if edgeRoot != "" {

		edgeRoot = strings.TrimLeft(edgeRoot, ".")

		normalized := strings.ToLower(strings.TrimSpace(req.Domain))
		if normalized != "" {
			if !strings.HasSuffix(normalized, "."+edgeRoot) && !strings.Contains(normalized, ".") {

				normalized = normalized + "." + edgeRoot
			} else if !strings.HasSuffix(normalized, "."+edgeRoot) {

			}
		} else {

			label := strings.ToLower(req.Name)
			reg := regexp.MustCompile("[^a-z0-9-]")
			label = reg.ReplaceAllString(label, "-")
			label = strings.Trim(label, "-")
			if label == "" {
				label = "site"
			}
			normalized = label + "." + edgeRoot
		}

		req.Domain = normalized

		baseLabel := strings.TrimSuffix(req.Domain, "."+edgeRoot)
		for i := 0; i < 5; i++ {
			existing, _ := db.GetSiteByDomain(h.DB, req.Domain)
			if existing == nil {
				break
			}

			suffix := cuid2.Generate()[:4]
			req.Domain = fmt.Sprintf("%s-%s.%s", baseLabel, suffix, edgeRoot)
		}
	}

	err := db.CreateSite(h.DB, siteID, userID, req.Name, req.Domain, req.GitURL, req.Branch, req.Subdir, req.BuildCommand, req.OutputDir)
	if err != nil {
		jsonError(w, "create-site-failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	site, _ := db.GetSiteByID(h.DB, userID, siteID)
	resp := site.ToResponse()
	if resp.EnvText != "" {
		resp.EnvText = lib.Decrypt(resp.EnvText)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SitesHandler) UpdateSiteEnv(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	siteID := chi.URLParam(r, "siteId")
	if siteID == "" {
		siteID = chi.URLParam(r, "id")
	}

	var req struct {
		EnvText string `json:"envText"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid-json", http.StatusBadRequest)
		return
	}

	site, err := db.GetSiteByID(h.DB, userID, siteID)
	if err != nil || site == nil {
		jsonError(w, "site-not-found", http.StatusNotFound)
		return
	}

	if site.UserID != userID {
		jsonError(w, "forbidden", http.StatusForbidden)
		return
	}

	encryptedEnv := lib.Encrypt(req.EnvText)
	err = db.UpdateSiteEnv(h.DB, siteID, encryptedEnv)
	if err != nil {
		jsonError(w, "update-failed", http.StatusInternalServerError)
		return
	}

	updated, _ := db.GetSiteByID(h.DB, userID, siteID)
	resp := updated.ToResponse()
	if resp.EnvText != "" {
		resp.EnvText = lib.Decrypt(resp.EnvText)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SitesHandler) UpdateSiteSettings(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	siteID := chi.URLParam(r, "siteId")
	if siteID == "" {
		siteID = chi.URLParam(r, "id")
	}

	var req struct {
		Name         string `json:"name"`
		GitURL       string `json:"gitUrl"`
		Branch       string `json:"branch"`
		Subdir       string `json:"subdir"`
		Domain       string `json:"domain"`
		BuildCommand string `json:"buildCommand"`
		OutputDir    string `json:"outputDir"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid-json", http.StatusBadRequest)
		return
	}

	site, err := db.GetSiteByID(h.DB, userID, siteID)
	if err != nil || site == nil {
		jsonError(w, "site-not-found", http.StatusNotFound)
		return
	}

	if req.Name == "" {
		req.Name = site.Name
	}
	if req.GitURL == "" && site.GitURL.Valid {
		req.GitURL = site.GitURL.String
	}
	if req.Branch == "" && site.GitBranch.Valid {
		req.Branch = site.GitBranch.String
	}
	if req.Subdir == "" && site.GitSubdir.Valid {
		req.Subdir = site.GitSubdir.String
	}
	if req.Domain == "" {
		req.Domain = site.Domain
	}
	if req.BuildCommand == "" && site.BuildCommand.Valid {
		req.BuildCommand = site.BuildCommand.String
	}
	if req.OutputDir == "" && site.OutputDir.Valid {
		req.OutputDir = site.OutputDir.String
	}

	if req.Domain != "" && req.Domain != site.Domain {
		edgeRoot := os.Getenv("FSD_EDGE_ROOT_DOMAIN")
		if edgeRoot != "" {
			normalized := strings.ToLower(strings.TrimSpace(req.Domain))
			if !strings.HasSuffix(normalized, "."+edgeRoot) && !strings.Contains(normalized, ".") {
				req.Domain = normalized + "." + edgeRoot
			}
		}
	}

	err = db.UpdateSiteSettings(h.DB, siteID, req.Name, req.Domain, req.GitURL, req.Branch, req.Subdir, req.BuildCommand, req.OutputDir)
	if err != nil {
		jsonError(w, "update-failed", http.StatusInternalServerError)
		return
	}

	updated, _ := db.GetSiteByID(h.DB, userID, siteID)
	resp := updated.ToResponse()
	if resp.EnvText != "" {
		resp.EnvText = lib.Decrypt(resp.EnvText)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SitesHandler) DeleteSite(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	siteID := chi.URLParam(r, "siteId")
	if siteID == "" {
		siteID = chi.URLParam(r, "id")
	}

	site, err := db.GetSiteByID(h.DB, userID, siteID)
	if err != nil || site == nil {
		jsonError(w, "site-not-found", http.StatusNotFound)
		return
	}

	if h.Engine != nil {

		cleanupErr := h.Engine.CleanupSite(siteID, userID)
		if cleanupErr != nil {
			fmt.Printf("Warning: Failed to cleanup external resources for site %s: %v\n", siteID, cleanupErr)
		}
	}

	err = db.DeleteSite(h.DB, userID, siteID)
	if err != nil {
		jsonError(w, "delete-failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}
