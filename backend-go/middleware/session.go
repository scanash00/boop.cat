package middleware

import (
	"context"
	"database/sql"
	"net/http"

	"boop-cat/db"
	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

func InitSessionStore(secret string, secure bool) {
	store = sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

func GetSession(r *http.Request) (*sessions.Session, error) {
	return store.Get(r, "fsd-session")
}

func WithUser(database *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := GetSession(r)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			userID, ok := session.Values["userId"].(string)
			if !ok || userID == "" {
				next.ServeHTTP(w, r)
				return
			}

			user, err := db.GetUserByID(database, userID)
			if err == nil && user != nil && !user.Banned {

				u := &db.User{
					ID:            user.ID,
					Email:         user.Email,
					Username:      user.Username,
					EmailVerified: user.EmailVerified,
					Banned:        user.Banned,
				}
				ctx := context.WithValue(r.Context(), UserContextKey, u)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

func LoginUser(w http.ResponseWriter, r *http.Request, userID string) error {
	session, _ := GetSession(r)
	session.Values["userId"] = userID
	return session.Save(r, w)
}

func LogoutUser(w http.ResponseWriter, r *http.Request) error {
	session, _ := GetSession(r)
	session.Values["userId"] = ""
	session.Options.MaxAge = -1
	return session.Save(r, w)
}

func RequireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r.Context())
		if user == nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
