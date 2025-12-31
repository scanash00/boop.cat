package db

import (
	"database/sql"
	"strings"
	"time"
)

type GitHubInstallation struct {
	ID             string
	OdID           sql.NullString
	UserID         sql.NullString
	InstallationID string
	AccountLogin   sql.NullString
	AccountType    sql.NullString
	CreatedAt      string
}

func AddGitHubInstallation(db *sql.DB, id, installationID, accountLogin, accountType string, userID string) error {

	var exists string
	err := db.QueryRow("SELECT id FROM githubAppInstallations WHERE installationId = ?", installationID).Scan(&exists)
	if err == nil {

		if userID != "" {
			_, err = db.Exec("UPDATE githubAppInstallations SET userId = ? WHERE installationId = ?", userID, installationID)
		}
		return err
	}

	_, err = db.Exec(`
		INSERT INTO githubAppInstallations (id, userId, installationId, accountLogin, accountType, createdAt)
		VALUES (?, ?, ?, ?, ?, ?)
	`, id, toNull(userID), installationID, toNull(accountLogin), toNull(accountType), time.Now().UTC().Format(time.RFC3339))
	return err
}

func RemoveGitHubInstallation(db *sql.DB, installationID string) error {
	_, err := db.Exec("DELETE FROM githubAppInstallations WHERE installationId = ?", installationID)
	return err
}

func GetSitesByRepo(db *sql.DB, repoURL, branch string) ([]Site, error) {

	rows, err := db.Query(`
		SELECT id, userId, name, domain, gitUrl, gitBranch, gitSubdir, 
		       path, buildCommand, outputDir, createdAt, currentDeploymentId
		FROM sites WHERE gitUrl IS NOT NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []Site

	clean := func(u string) string {
		return strings.TrimSuffix(strings.ToLower(u), ".git")
	}
	target := clean(repoURL)
	targetBranch := branch

	for rows.Next() {
		var s Site
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Domain, &s.GitURL, &s.GitBranch,
			&s.GitSubdir, &s.Path, &s.BuildCommand, &s.OutputDir, &s.CreatedAt, &s.CurrentDeploymentID); err != nil {
			continue
		}

		if !s.GitURL.Valid || s.GitURL.String == "" {
			continue
		}

		siteBranch := "main"
		if s.GitBranch.Valid && s.GitBranch.String != "" {
			siteBranch = s.GitBranch.String
		}
		if siteBranch != targetBranch {
			continue
		}

		siteGit := clean(s.GitURL.String)

		if siteGit == target || strings.HasSuffix(target, siteGit) || strings.HasSuffix(siteGit, target) {
			matches = append(matches, s)
		}
	}
	return matches, nil
}

func toNull(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
