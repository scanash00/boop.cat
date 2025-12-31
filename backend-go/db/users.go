// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package db

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserFull struct {
	ID            string
	Email         string
	Username      sql.NullString
	AvatarURL     sql.NullString
	PasswordHash  sql.NullString
	EmailVerified bool
	Banned        bool
	CreatedAt     string
	LastLoginAt   sql.NullString
}

func CreateUser(db *sql.DB, id, email, password string) (*UserFull, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = db.Exec(`
		INSERT INTO users (id, email, passwordHash, createdAt, emailVerified, banned)
		VALUES (?, ?, ?, ?, 0, 0)
	`, id, email, string(passwordHash), now)
	if err != nil {
		return nil, err
	}

	return &UserFull{
		ID:            id,
		Email:         email,
		PasswordHash:  sql.NullString{String: string(passwordHash), Valid: true},
		CreatedAt:     now,
		EmailVerified: false,
		Banned:        false,
	}, nil
}

func GetUserByEmail(db *sql.DB, email string) (*UserFull, error) {
	var u UserFull
	var emailVerified, banned int
	err := db.QueryRow(`
		SELECT id, email, username, avatarUrl, passwordHash, emailVerified, banned, createdAt, lastLoginAt
		FROM users WHERE email = ?
	`, email).Scan(&u.ID, &u.Email, &u.Username, &u.AvatarURL, &u.PasswordHash, &emailVerified, &banned, &u.CreatedAt, &u.LastLoginAt)
	if err != nil {
		return nil, err
	}
	u.EmailVerified = emailVerified != 0
	u.Banned = banned != 0
	return &u, nil
}

func GetUserByID(db *sql.DB, id string) (*UserFull, error) {
	var u UserFull
	var emailVerified, banned int
	err := db.QueryRow(`
		SELECT id, email, username, avatarUrl, passwordHash, emailVerified, banned, createdAt, lastLoginAt
		FROM users WHERE id = ?
	`, id).Scan(&u.ID, &u.Email, &u.Username, &u.AvatarURL, &u.PasswordHash, &emailVerified, &banned, &u.CreatedAt, &u.LastLoginAt)
	if err != nil {
		return nil, err
	}
	u.EmailVerified = emailVerified != 0
	u.Banned = banned != 0
	return &u, nil
}

func UpdateLastLogin(db *sql.DB, userID string) error {
	_, err := db.Exec(`UPDATE users SET lastLoginAt = ? WHERE id = ?`, time.Now().UTC().Format(time.RFC3339), userID)
	return err
}

func (u *UserFull) VerifyPassword(password string) bool {
	if !u.PasswordHash.Valid {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash.String), []byte(password))
	return err == nil
}

func SetUserBanned(db *sql.DB, userID string, banned bool) error {
	b := 0
	if banned {
		b = 1
	}
	_, err := db.Exec(`UPDATE users SET banned = ? WHERE id = ?`, b, userID)
	return err
}
