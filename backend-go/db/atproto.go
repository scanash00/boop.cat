package db

import (
	"database/sql"
	"strings"
	"time"

	"github.com/nrednav/cuid2"
)

type ATProtoUserResult struct {
	User   *UserFull
	Error  string
	Linked bool
}

func FindOrCreateUserFromAtproto(db *sql.DB, did, handle, email, avatar, linkToUserID string) (*ATProtoUserResult, error) {

	var existingUserID string
	var existingDisplayName sql.NullString
	err := db.QueryRow(`SELECT userId, displayName FROM oauthAccounts WHERE provider = 'atproto' AND providerAccountId = ?`, did).Scan(&existingUserID, &existingDisplayName)

	if err == nil {

		if linkToUserID != "" && existingUserID != linkToUserID {
			return &ATProtoUserResult{Error: "oauth-account-already-linked"}, nil
		}

		user, err := GetUserByID(db, existingUserID)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return &ATProtoUserResult{User: nil}, nil
		}

		if handle != "" && (!existingDisplayName.Valid || existingDisplayName.String == "") {
			db.Exec(`UPDATE oauthAccounts SET displayName = ? WHERE provider = 'atproto' AND providerAccountId = ?`, handle, did)
		}

		updates := false
		if (!user.Username.Valid || user.Username.String == "" || user.Username.String == did) && handle != "" {
			user.Username = sql.NullString{String: handle, Valid: true}
			updates = true
		}
		if (!user.AvatarURL.Valid || user.AvatarURL.String == "") && avatar != "" {
			user.AvatarURL = sql.NullString{String: avatar, Valid: true}
			updates = true
		}

		if email != "" && (isTempEmail(user.Email) || strings.HasPrefix(user.Email, did)) {

			existing, _ := GetUserByEmail(db, email)
			if existing == nil || existing.ID == user.ID {
				user.Email = email
				updates = true
			}
		}

		if updates {

			_, _ = db.Exec(`UPDATE users SET username = ?, avatarUrl = ?, email = ? WHERE id = ?`,
				user.Username, user.AvatarURL, user.Email, user.ID)
		}

		UpdateLastLogin(db, user.ID)

		return &ATProtoUserResult{User: user}, nil
	} else if err != sql.ErrNoRows {
		return nil, err
	}

	if linkToUserID != "" {

		user, err := GetUserByID(db, linkToUserID)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return &ATProtoUserResult{Error: "user-not-found"}, nil
		}

		id := cuid2.Generate()
		now := time.Now().UTC().Format(time.RFC3339)
		_, err = db.Exec(`INSERT INTO oauthAccounts (id, provider, providerAccountId, displayName, userId, createdAt) VALUES (?, ?, ?, ?, ?, ?)`,
			id, "atproto", did, handle, user.ID, now)
		if err != nil {
			return nil, err
		}

		return &ATProtoUserResult{User: user, Linked: true}, nil
	}

	finalEmail := email
	if finalEmail == "" {
		if handle != "" {
			finalEmail = handle + "@atproto.local"
		} else {
			finalEmail = did + "@atproto.local"
		}
	}

	if email != "" {
		existing, _ := GetUserByEmail(db, email)
		if existing != nil {
			finalEmail = did + "@atproto.local"
		}
	}

	username := handle
	if username == "" {
		username = did
	}

	uid := cuid2.Generate()
	now := time.Now().UTC().Format(time.RFC3339)

	emailVerified := 0
	if email != "" {
		emailVerified = 1
	}

	_, err = db.Exec(`INSERT INTO users (id, email, username, avatarUrl, emailVerified, createdAt, lastLoginAt) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		uid, finalEmail, username, avatar, emailVerified, now, now)
	if err != nil {
		return nil, err
	}

	oauthID := cuid2.Generate()
	_, err = db.Exec(`INSERT INTO oauthAccounts (id, provider, providerAccountId, displayName, userId, createdAt) VALUES (?, ?, ?, ?, ?, ?)`,
		oauthID, "atproto", did, handle, uid, now)
	if err != nil {
		return nil, err
	}

	user, _ := GetUserByID(db, uid)
	return &ATProtoUserResult{User: user}, nil
}

func isTempEmail(email string) bool {

	return len(email) > 14 && email[len(email)-14:] == "@atproto.local"
}
