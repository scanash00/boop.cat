// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"boop-cat/db"
	"boop-cat/middleware"
)

type SimplifiedRepo struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"fullName"`
	CloneURL      string `json:"cloneUrl"`
	HTMLURL       string `json:"htmlUrl"`
	DefaultBranch string `json:"defaultBranch"`
	Private       bool   `json:"private"`
	Description   string `json:"description"`
	Language      string `json:"language"`
	UpdatedAt     string `json:"updatedAt"`
	PushedAt      string `json:"pushedAt"`
}

func (h *AuthHandler) GetGitHubRepos(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	accounts, err := db.ListOAuthAccounts(h.DB, user.ID)
	if err != nil {
		http.Error(w, `{"error":"db-error"}`, http.StatusInternalServerError)
		return
	}

	var accessToken string
	for _, a := range accounts {
		if a.Provider == "github" && a.AccessToken.Valid {
			accessToken = a.AccessToken.String
			break
		}
	}

	if accessToken == "" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"repos":[],"githubConnected":false}`))
		return
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			page = val
		}
	}
	perPage := 30
	if p := r.URL.Query().Get("per_page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			perPage = val
			if perPage > 100 {
				perPage = 100
			}
		}
	}
	searchQuery := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("q")))

	var repos []SimplifiedRepo

	if searchQuery != "" {

		url := "https://api.github.com/user/repos?sort=updated&per_page=100&visibility=all"
		fetched, _, err := fetchGithubRepos(url, accessToken)
		if err != nil {
			http.Error(w, `{"error":"github-api-error"}`, http.StatusBadGateway)
			return
		}

		for _, repo := range fetched {
			if strings.Contains(strings.ToLower(repo.Name), searchQuery) ||
				strings.Contains(strings.ToLower(repo.Description), searchQuery) {
				repos = append(repos, repo)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"repos":           repos,
			"githubConnected": true,
			"page":            1,
			"hasNextPage":     false,
			"hasPrevPage":     false,
		})
		return
	}

	url := fmt.Sprintf("https://api.github.com/user/repos?sort=updated&per_page=%d&page=%d&visibility=all", perPage, page)
	fetched, linkHeader, err := fetchGithubRepos(url, accessToken)
	if err != nil {
		http.Error(w, `{"error":"github-api-error"}`, http.StatusBadGateway)
		return
	}
	repos = fetched

	hasNextPage := strings.Contains(linkHeader, `rel="next"`)
	hasPrevPage := strings.Contains(linkHeader, `rel="prev"`)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"repos":           repos,
		"githubConnected": true,
		"page":            page,
		"perPage":         perPage,
		"hasNextPage":     hasNextPage,
		"hasPrevPage":     hasPrevPage,
	})
}

type githubRepoInternal struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	CloneURL      string `json:"clone_url"`
	HTMLURL       string `json:"html_url"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
	Description   string `json:"description"`
	Language      string `json:"language"`
	UpdatedAt     string `json:"updated_at"`
	PushedAt      string `json:"pushed_at"`
}

func fetchGithubRepos(url, token string) ([]SimplifiedRepo, string, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "free-static-host")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("github api status: %d", resp.StatusCode)
	}

	var raw []githubRepoInternal
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, "", err
	}

	var simplified []SimplifiedRepo
	for _, r := range raw {
		simplified = append(simplified, SimplifiedRepo{
			ID:            r.ID,
			Name:          r.Name,
			FullName:      r.FullName,
			CloneURL:      r.CloneURL,
			HTMLURL:       r.HTMLURL,
			DefaultBranch: r.DefaultBranch,
			Private:       r.Private,
			Description:   r.Description,
			Language:      r.Language,
			UpdatedAt:     r.UpdatedAt,
			PushedAt:      r.PushedAt,
		})
	}

	return simplified, resp.Header.Get("Link"), nil
}
