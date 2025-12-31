// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"boop-cat/db"
)

type ContextKey string

const (
	UserContextKey ContextKey = "user"

	APIKeyIDContextKey ContextKey = "apiKeyId"
)

func RequireAPIKey(database *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"missing-api-key","message":"Authorization header with Bearer token required"}`, http.StatusUnauthorized)
				return
			}

			key := strings.TrimPrefix(authHeader, "Bearer ")
			if !strings.HasPrefix(key, "sk_") {
				http.Error(w, `{"error":"invalid-api-key","message":"Invalid or expired API key"}`, http.StatusUnauthorized)
				return
			}

			user, keyID, err := db.ValidateAPIKey(database, key)
			if err != nil {
				http.Error(w, `{"error":"invalid-api-key","message":"Invalid or expired API key"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			ctx = context.WithValue(ctx, APIKeyIDContextKey, keyID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUser(ctx context.Context) *db.User {
	if user, ok := ctx.Value(UserContextKey).(*db.User); ok {
		return user
	}
	return nil
}

func GetAPIKeyID(ctx context.Context) string {
	if keyID, ok := ctx.Value(APIKeyIDContextKey).(string); ok {
		return keyID
	}
	return ""
}

func GetUserID(ctx context.Context) string {
	user := GetUser(ctx)
	if user != nil {
		return user.ID
	}
	return ""
}
