package handlers

import (
	"net/http"

	"boop-cat/middleware"
)

func (h *AuthHandler) GitHubInstalled(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	installationID := r.URL.Query().Get("installation_id")
	setupAction := r.URL.Query().Get("setup_action")

	if installationID != "" && setupAction == "install" {

		_, _ = h.DB.Exec(`
			INSERT OR IGNORE INTO githubAppInstallations (id, userId, installationId, createdAt)
			VALUES (?, ?, ?, datetime('now'))
		`, generateToken()[:16], user.ID, installationID)
	}

	http.Redirect(w, r, "/dashboard", http.StatusFound)
}
