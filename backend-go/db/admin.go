package db

import (
	"database/sql"
	"time"
)

type BannedIP struct {
	IP        string
	Reason    string
	UserID    string
	CreatedAt string
}

func BanIP(db *sql.DB, ip, userID, reason string) error {

	var exists string
	err := db.QueryRow(`SELECT ip FROM banned_ips WHERE ip = ?`, ip).Scan(&exists)
	if err == nil {
		return nil
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = db.Exec(`
		INSERT INTO banned_ips (ip, reason, userId, createdAt)
		VALUES (?, ?, ?, ?)
	`, ip, reason, userID, now)
	return err
}

func IsIPBanned(db *sql.DB, ip string) (bool, string) {
	var reason string
	err := db.QueryRow(`SELECT reason FROM banned_ips WHERE ip = ?`, ip).Scan(&reason)
	if err != nil {
		return false, ""
	}
	return true, reason
}
