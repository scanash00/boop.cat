package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"boop-cat/db"
	"boop-cat/deploy"
	"boop-cat/middleware"
)

type DeployHandler struct {
	DB     *sql.DB
	Engine *deploy.Engine
}

func NewDeployHandler(database *sql.DB) *DeployHandler {

	engine := deploy.NewEngine(
		database,
		os.Getenv("B2_KEY_ID"),
		os.Getenv("B2_APP_KEY"),
		os.Getenv("B2_BUCKET_ID"),
		os.Getenv("CF_API_TOKEN"),
		os.Getenv("CF_ACCOUNT_ID"),
		os.Getenv("CF_KV_NAMESPACE_ID"),
	)
	return &DeployHandler{DB: database, Engine: engine}
}

func (h *DeployHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireLogin)

	r.Post("/sites/{siteId}/deploy", h.TriggerDeploy)
	r.Get("/sites/{siteId}/deployments", h.ListDeployments)
	r.Get("/deployments/{id}", h.GetDeployment)

	return r
}

func (h *DeployHandler) TriggerDeploy(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	siteID := chi.URLParam(r, "siteId")

	site, err := db.GetSiteByID(h.DB, userID, siteID)
	if err != nil || site == nil {
		jsonError(w, "site-not-found", http.StatusNotFound)
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

func (h *DeployHandler) ListDeployments(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	siteID := chi.URLParam(r, "siteId")

	site, err := db.GetSiteByID(h.DB, userID, siteID)
	if err != nil || site == nil {
		jsonError(w, "site-not-found", http.StatusNotFound)
		return
	}

	deployments, err := db.ListDeployments(h.DB, userID, siteID)
	if err != nil {
		jsonError(w, "list-failed", http.StatusInternalServerError)
		return
	}

	var resp []db.DeploymentResponse
	for _, d := range deployments {
		resp = append(resp, d.ToResponse())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *DeployHandler) GetDeployment(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	deployID := chi.URLParam(r, "id")

	d, err := db.GetDeploymentByID(h.DB, deployID)
	if err != nil {
		jsonError(w, "not-found", http.StatusNotFound)
		return
	}

	if d.UserID != userID {
		jsonError(w, "forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d.ToResponse())
}

func (h *DeployHandler) toResponse(d *db.Deployment) map[string]interface{} {
	resp := map[string]interface{}{
		"id":        d.ID,
		"status":    d.Status,
		"createdAt": d.CreatedAt,
	}
	if d.URL.Valid {
		resp["url"] = d.URL.String
	}
	if d.CommitSha.Valid {
		resp["commitSha"] = d.CommitSha.String
	}
	if d.CommitMessage.Valid {
		resp["commitMessage"] = d.CommitMessage.String
	}
	if d.CommitAuthor.Valid {
		resp["commitAuthor"] = d.CommitAuthor.String
	}
	return resp
}
