// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/markbates/goth/gothic"
	"github.com/nrednav/cuid2"

	"boop-cat/db"
	"boop-cat/middleware"
)

func (h *AuthHandler) MountOAuthRoutes(r chi.Router) {
	r.Get("/{provider}", h.BeginOAuth)
	r.Get("/{provider}/callback", h.OAuthCallback)
}

func (h *AuthHandler) BeginOAuth(w http.ResponseWriter, r *http.Request) {

	gothic.BeginAuthHandler(w, r)
}

func (h *AuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Redirect(w, r, "/?error=auth-failed", http.StatusTemporaryRedirect)
		return
	}

	provider := chi.URLParam(r, "provider")

	loggedInUser := middleware.GetUser(r.Context())

	existingAcc, err := db.FindOAuthAccount(h.DB, provider, gothUser.UserID)
	if err == nil && existingAcc != nil {

		if loggedInUser != nil && existingAcc.UserID != loggedInUser.ID {

			http.Redirect(w, r, "/dashboard/account?error=already-linked", http.StatusTemporaryRedirect)
			return
		}

		_, _ = h.DB.Exec(`UPDATE oauthAccounts SET accessToken = ? WHERE id = ?`, gothUser.AccessToken, existingAcc.ID)

		if err := middleware.LoginUser(w, r, existingAcc.UserID); err != nil {
			http.Redirect(w, r, "/?error=session-error", http.StatusTemporaryRedirect)
			return
		}
		h.finalizeGitHubLogin(w, r, existingAcc.UserID, provider, gothUser.AccessToken)
		return
	}

	if loggedInUser != nil {
		err = db.CreateOAuthAccount(h.DB, cuid2.Generate(), provider, gothUser.UserID, loggedInUser.ID, gothUser.AccessToken, gothUser.Name)
		if err != nil {
			http.Redirect(w, r, "/dashboard/account?error=link-failed", http.StatusTemporaryRedirect)
			return
		}

		http.Redirect(w, r, "/dashboard/account", http.StatusTemporaryRedirect)
		return
	}

	existingUser, err := db.GetUserByEmail(h.DB, gothUser.Email)
	if err == nil && existingUser != nil {

		err = db.CreateOAuthAccount(h.DB, cuid2.Generate(), provider, gothUser.UserID, existingUser.ID, gothUser.AccessToken, gothUser.Name)
		if err != nil {
			http.Redirect(w, r, "/?error=link-failed", http.StatusTemporaryRedirect)
			return
		}

		if err := middleware.LoginUser(w, r, existingUser.ID); err != nil {
			http.Redirect(w, r, "/?error=session-error", http.StatusTemporaryRedirect)
			return
		}
		h.finalizeGitHubLogin(w, r, existingUser.ID, provider, gothUser.AccessToken)
		return
	}

	userID := cuid2.Generate()

	randomPwd := cuid2.Generate() + cuid2.Generate()
	_, err = db.CreateUser(h.DB, userID, gothUser.Email, randomPwd)
	if err != nil {
		http.Redirect(w, r, "/?error=create-user-failed", http.StatusTemporaryRedirect)
		return
	}

	verified := false
	if v, ok := gothUser.RawData["verified"].(bool); ok && v {
		verified = true
	} else if v, ok := gothUser.RawData["email_verified"].(bool); ok && v {
		verified = true
	}

	if verified {
		_, _ = h.DB.Exec(`UPDATE users SET emailVerified = 1 WHERE id = ?`, userID)
	}

	err = db.CreateOAuthAccount(h.DB, cuid2.Generate(), provider, gothUser.UserID, userID, gothUser.AccessToken, gothUser.Name)
	if err != nil {
		http.Redirect(w, r, "/?error=link-failed", http.StatusTemporaryRedirect)
		return
	}

	if err := middleware.LoginUser(w, r, userID); err != nil {
		http.Redirect(w, r, "/?error=session-error", http.StatusTemporaryRedirect)
		return
	}

	if err := middleware.LoginUser(w, r, userID); err != nil {
		http.Redirect(w, r, "/?error=session-error", http.StatusTemporaryRedirect)
		return
	}

	h.finalizeGitHubLogin(w, r, userID, provider, gothUser.AccessToken)
}

func (h *AuthHandler) finalizeGitHubLogin(w http.ResponseWriter, r *http.Request, userID, provider, accessToken string) {

	if provider != "github" {
		http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
		return
	}

	installURL := os.Getenv("GITHUB_APP_INSTALL_URL")
	appID := os.Getenv("GITHUB_APP_ID")
	if installURL == "" || appID == "" {
		http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
		return
	}

	var exists bool
	err := h.DB.QueryRow(`SELECT 1 FROM githubAppInstallations WHERE userId = ? LIMIT 1`, userID).Scan(&exists)
	if err == nil && exists {
		http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
		return
	}

	hasInstall := false
	if accessToken != "" {
		req, _ := http.NewRequest("GET", "https://api.github.com/user/installations", nil)
		req.Header.Set("Authorization", "token "+accessToken)
		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := http.DefaultClient.Do(req)
		if err == nil && resp.StatusCode == 200 {
			var data struct {
				Installations []struct {
					ID    int64 `json:"id"`
					AppID int64 `json:"app_id"`
				} `json:"installations"`
			}
			if json.NewDecoder(resp.Body).Decode(&data) == nil {
				for _, inst := range data.Installations {
					if fmt.Sprintf("%d", inst.AppID) == appID {

						h.DB.Exec(`INSERT OR IGNORE INTO githubAppInstallations (id, userId, installationId, createdAt) VALUES (?, ?, ?, datetime('now'))`,
							cuid2.Generate(), userID, fmt.Sprintf("%d", inst.ID))
						hasInstall = true
						break
					}
				}
			}
			resp.Body.Close()
		}
	}

	if hasInstall {
		http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
	} else {

		http.Redirect(w, r, installURL, http.StatusTemporaryRedirect)
	}
}
