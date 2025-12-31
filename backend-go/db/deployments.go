package db

import (
	"database/sql"
	"time"
)

type Deployment struct {
	ID            string
	UserID        string
	SiteID        string
	CreatedAt     string
	Status        string
	URL           sql.NullString
	CommitSha     sql.NullString
	CommitMessage sql.NullString
	CommitAuthor  sql.NullString
	CommitAvatar  sql.NullString
	LogsPath      sql.NullString
}

type DeploymentResponse struct {
	ID            string  `json:"id"`
	Status        string  `json:"status"`
	URL           *string `json:"url"`
	CreatedAt     string  `json:"createdAt"`
	CommitSha     *string `json:"commitSha"`
	CommitMessage *string `json:"commitMessage"`
	CommitAuthor  *string `json:"commitAuthor"`
	CommitAvatar  *string `json:"commitAvatar"`
}

func (d *Deployment) ToResponse() DeploymentResponse {
	resp := DeploymentResponse{
		ID:        d.ID,
		Status:    d.Status,
		CreatedAt: d.CreatedAt,
	}
	if d.URL.Valid {
		resp.URL = &d.URL.String
	}
	if d.CommitSha.Valid {
		resp.CommitSha = &d.CommitSha.String
	}
	if d.CommitMessage.Valid {
		resp.CommitMessage = &d.CommitMessage.String
	}
	if d.CommitAuthor.Valid {
		resp.CommitAuthor = &d.CommitAuthor.String
	}
	if d.CommitAvatar.Valid {
		resp.CommitAvatar = &d.CommitAvatar.String
	}
	return resp
}

func CreateDeployment(db *sql.DB, id, userID, siteID, status string, commitSha, commitMessage, commitAuthor, commitAvatar *string) error {
	_, err := db.Exec(`
		INSERT INTO deployments (id, userId, siteId, createdAt, status, commitSha, commitMessage, commitAuthor, commitAvatar)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, userID, siteID, time.Now().UTC().Format(time.RFC3339), status, commitSha, commitMessage, commitAuthor, commitAvatar)
	return err
}

func UpdateDeploymentStatus(db *sql.DB, id, status, urlStr string) error {
	var urlVal sql.NullString
	if urlStr != "" {
		urlVal = sql.NullString{String: urlStr, Valid: true}
	}
	_, err := db.Exec(`UPDATE deployments SET status = ?, url = ? WHERE id = ?`, status, urlVal, id)
	return err
}

func UpdateDeploymentLogs(db *sql.DB, id, logsPath string) error {
	_, err := db.Exec(`UPDATE deployments SET logsPath = ? WHERE id = ?`, logsPath, id)
	return err
}

func StopOtherDeployments(db *sql.DB, siteID, currentDeployID string) error {
	_, err := db.Exec(`
		UPDATE deployments 
		SET status = 'stopped' 
		WHERE siteId = ? AND id != ? AND status = 'running'
	`, siteID, currentDeployID)
	return err
}

func GetDeploymentByID(db *sql.DB, id string) (*Deployment, error) {
	var d Deployment
	err := db.QueryRow(`
		SELECT id, userId, siteId, createdAt, status, url, commitSha, commitMessage, commitAuthor, commitAvatar, logsPath
		FROM deployments WHERE id = ?
	`, id).Scan(&d.ID, &d.UserID, &d.SiteID, &d.CreatedAt, &d.Status,
		&d.URL, &d.CommitSha, &d.CommitMessage, &d.CommitAuthor, &d.CommitAvatar, &d.LogsPath)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func ListDeployments(db *sql.DB, userID, siteID string) ([]Deployment, error) {
	rows, err := db.Query(`
		SELECT id, userId, siteId, createdAt, status, url, commitSha, commitMessage, commitAuthor, commitAvatar, logsPath
		FROM deployments WHERE userId = ? AND siteId = ?
		ORDER BY createdAt DESC
	`, userID, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []Deployment
	for rows.Next() {
		var d Deployment
		if err := rows.Scan(&d.ID, &d.UserID, &d.SiteID, &d.CreatedAt, &d.Status,
			&d.URL, &d.CommitSha, &d.CommitMessage, &d.CommitAuthor, &d.CommitAvatar, &d.LogsPath); err != nil {
			return nil, err
		}
		deps = append(deps, d)
	}
	return deps, rows.Err()
}
