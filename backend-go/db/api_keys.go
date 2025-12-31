// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"
)

type APIKey struct {
	ID         string         `json:"id"`
	UserID     string         `json:"userId"`
	Name       string         `json:"name"`
	KeyHash    string         `json:"-"`
	KeyPrefix  string         `json:"prefix"`
	CreatedAt  string         `json:"createdAt"`
	LastUsedAt sql.NullString `json:"lastUsedAt"`
}

type User struct {
	ID            string
	Email         string
	Username      sql.NullString
	EmailVerified bool
	Banned        bool
}

func ListAPIKeys(db *sql.DB, userID string) ([]APIKey, error) {
	rows, err := db.Query(`
		SELECT id, userId, name, keyHash, keyPrefix, createdAt, lastUsedAt
		FROM apiKeys WHERE userId = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var k APIKey
		if err := rows.Scan(&k.ID, &k.UserID, &k.Name, &k.KeyHash, &k.KeyPrefix, &k.CreatedAt, &k.LastUsedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func CreateAPIKey(db *sql.DB, id, userID, name, keyHash, keyPrefix string) error {
	_, err := db.Exec(`
		INSERT INTO apiKeys (id, userId, name, keyHash, keyPrefix, createdAt, lastUsedAt)
		VALUES (?, ?, ?, ?, ?, ?, NULL)
	`, id, userID, name, keyHash, keyPrefix, time.Now().UTC().Format(time.RFC3339))
	return err
}

func DeleteAPIKey(db *sql.DB, userID, keyID string) error {
	result, err := db.Exec(`DELETE FROM apiKeys WHERE id = ? AND userId = ?`, keyID, userID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func CountAPIKeys(db *sql.DB, userID string) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM apiKeys WHERE userId = ?`, userID).Scan(&count)
	return count, err
}

func ValidateAPIKey(db *sql.DB, key string) (*User, string, error) {

	hash := sha256.Sum256([]byte(key))
	keyHash := hex.EncodeToString(hash[:])

	var apiKey APIKey
	err := db.QueryRow(`
		SELECT id, userId, name, keyHash, keyPrefix, createdAt, lastUsedAt
		FROM apiKeys WHERE keyHash = ?
	`, keyHash).Scan(&apiKey.ID, &apiKey.UserID, &apiKey.Name, &apiKey.KeyHash, &apiKey.KeyPrefix, &apiKey.CreatedAt, &apiKey.LastUsedAt)
	if err != nil {
		return nil, "", err
	}

	var user User
	var emailVerified, banned int
	err = db.QueryRow(`
		SELECT id, email, username, emailVerified, banned
		FROM users WHERE id = ?
	`, apiKey.UserID).Scan(&user.ID, &user.Email, &user.Username, &emailVerified, &banned)
	if err != nil {
		return nil, "", err
	}
	user.EmailVerified = emailVerified != 0
	user.Banned = banned != 0

	if user.Banned {
		return nil, "", sql.ErrNoRows
	}
	if !user.EmailVerified {
		return nil, "", sql.ErrNoRows
	}

	_, _ = db.Exec(`UPDATE apiKeys SET lastUsedAt = ? WHERE id = ?`,
		time.Now().UTC().Format(time.RFC3339), apiKey.ID)

	return &user, apiKey.ID, nil
}
