// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package db

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Site struct {
	ID                  string
	UserID              string
	Name                string
	Domain              string
	GitURL              sql.NullString
	GitBranch           sql.NullString
	GitSubdir           sql.NullString
	Path                sql.NullString
	EnvText             sql.NullString
	BuildCommand        sql.NullString
	OutputDir           sql.NullString
	CreatedAt           string
	CurrentDeploymentID sql.NullString
}

type SiteResponse struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	Domain              string  `json:"domain"`
	GitURL              *string `json:"gitUrl"`
	GitBranch           *string `json:"gitBranch"`
	GitSubdir           *string `json:"gitSubdir,omitempty"`
	BuildCommand        *string `json:"buildCommand,omitempty"`
	OutputDir           *string `json:"outputDir,omitempty"`
	CreatedAt           string  `json:"createdAt"`
	CurrentDeploymentID *string `json:"currentDeploymentId"`
	EnvText             string  `json:"envText,omitempty"`
}

func (s *Site) ToResponse() SiteResponse {
	resp := SiteResponse{
		ID:        s.ID,
		Name:      s.Name,
		Domain:    s.Domain,
		CreatedAt: s.CreatedAt,
	}
	if s.EnvText.Valid {
		resp.EnvText = s.EnvText.String
	}
	if s.GitURL.Valid {
		resp.GitURL = &s.GitURL.String
	}
	if s.GitBranch.Valid {
		resp.GitBranch = &s.GitBranch.String
	}
	if s.GitSubdir.Valid && s.GitSubdir.String != "" {
		resp.GitSubdir = &s.GitSubdir.String
	}
	if s.BuildCommand.Valid && s.BuildCommand.String != "" {
		resp.BuildCommand = &s.BuildCommand.String
	}
	if s.OutputDir.Valid && s.OutputDir.String != "" {
		resp.OutputDir = &s.OutputDir.String
	}
	if s.CurrentDeploymentID.Valid {
		resp.CurrentDeploymentID = &s.CurrentDeploymentID.String
	}
	return resp
}

func ListSites(db *sql.DB, userID string) ([]Site, error) {
	rows, err := db.Query(`
		SELECT id, userId, name, domain, gitUrl, gitBranch, gitSubdir, 
		       path, envText, buildCommand, outputDir, createdAt, currentDeploymentId
		FROM sites WHERE userId = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []Site
	for rows.Next() {
		var s Site
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Domain, &s.GitURL, &s.GitBranch,
			&s.GitSubdir, &s.Path, &s.EnvText, &s.BuildCommand, &s.OutputDir, &s.CreatedAt, &s.CurrentDeploymentID); err != nil {
			return nil, err
		}
		sites = append(sites, s)
	}
	return sites, rows.Err()
}

func GetAllSites(db *sql.DB) ([]Site, error) {
	rows, err := db.Query(`
		SELECT id, userId, name, domain, gitUrl, gitBranch, gitSubdir, 
		       path, envText, buildCommand, outputDir, createdAt, currentDeploymentId
		FROM sites
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []Site
	for rows.Next() {
		var s Site
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Domain, &s.GitURL, &s.GitBranch,
			&s.GitSubdir, &s.Path, &s.EnvText, &s.BuildCommand, &s.OutputDir, &s.CreatedAt, &s.CurrentDeploymentID); err != nil {
			return nil, err
		}
		sites = append(sites, s)
	}
	return sites, rows.Err()
}

func GetSiteByID(db *sql.DB, userID, siteID string) (*Site, error) {
	var s Site
	err := db.QueryRow(`
		SELECT id, userId, name, domain, gitUrl, gitBranch, gitSubdir,
		       path, envText, buildCommand, outputDir, createdAt, currentDeploymentId
		FROM sites WHERE id = ? AND userId = ?
	`, siteID, userID).Scan(&s.ID, &s.UserID, &s.Name, &s.Domain, &s.GitURL, &s.GitBranch,
		&s.GitSubdir, &s.Path, &s.EnvText, &s.BuildCommand, &s.OutputDir, &s.CreatedAt, &s.CurrentDeploymentID)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func GetSiteByIDAdmin(db *sql.DB, siteID string) (*Site, error) {
	var s Site
	err := db.QueryRow(`
		SELECT id, userId, name, domain, gitUrl, gitBranch, gitSubdir,
		       path, envText, buildCommand, outputDir, createdAt, currentDeploymentId
		FROM sites WHERE id = ?
	`, siteID).Scan(&s.ID, &s.UserID, &s.Name, &s.Domain, &s.GitURL, &s.GitBranch,
		&s.GitSubdir, &s.Path, &s.EnvText, &s.BuildCommand, &s.OutputDir, &s.CreatedAt, &s.CurrentDeploymentID)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func UpdateSiteCurrentDeployment(db *sql.DB, siteID, deployID string) error {
	_, err := db.Exec(`UPDATE sites SET currentDeploymentId = ? WHERE id = ?`, deployID, siteID)
	return err
}

func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func GetSitesByUserID(db *sql.DB, userID string) ([]Site, error) {
	return ListSites(db, userID)
}

func DeleteSite(db *sql.DB, userID, siteID string) error {

	_, err := db.Exec(`DELETE FROM sites WHERE id = ? AND userId = ?`, siteID, userID)
	return err
}

func GetSiteByDomain(db *sql.DB, domain string) (*Site, error) {
	var s Site
	err := db.QueryRow(`
		SELECT id, userId, name, domain, gitUrl, gitBranch, gitSubdir,
		       path, envText, buildCommand, outputDir, createdAt, currentDeploymentId
		FROM sites WHERE domain = ?
	`, domain).Scan(&s.ID, &s.UserID, &s.Name, &s.Domain, &s.GitURL, &s.GitBranch,
		&s.GitSubdir, &s.Path, &s.EnvText, &s.BuildCommand, &s.OutputDir, &s.CreatedAt, &s.CurrentDeploymentID)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func mustMarshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

func CreateSite(db *sql.DB, id, userID, name, domain, gitUrl, gitBranch, gitSubdir, buildCommand, outputDir string) error {

	toNull := func(s string) sql.NullString {
		if s == "" {
			return sql.NullString{Valid: false}
		}
		return sql.NullString{String: s, Valid: true}
	}

	_, err := db.Exec(`
		INSERT INTO sites (id, userId, name, domain, gitUrl, gitBranch, gitSubdir, buildCommand, outputDir, createdAt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, userID, name, domain, toNull(gitUrl), toNull(gitBranch), toNull(gitSubdir), toNull(buildCommand), toNull(outputDir),
		time.Now().UTC().Format(time.RFC3339))
	return err
}

func UpdateSiteSettings(db *sql.DB, id, name, domain, gitUrl, branch, subdir, buildCmd, outputDir string) error {
	toNull := func(s string) sql.NullString {
		if s == "" {
			return sql.NullString{Valid: false}
		}
		return sql.NullString{String: s, Valid: true}
	}

	_, err := db.Exec(`
		UPDATE sites 
		SET name = ?, domain = ?, gitUrl = ?, gitBranch = ?, gitSubdir = ?, buildCommand = ?, outputDir = ?
		WHERE id = ?
	`, name, domain, toNull(gitUrl), toNull(branch), toNull(subdir), toNull(buildCmd), toNull(outputDir), id)
	return err
}

func UpdateSiteEnv(db *sql.DB, id, envText string) error {

	_, err := db.Exec(`UPDATE sites SET envText = ? WHERE id = ?`, envText, id)
	return err
}
