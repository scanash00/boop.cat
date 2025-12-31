package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/nrednav/cuid2"

	"boop-cat/db"
	"boop-cat/deploy"
)

type GitHubWebhookHandler struct {
	DB     *sql.DB
	Engine *deploy.Engine
}

func NewGitHubWebhookHandler(database *sql.DB, engine *deploy.Engine) *GitHubWebhookHandler {
	return &GitHubWebhookHandler{DB: database, Engine: engine}
}

func (h *GitHubWebhookHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.HandleWebhook)
	return r
}

func verifySignature(payload []byte, signature, secret string) bool {
	if signature == "" || secret == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	expectedSignature := "sha256=" + hex.EncodeToString(expectedMAC)

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func (h *GitHubWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	secret := os.Getenv("GITHUB_APP_WEBHOOK_SECRET")
	signature := r.Header.Get("X-Hub-Signature-256")
	eventType := r.Header.Get("X-GitHub-Event")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read-failed", http.StatusInternalServerError)
		return
	}

	if secret != "" {
		if !verifySignature(body, signature, secret) {
			http.Error(w, "invalid-signature", http.StatusUnauthorized)
			return
		}
	}

	var event map[string]interface{}
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "invalid-json", http.StatusBadRequest)
		return
	}

	if eventType == "push" {
		h.handlePush(w, event)
		return
	}

	if eventType == "installation" {
		h.handleInstallation(w, event)
		return
	}

	w.Write([]byte(`{"ok":true,"ignored":true}`))
}

func (h *GitHubWebhookHandler) handlePush(w http.ResponseWriter, event map[string]interface{}) {

	repoMap, _ := event["repository"].(map[string]interface{})
	if repoMap == nil {
		w.Write([]byte(`{"ok":true,"ignored":"no-repo"}`))
		return
	}

	repoURL, _ := repoMap["clone_url"].(string)

	ref, _ := event["ref"].(string)
	branch := strings.TrimPrefix(ref, "refs/heads/")

	if repoURL == "" || branch == "" {
		w.Write([]byte(`{"ok":true,"ignored":"no-url-or-branch"}`))
		return
	}

	sites, err := db.GetSitesByRepo(h.DB, repoURL, branch)
	if err != nil {
		fmt.Printf("[Webhook] Failed to find sites: %v\n", err)
		http.Error(w, "db-error", http.StatusInternalServerError)
		return
	}

	processed := 0
	for _, site := range sites {
		fmt.Printf("[Webhook] Triggering deploy for site %s\n", site.ID)
		_, err := h.Engine.DeploySite(site.ID, site.UserID, nil)
		if err != nil {
			fmt.Printf("[Webhook] Deploy failed for %s: %v\n", site.ID, err)
		} else {
			processed++
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":       true,
		"deployed": processed,
		"matched":  len(sites),
	})
}

func (h *GitHubWebhookHandler) handleInstallation(w http.ResponseWriter, event map[string]interface{}) {
	action, _ := event["action"].(string)
	installMap, _ := event["installation"].(map[string]interface{})
	if installMap == nil {
		w.Write([]byte(`{"ok":true}`))
		return
	}

	instID := fmt.Sprintf("%.0f", installMap["id"].(float64))

	if action == "deleted" {
		db.RemoveGitHubInstallation(h.DB, instID)
	} else if action == "created" {
		account, _ := installMap["account"].(map[string]interface{})
		login, _ := account["login"].(string)
		accType, _ := account["type"].(string)

		id := cuid2.Generate()
		db.AddGitHubInstallation(h.DB, id, instID, login, accType, "")
	}

	w.Write([]byte(`{"ok":true}`))
}
