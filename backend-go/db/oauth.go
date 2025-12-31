package db

import (
	"database/sql"
	"time"
)

type OAuthAccount struct {
	ID                string
	Provider          string
	ProviderAccountID string
	DisplayName       sql.NullString
	UserID            string
	AccessToken       sql.NullString
	CreatedAt         string
}

func CreateOAuthAccount(db *sql.DB, id, provider, providerAccountID, userID, accessToken, displayName string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.Exec(`
		INSERT INTO oauthAccounts (id, provider, providerAccountId, userId, accessToken, displayName, createdAt)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, provider, providerAccountID, userID, accessToken, displayName, now)
	return err
}

func FindOAuthAccount(db *sql.DB, provider, providerAccountID string) (*OAuthAccount, error) {
	var acc OAuthAccount
	err := db.QueryRow(`
		SELECT id, provider, providerAccountId, displayName, userId, accessToken, createdAt
		FROM oauthAccounts WHERE provider = ? AND providerAccountId = ?
	`, provider, providerAccountID).Scan(&acc.ID, &acc.Provider, &acc.ProviderAccountID, &acc.DisplayName, &acc.UserID, &acc.AccessToken, &acc.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

func ListOAuthAccounts(db *sql.DB, userID string) ([]OAuthAccount, error) {
	rows, err := db.Query(`
		SELECT id, provider, providerAccountId, displayName, userId, accessToken, createdAt
		FROM oauthAccounts WHERE userId = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []OAuthAccount
	for rows.Next() {
		var acc OAuthAccount
		if err := rows.Scan(&acc.ID, &acc.Provider, &acc.ProviderAccountID, &acc.DisplayName, &acc.UserID, &acc.AccessToken, &acc.CreatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, rows.Err()
}

func DeleteOAuthAccount(db *sql.DB, userID, accountID string) error {
	_, err := db.Exec(`DELETE FROM oauthAccounts WHERE id = ? AND userId = ?`, accountID, userID)
	return err
}

func GetGitHubToken(db *sql.DB, userID string) (string, error) {
	var token sql.NullString
	err := db.QueryRow(`
		SELECT accessToken FROM oauthAccounts WHERE userId = ? AND provider = 'github' LIMIT 1
	`, userID).Scan(&token)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if token.Valid {
		return token.String, nil
	}
	return "", nil
}
