// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"boop-cat/config"
	"boop-cat/db"
	"boop-cat/handlers"
	"boop-cat/lib"
	"boop-cat/middleware"
	"boop-cat/oauth"
)

func main() {

	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	lib.StartDMCAMonitor()

	cfg := config.Load()

	if cfg.SessionSecret == "" {
		fmt.Fprintln(os.Stderr, "Missing SESSION_SECRET. Generate one: openssl rand -base64 32")
		os.Exit(1)
	}

	database, err := db.GetDB(cfg.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
		os.Exit(1)
	}

	r := chi.NewRouter()

	middleware.InitSessionStore(cfg.SessionSecret, cfg.CookieSecure)

	oauth.InitProviders(cfg.SessionSecret)

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.WithUser(database))
	r.Use(middleware.RateLimit(100, 60*time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"backend":"go"}`))
	})

	r.Get("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"deliveryMode":"%s","edgeRootDomain":"%s"}`,
			cfg.DeliveryMode, cfg.EdgeRootDomain)
	})

	deployHandler := handlers.NewDeployHandler(database)

	authHandler := handlers.NewAuthHandler(database, deployHandler.Engine)
	r.Mount("/api/auth", authHandler.Routes())
	r.Route("/auth", func(r chi.Router) {
		authHandler.MountOAuthRoutes(r)
	})
	r.Get("/api/github/repos", authHandler.GetGitHubRepos)
	r.Get("/github/installed", authHandler.GitHubInstalled)

	apiKeysHandler := handlers.NewAPIKeysHandler(database)
	r.Mount("/api/api-keys", apiKeysHandler.Routes())

	sitesHandler := handlers.NewSitesHandler(database, deployHandler.Engine)

	r.Mount("/api/account", handlers.NewAccountHandler(database).Routes())

	cdHandler := handlers.NewCustomDomainHandler(database, deployHandler.Engine)

	r.Route("/api/sites", func(r chi.Router) {
		r.Use(middleware.RequireLogin)

		r.Get("/", sitesHandler.ListSites)
		r.Post("/", sitesHandler.CreateSite)

		r.Route("/{siteId}", func(r chi.Router) {
			r.Patch("/", sitesHandler.UpdateSiteEnv)
			r.Patch("/settings", sitesHandler.UpdateSiteSettings)
			r.Put("/settings", sitesHandler.UpdateSiteSettings)
			r.Post("/settings", sitesHandler.UpdateSiteSettings)
			r.Delete("/", sitesHandler.DeleteSite)

			r.Post("/deploy", deployHandler.TriggerDeploy)
			r.Get("/deployments", deployHandler.ListDeployments)

			r.Get("/custom-domains", cdHandler.ListCustomDomains)
			r.Post("/custom-domains", cdHandler.CreateCustomDomain)
			r.Delete("/custom-domains/{id}", cdHandler.DeleteCustomDomain)
			r.Post("/custom-domains/{id}/poll", cdHandler.PollCustomDomain)
		})

		r.Post("/preview", deployHandler.PreviewSite)
	})

	r.Route("/api/deployments/{id}", func(r chi.Router) {
		r.Use(middleware.RequireLogin)
		r.Get("/", deployHandler.GetDeployment)
		r.Get("/logs", deployHandler.GetDeploymentLogs)
		r.Post("/stop", deployHandler.StopDeployment)
	})

	r.Delete("/api/account", func(w http.ResponseWriter, r *http.Request) {
		middleware.RequireLogin(http.HandlerFunc(authHandler.DeleteAccount)).ServeHTTP(w, r)
	})

	ghWebhookHandler := handlers.NewGitHubWebhookHandler(database, deployHandler.Engine)
	r.Mount("/api/github/webhook", ghWebhookHandler.Routes())

	apiV1Handler := handlers.NewAPIV1Handler(database, deployHandler.Engine)
	r.Mount("/api/v1", apiV1Handler.Routes())

	adminHandler := handlers.NewAdminHandler(database)
	r.Mount("/api/admin", adminHandler.Routes())

	atprotoHandler := handlers.NewATProtoHandler(database)
	r.Get("/client-metadata.json", atprotoHandler.ServeClientMetadata)
	r.Get("/jwks.json", atprotoHandler.ServeJWKS)
	r.Get("/auth/atproto", atprotoHandler.BeginAuth)
	r.Get("/auth/atproto/callback", atprotoHandler.Callback)

	clientDist := filepath.Join("..", "client", "dist")
	if _, err := os.Stat(clientDist); err == nil {

		fs := http.FileServer(http.Dir(clientDist))
		r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			path := filepath.Join(clientDist, r.URL.Path)
			if _, err := os.Stat(path); os.IsNotExist(err) {

				http.ServeFile(w, r, filepath.Join(clientDist, "index.html"))
				return
			}
			fs.ServeHTTP(w, r)
		}))
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	fmt.Printf("boop.cat (Go) listening on http://127.0.0.1%s\n", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
