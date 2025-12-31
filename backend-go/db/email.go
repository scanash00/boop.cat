// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package db

import (
	"database/sql"
	"time"
)

type EmailVerification struct {
	ID        string
	UserID    string
	Token     string
	NewEmail  sql.NullString
	CreatedAt string
	ExpiresAt int64
	UsedAt    sql.NullString
}

func CreateVerificationToken(db *sql.DB, id, userID, token string, expiresAt int64) error {
	_, err := db.Exec(`
		INSERT INTO emailVerifications (id, userId, token, createdAt, expiresAt)
		VALUES (?, ?, ?, ?, ?)
	`, id, userID, token, time.Now().UTC().Format(time.RFC3339), expiresAt)
	return err
}

func GetVerificationToken(db *sql.DB, token string) (*EmailVerification, error) {
	var ev EmailVerification
	err := db.QueryRow(`
		SELECT id, userId, token, newEmail, createdAt, expiresAt, usedAt
		FROM emailVerifications 
		WHERE token = ? AND usedAt IS NULL
	`, token).Scan(&ev.ID, &ev.UserID, &ev.Token, &ev.NewEmail, &ev.CreatedAt, &ev.ExpiresAt, &ev.UsedAt)
	if err != nil {
		return nil, err
	}
	return &ev, nil
}

func MarkTokenUsed(db *sql.DB, id string) error {
	_, err := db.Exec(`UPDATE emailVerifications SET usedAt = ? WHERE id = ?`, time.Now().UTC().Format(time.RFC3339), id)
	return err
}

func UpdateUserEmailVerified(db *sql.DB, userID string) error {
	_, err := db.Exec(`UPDATE users SET emailVerified = 1 WHERE id = ?`, userID)
	return err
}

func UpdateUserPassword(db *sql.DB, userID, passwordHash string) error {
	_, err := db.Exec(`UPDATE users SET passwordHash = ? WHERE id = ?`, passwordHash, userID)
	return err
}
