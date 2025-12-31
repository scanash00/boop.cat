// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"boop-cat/db"
	"boop-cat/deploy"
	"boop-cat/middleware"
)

type APIV1Handler struct {
	DB     *sql.DB
	Engine *deploy.Engine
}

func NewAPIV1Handler(database *sql.DB, engine *deploy.Engine) *APIV1Handler {
	return &APIV1Handler{DB: database, Engine: engine}
}

func (h *APIV1Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireAPIKey(h.DB))

	r.Get("/sites", h.ListSites)
	r.Get("/sites/{id}", h.GetSite)
	r.Post("/sites/{id}/deploy", h.TriggerDeploy)
	r.Get("/sites/{id}/deployments", h.ListDeployments)

	return r
}

func (h *APIV1Handler) ListSites(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	sites, err := db.ListSites(h.DB, userID)
	if err != nil {
		jsonError(w, "list-sites-failed", http.StatusInternalServerError)
		return
	}

	var resp []db.SiteResponse
	for _, s := range sites {
		resp = append(resp, s.ToResponse())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sites": resp,
	})
}

func (h *APIV1Handler) GetSite(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	siteID := chi.URLParam(r, "id")

	site, err := db.GetSiteByID(h.DB, userID, siteID)
	if err != nil || site == nil {
		jsonError(w, "site-not-found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(site.ToResponse())
}

func (h *APIV1Handler) TriggerDeploy(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	siteID := chi.URLParam(r, "id")

	site, err := db.GetSiteByID(h.DB, userID, siteID)
	if err != nil || site == nil {
		jsonError(w, "site-not-found", http.StatusNotFound)
		return
	}

	wait := r.URL.Query().Get("wait") == "true"

	if wait {

		logStream := make(chan string, 10)

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}

		go func() {
			_, err := h.Engine.DeploySite(siteID, userID, logStream)
			if err != nil {
				logStream <- fmt.Sprintf("Error starting deployment: %v", err)
				close(logStream)
			}
		}()

		for msg := range logStream {
			fmt.Fprintf(w, "%s\n", msg)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
		return
	}

	d, err := h.Engine.DeploySite(siteID, userID, nil)
	if err != nil {
		jsonError(w, "deploy-failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d.ToResponse())
}

func (h *APIV1Handler) ListDeployments(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	siteID := chi.URLParam(r, "id")

	site, err := db.GetSiteByID(h.DB, userID, siteID)
	if err != nil || site == nil {
		jsonError(w, "site-not-found", http.StatusNotFound)
		return
	}

	deps, err := db.ListDeployments(h.DB, userID, siteID)
	if err != nil {
		jsonError(w, "list-deployments-failed", http.StatusInternalServerError)
		return
	}

	var resp []db.DeploymentResponse
	for _, d := range deps {
		resp = append(resp, d.ToResponse())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deployments": resp,
	})
}
