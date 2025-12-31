package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	instance *sql.DB
	once     sync.Once
)

func GetDB(dbPath string) (*sql.DB, error) {
	var initErr error

	once.Do(func() {
		path := dbPath
		if path == "" {
			dataDir := os.Getenv("FSD_DATA_DIR")
			if dataDir == "" {
				dataDir = ".fsd"
			}
			if err := os.MkdirAll(dataDir, 0755); err != nil {
				initErr = err
				return
			}
			path = filepath.Join(dataDir, "data.sqlite")
		}

		db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=on")
		if err != nil {
			initErr = err
			return
		}

		if err := initSchema(db); err != nil {
			initErr = err
			return
		}

		instance = db
	})

	return instance, initErr
}

func initSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		username TEXT UNIQUE,
		avatarUrl TEXT,
		passwordHash TEXT,
		emailVerified INTEGER NOT NULL DEFAULT 0,
		banned INTEGER DEFAULT 0,
		createdAt TEXT,
		lastLoginAt TEXT
	);

	CREATE TABLE IF NOT EXISTS sites (
		id TEXT PRIMARY KEY,
		userId TEXT,
		name TEXT NOT NULL,
		domain TEXT,
		gitUrl TEXT,
		gitBranch TEXT,
		gitSubdir TEXT,
		path TEXT,
		envJson TEXT,
		configJson TEXT,
		envText TEXT,
		buildCommand TEXT,
		outputDir TEXT,
		createdAt TEXT,
		currentDeploymentId TEXT,
		FOREIGN KEY(userId) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS deployments (
		id TEXT PRIMARY KEY,
		userId TEXT,
		siteId TEXT,
		createdAt TEXT,
		status TEXT,
		image TEXT,
		containerName TEXT,
		containerId TEXT,
		hostPort INTEGER,
		containerPort INTEGER,
		url TEXT,
		logsPath TEXT,
		commitSha TEXT,
		commitMessage TEXT,
		commitAuthor TEXT,
		commitAvatar TEXT,
		FOREIGN KEY(userId) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY(siteId) REFERENCES sites(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS oauthAccounts (
		id TEXT PRIMARY KEY,
		provider TEXT,
		providerAccountId TEXT,
		displayName TEXT,
		userId TEXT,
		accessToken TEXT,
		createdAt TEXT,
		FOREIGN KEY(userId) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS emailVerifications (
		id TEXT PRIMARY KEY,
		userId TEXT,
		token TEXT UNIQUE,
		newEmail TEXT,
		createdAt TEXT,
		expiresAt INTEGER,
		usedAt TEXT,
		FOREIGN KEY(userId) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS customDomains (
		id TEXT PRIMARY KEY,
		siteId TEXT,
		hostname TEXT NOT NULL,
		cfCustomHostnameId TEXT,
		status TEXT NOT NULL DEFAULT 'pending',
		sslStatus TEXT,
		verificationRecords TEXT,
		createdAt TEXT,
		FOREIGN KEY(siteId) REFERENCES sites(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS apiKeys (
		id TEXT PRIMARY KEY,
		userId TEXT NOT NULL,
		name TEXT NOT NULL,
		keyHash TEXT NOT NULL,
		keyPrefix TEXT NOT NULL,
		createdAt TEXT,
		lastUsedAt TEXT,
		FOREIGN KEY(userId) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS bannedIPs (
		ip TEXT PRIMARY KEY,
		reason TEXT,
		userId TEXT,
		createdAt TEXT
	);

	CREATE TABLE IF NOT EXISTS userIPs (
		id TEXT PRIMARY KEY,
		userId TEXT,
		ipHash TEXT NOT NULL,
		createdAt TEXT,
		FOREIGN KEY(userId) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS githubAppInstallations (
		id TEXT PRIMARY KEY,
		odId TEXT,
		userId TEXT,
		installationId TEXT NOT NULL,
		accountLogin TEXT,
		accountType TEXT,
		createdAt TEXT,
		FOREIGN KEY(userId) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS atprotoStates (
		key TEXT PRIMARY KEY,
		internalStateJson TEXT,
		createdAt TEXT,
		expiresAt INTEGER
	);

	CREATE TABLE IF NOT EXISTS atprotoSessions (
		sub TEXT PRIMARY KEY,
		sessionJson TEXT,
		updatedAt TEXT
	);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return err
	}

	db.Exec(`ALTER TABLE customDomains ADD COLUMN cfCustomHostnameId TEXT`)

	db.Exec(`ALTER TABLE customDomains ADD COLUMN sslStatus TEXT`)

	db.Exec(`ALTER TABLE deployments ADD COLUMN commitAvatar TEXT`)

	return nil
}
