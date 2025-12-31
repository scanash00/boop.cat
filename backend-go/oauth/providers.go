package oauth

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

func InitProviders(sessionSecret string) {

	store := sessions.NewCookieStore([]byte(sessionSecret))
	store.Options.HttpOnly = true
	store.Options.Secure = os.Getenv("NODE_ENV") == "production"
	gothic.Store = store

	publicURL := os.Getenv("PUBLIC_URL")
	if publicURL == "" {
		publicURL = "http://localhost:8080"
	}

	publicURL = strings.TrimRight(publicURL, "/")

	githubCallback := os.Getenv("GITHUB_CALLBACK_URL")
	if githubCallback == "" {
		githubCallback = fmt.Sprintf("%s/auth/github/callback", publicURL)
	}

	googleCallback := os.Getenv("GOOGLE_CALLBACK_URL")
	if googleCallback == "" {
		googleCallback = fmt.Sprintf("%s/auth/google/callback", publicURL)
	}

	goth.UseProviders(
		github.New(
			os.Getenv("GITHUB_CLIENT_ID"),
			os.Getenv("GITHUB_CLIENT_SECRET"),
			githubCallback,
			"read:user", "user:email", "repo",
		),
		google.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			googleCallback,
			"email", "profile",
		),
	)

	gothic.GetProviderName = func(req *http.Request) (string, error) {
		provider := chi.URLParam(req, "provider")
		if provider == "" {
			return "", nil
		}
		return provider, nil
	}
}
