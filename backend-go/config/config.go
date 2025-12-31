package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port       int
	Env        string
	TrustProxy bool

	DBPath string

	SessionSecret string
	CookieSecure  bool

	DeliveryMode   string
	EdgeRootDomain string

	B2KeyID      string
	B2AppKey     string
	B2BucketID   string
	B2BucketName string

	CFAPIToken   string
	CFAccountID  string
	CFZoneID     string
	CFKVNSRoutes string
	CFKVNSFBound string

	GitHubClientID     string
	GitHubClientSecret string
	GoogleClientID     string
	GoogleClientSecret string
	AtprotoPrivateKey  string

	AdminAPIKey string

	RateAPIV1WindowMs int
	RateAPIV1Max      int
}

func Load() *Config {
	return &Config{
		Port:       getEnvInt("PORT", 8787),
		Env:        getEnv("NODE_ENV", "development"),
		TrustProxy: getEnvBool("TRUST_PROXY", false),

		DBPath: getEnv("FSD_DB_PATH", ""),

		SessionSecret: getEnv("SESSION_SECRET", ""),
		CookieSecure:  getEnvBool("COOKIE_SECURE", false),

		DeliveryMode:   strings.ToLower(getEnv("FSD_DELIVERY", "")),
		EdgeRootDomain: strings.ToLower(strings.TrimSpace(getEnv("FSD_EDGE_ROOT_DOMAIN", ""))),

		B2KeyID:      getEnv("B2_KEY_ID", ""),
		B2AppKey:     getEnv("B2_APP_KEY", ""),
		B2BucketID:   getEnv("B2_BUCKET_ID", ""),
		B2BucketName: getEnv("B2_BUCKET_NAME", ""),

		CFAPIToken:   getEnv("CF_API_TOKEN", ""),
		CFAccountID:  getEnv("CF_ACCOUNT_ID", ""),
		CFZoneID:     getEnv("CF_ZONE_ID", ""),
		CFKVNSRoutes: getEnv("CF_KV_NS_ROUTES", ""),
		CFKVNSFBound: getEnv("CF_KV_NS_FBOUND", ""),

		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		AtprotoPrivateKey:  getEnv("ATPROTO_PRIVATE_KEY_1", ""),

		AdminAPIKey: getEnv("ADMIN_API_KEY", ""),

		RateAPIV1WindowMs: getEnvInt("RATE_API_V1_WINDOW_MS", 15*60*1000),
		RateAPIV1Max:      getEnvInt("RATE_API_V1_MAX", 100),
	}
}

func (c *Config) IsProd() bool {
	return c.Env == "production"
}

func (c *Config) EdgeEnabled() bool {
	return c.DeliveryMode == "edge" && c.EdgeRootDomain != ""
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		return v == "1" || v == "true"
	}
	return fallback
}
