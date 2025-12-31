package db

import (
	"database/sql"
	"encoding/json"
	"time"
)

type CustomDomain struct {
	ID                  string
	SiteID              string
	Hostname            string
	CFCustomHostnameID  sql.NullString
	Status              string
	SSLStatus           sql.NullString
	VerificationRecords sql.NullString
	CreatedAt           string
}

type CustomDomainResponse struct {
	ID                  string      `json:"id"`
	SiteID              string      `json:"siteId"`
	Hostname            string      `json:"hostname"`
	CfID                string      `json:"cfId,omitempty"`
	Status              string      `json:"status"`
	SSLStatus           string      `json:"sslStatus"`
	VerificationRecords interface{} `json:"verificationRecords"`
	CreatedAt           string      `json:"createdAt"`
}

func (d *CustomDomain) ToResponse() CustomDomainResponse {
	resp := CustomDomainResponse{
		ID:        d.ID,
		SiteID:    d.SiteID,
		Hostname:  d.Hostname,
		Status:    d.Status,
		CreatedAt: d.CreatedAt,
	}
	if d.CFCustomHostnameID.Valid {
		resp.CfID = d.CFCustomHostnameID.String
	}
	if d.SSLStatus.Valid {
		resp.SSLStatus = d.SSLStatus.String
	}
	if d.VerificationRecords.Valid && d.VerificationRecords.String != "" {
		json.Unmarshal([]byte(d.VerificationRecords.String), &resp.VerificationRecords)
	} else {
		resp.VerificationRecords = []interface{}{}
	}
	return resp
}

func CreateCustomDomain(db *sql.DB, id, siteID, hostname, cfID, status, sslStatus, records string) error {
	_, err := db.Exec(`
		INSERT INTO customDomains (id, siteId, hostname, cfCustomHostnameId, status, sslStatus, verificationRecords, createdAt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, id, siteID, hostname, toNull(cfID), status, toNull(sslStatus), toNull(records), time.Now().UTC().Format(time.RFC3339))
	return err
}

func GetCustomDomainByID(db *sql.DB, id string) (*CustomDomain, error) {
	var d CustomDomain
	err := db.QueryRow(`
		SELECT id, siteId, hostname, cfCustomHostnameId, status, sslStatus, verificationRecords, createdAt
		FROM customDomains WHERE id = ?
	`, id).Scan(&d.ID, &d.SiteID, &d.Hostname, &d.CFCustomHostnameID, &d.Status, &d.SSLStatus, &d.VerificationRecords, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func GetCustomDomainByHostname(db *sql.DB, hostname string) (*CustomDomain, error) {
	var d CustomDomain
	err := db.QueryRow(`
		SELECT id, siteId, hostname, cfCustomHostnameId, status, sslStatus, verificationRecords, createdAt
		FROM customDomains WHERE hostname = ?
	`, hostname).Scan(&d.ID, &d.SiteID, &d.Hostname, &d.CFCustomHostnameID, &d.Status, &d.SSLStatus, &d.VerificationRecords, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func ListCustomDomains(db *sql.DB, siteID string) ([]CustomDomain, error) {
	rows, err := db.Query(`
		SELECT id, siteId, hostname, cfCustomHostnameId, status, sslStatus, verificationRecords, createdAt
		FROM customDomains WHERE siteId = ?
		ORDER BY createdAt ASC
	`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []CustomDomain
	for rows.Next() {
		var d CustomDomain
		if err := rows.Scan(&d.ID, &d.SiteID, &d.Hostname, &d.CFCustomHostnameID, &d.Status, &d.SSLStatus, &d.VerificationRecords, &d.CreatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, nil
}

func UpdateCustomDomainStatus(db *sql.DB, id, status, sslStatus, records string, cfID string) error {

	if cfID != "" {
		_, err := db.Exec(`
			UPDATE customDomains SET status = ?, sslStatus = ?, verificationRecords = ?, cfCustomHostnameId = ? WHERE id = ?
		`, status, sslStatus, records, cfID, id)
		return err
	}
	_, err := db.Exec(`
		UPDATE customDomains SET status = ?, sslStatus = ?, verificationRecords = ? WHERE id = ?
	`, status, sslStatus, records, id)
	return err
}

func DeleteCustomDomain(db *sql.DB, id string) error {
	_, err := db.Exec("DELETE FROM customDomains WHERE id = ?", id)
	return err
}

func CountCustomDomainsForUser(db *sql.DB, userID string) (int, error) {

	var count int
	err := db.QueryRow(`
		SELECT COUNT(cd.id)
		FROM customDomains cd
		JOIN sites s ON cd.siteId = s.id
		WHERE s.userId = ?
	`, userID).Scan(&count)
	return count, err
}
