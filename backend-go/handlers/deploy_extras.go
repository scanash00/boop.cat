// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"

	"boop-cat/db"
	"boop-cat/deploy"
	"boop-cat/middleware"
)

func (h *DeployHandler) GetDeploymentLogs(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	deployID := chi.URLParam(r, "id")

	d, err := db.GetDeploymentByID(h.DB, deployID)
	if err != nil {
		http.Error(w, "not-found", http.StatusNotFound)
		return
	}

	if d.UserID != userID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if !d.LogsPath.Valid || d.LogsPath.String == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(""))
		return
	}

	content, err := os.ReadFile(d.LogsPath.String)
	if err != nil {

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(""))
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(content)
}

func (h *AuthHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	if h.Engine != nil {
		sites, _ := db.GetSitesByUserID(h.DB, userID)
		for _, site := range sites {
			err := h.Engine.CleanupSite(site.ID, userID)
			if err != nil {

			}
		}
	}

	_, err := h.DB.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		jsonError(w, "delete-failed", http.StatusInternalServerError)
		return
	}

	middleware.LogoutUser(w, r)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func (h *DeployHandler) StopDeployment(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	deployID := chi.URLParam(r, "id")

	d, err := db.GetDeploymentByID(h.DB, deployID)
	if err != nil {
		http.Error(w, "not-found", http.StatusNotFound)
		return
	}
	if d.UserID != userID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	err = h.Engine.CancelDeployment(deployID)
	if err != nil {

		dCheck, errDb := db.GetDeploymentByID(h.DB, deployID)

		if errDb == nil && (dCheck.Status == "running" || dCheck.Status == "active") {

			cf := deploy.NewCloudflareClient(
				os.Getenv("CF_ACCOUNT_ID"),
				os.Getenv("CF_KV_NAMESPACE_ID"),
				os.Getenv("CF_API_TOKEN"),
			)

			site, _ := db.GetSiteByID(h.DB, userID, dCheck.SiteID)
			if site != nil {

				rootDomain := os.Getenv("FSD_EDGE_ROOT_DOMAIN")
				routingKey := site.Domain
				if rootDomain != "" && strings.HasSuffix(site.Domain, "."+rootDomain) {
					routingKey = strings.TrimSuffix(site.Domain, "."+rootDomain)
				}

				cf.RemoveRouting(routingKey, site.ID, site.Domain)

				customDomains, _ := db.ListCustomDomains(h.DB, site.ID)
				for _, cd := range customDomains {

					cf.RemoveRouting("", site.ID, cd.Hostname)
				}
			}

			db.UpdateDeploymentStatus(h.DB, deployID, "stopped", "")
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"ok":true}`))
			return
		}

		if errDb == nil && dCheck.Status == "building" {
			db.UpdateDeploymentStatus(h.DB, deployID, "canceled", "")
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"ok":true}`))
			return
		}

		jsonError(w, "cancel-failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func (h *DeployHandler) PreviewSite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		GitURL string `json:"gitUrl"`
		Branch string `json:"branch"`
		Subdir string `json:"subdir"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid-json", http.StatusBadRequest)
		return
	}

	result, err := h.Engine.PreviewGitRepo(req.GitURL)
	if err != nil {

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "preview-failed",
			"message": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
