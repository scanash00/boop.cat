package handlers

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"boop-cat/db"
	"boop-cat/lib"
	"boop-cat/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
)

type atprotoStateData struct {
	DID          string
	CodeVerifier string
	PDS          string
	LinkToUserID string
	CreatedAt    time.Time
}

var (
	atprotoStates    = make(map[string]*atprotoStateData)
	atprotoStatesMux sync.RWMutex
)

func storeATProtoState(state, did, verifier, pds, linkToUserID string) {
	atprotoStatesMux.Lock()
	defer atprotoStatesMux.Unlock()
	atprotoStates[state] = &atprotoStateData{
		DID:          did,
		CodeVerifier: verifier,
		PDS:          pds,
		LinkToUserID: linkToUserID,
		CreatedAt:    time.Now(),
	}
}

func getATProtoState(state string) *atprotoStateData {
	atprotoStatesMux.RLock()
	defer atprotoStatesMux.RUnlock()
	return atprotoStates[state]
}

func deleteATProtoState(state string) {
	atprotoStatesMux.Lock()
	defer atprotoStatesMux.Unlock()
	delete(atprotoStates, state)
}

type ATProtoHandler struct {
	DB *sql.DB
}

func NewATProtoHandler(database *sql.DB) *ATProtoHandler {
	return &ATProtoHandler{DB: database}
}

func (h *ATProtoHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/client-metadata.json", h.ServeClientMetadata)
	r.Get("/jwks.json", h.ServeJWKS)
	r.Get("/auth/atproto", h.BeginAuth)
	r.Get("/auth/atproto/callback", h.Callback)
	return r
}

func baseUrl(r *http.Request) string {

	publicURL := os.Getenv("PUBLIC_URL")
	if publicURL != "" {
		return publicURL
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	if r.Host == "localhost" || r.Host == "127.0.0.1" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8788"
		}
		return fmt.Sprintf("%s://%s:%s", scheme, r.Host, port)
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

func (h *ATProtoHandler) ServeClientMetadata(w http.ResponseWriter, r *http.Request) {
	b := baseUrl(r)
	clientID := os.Getenv("ATPROTO_CLIENT_ID")
	if clientID == "" {
		clientID = fmt.Sprintf("%s/client-metadata.json", b)
	}

	meta := map[string]interface{}{
		"client_id":                       clientID,
		"client_name":                     os.Getenv("ATPROTO_CLIENT_NAME"),
		"client_uri":                      b,
		"logo_uri":                        os.Getenv("ATPROTO_LOGO_URI"),
		"tos_uri":                         os.Getenv("ATPROTO_TOS_URI"),
		"policy_uri":                      os.Getenv("ATPROTO_POLICY_URI"),
		"redirect_uris":                   []string{fmt.Sprintf("%s/auth/atproto/callback", b)},
		"grant_types":                     []string{"authorization_code", "refresh_token"},
		"scope":                           os.Getenv("ATPROTO_SCOPE"),
		"response_types":                  []string{"code"},
		"application_type":                "web",
		"token_endpoint_auth_method":      "private_key_jwt",
		"token_endpoint_auth_signing_alg": "ES256",
		"dpop_bound_access_tokens":        true,
		"jwks_uri":                        fmt.Sprintf("%s/jwks.json", b),
	}

	if meta["client_name"] == "" || meta["client_name"] == nil {
		meta["client_name"] = "boop.cat"
	}
	if meta["logo_uri"] == "" || meta["logo_uri"] == nil {
		meta["logo_uri"] = fmt.Sprintf("%s/public/logo.svg", b)
	}
	if meta["tos_uri"] == "" || meta["tos_uri"] == nil {
		meta["tos_uri"] = fmt.Sprintf("%s/tos", b)
	}
	if meta["policy_uri"] == "" || meta["policy_uri"] == nil {
		meta["policy_uri"] = fmt.Sprintf("%s/privacy", b)
	}
	if meta["scope"] == "" || meta["scope"] == nil {
		meta["scope"] = "atproto transition:generic"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meta)
}

func (h *ATProtoHandler) ServeJWKS(w http.ResponseWriter, r *http.Request) {
	keyPEM := os.Getenv("ATPROTO_PRIVATE_KEY_1")
	if keyPEM == "" {
		jsonError(w, "atproto-not-configured", http.StatusNotFound)
		return
	}

	block, _ := pem.Decode([]byte(keyPEM))
	if block == nil {
		jsonError(w, "invalid-key-format", http.StatusInternalServerError)
		return
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {

		privKey, err = x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			jsonError(w, "invalid-key-parse", http.StatusInternalServerError)
			return
		}
	}

	ecKey, ok := privKey.(*ecdsa.PrivateKey)
	if !ok {
		jsonError(w, "key-not-ec", http.StatusInternalServerError)
		return
	}

	jwk := jose.JSONWebKey{
		Key:       &ecKey.PublicKey,
		KeyID:     "key1",
		Algorithm: "ES256",
		Use:       "sig",
	}

	jwks := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{jwk},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jwks)
}

type AuthMetadata struct {
	Issuer                    string   `json:"issuer"`
	AuthEndpoint              string   `json:"authorization_endpoint"`
	TokenEndpoint             string   `json:"token_endpoint"`
	PushedAuthRequestEndpoint string   `json:"pushed_authorization_request_endpoint"`
	DPoPSigningAlgs           []string `json:"dpop_signing_alg_values_supported"`
}

func (h *ATProtoHandler) BeginAuth(w http.ResponseWriter, r *http.Request) {
	handle := r.URL.Query().Get("handle")
	if handle == "" {
		http.Redirect(w, r, "/login?error=missing_handle", http.StatusFound)
		return
	}
	handle = strings.TrimSpace(handle)

	did, err := resolveHandle(handle)
	if err != nil {
		fmt.Printf("Resolve handle error: %v\n", err)
		http.Redirect(w, r, "/login?error=handle_resolution_failed", http.StatusFound)
		return
	}

	pds, err := getPDSEndpoint(did)
	if err != nil {
		fmt.Printf("Get PDS error: %v\n", err)
		http.Redirect(w, r, "/login?error=pds_discovery_failed", http.StatusFound)
		return
	}

	authMeta, err := getAuthMetadata(pds)
	if err != nil {
		fmt.Printf("Get Auth Meta error: %v\n", err)
		http.Redirect(w, r, "/login?error=oauth_discovery_failed", http.StatusFound)
		return
	}

	pkce := lib.GeneratePKCE()
	state := lib.GenerateSecureState()

	linkToUserID := ""
	if user := middleware.GetUser(r.Context()); user != nil {
		linkToUserID = user.ID
	}

	storeATProtoState(state, did, pkce.Verifier, pds, linkToUserID)

	redirectURI := fmt.Sprintf("%s/auth/atproto/callback", baseUrl(r))
	clientID := fmt.Sprintf("%s/client-metadata.json", baseUrl(r))

	v := url.Values{}
	v.Set("client_id", clientID)
	v.Set("response_type", "code")
	v.Set("redirect_uri", redirectURI)
	scope := os.Getenv("ATPROTO_SCOPE")
	if scope == "" {
		scope = "atproto transition:generic"
	}
	v.Set("scope", scope)
	v.Set("state", state)
	v.Set("code_challenge", pkce.Challenge)
	v.Set("code_challenge_method", "S256")
	v.Set("login_hint", handle)

	assertion, err := makeClientAssertion(clientID, authMeta.Issuer)
	if err != nil {
		fmt.Printf("Assertion Error: %v\n", err)
		http.Redirect(w, r, "/login?error=assertion_failed", http.StatusFound)
		return
	}
	v.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	v.Set("client_assertion", assertion)

	if authMeta.PushedAuthRequestEndpoint != "" {

		executePAR := func(nonce string) (*http.Response, error) {
			dpop, err := makeDPoPHeader("POST", authMeta.PushedAuthRequestEndpoint, nonce)
			if err != nil {
				return nil, err
			}
			req, _ := http.NewRequest("POST", authMeta.PushedAuthRequestEndpoint, strings.NewReader(v.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("DPoP", dpop)

			client := &http.Client{Timeout: 10 * time.Second}
			return client.Do(req)
		}

		resp, err := executePAR("")
		if err != nil {
			fmt.Printf("PAR Request Error: %v\n", err)
			http.Redirect(w, r, "/login?error=par_request_failed", http.StatusFound)
			return
		}

		if resp.StatusCode == 400 {
			nonce := resp.Header.Get("DPoP-Nonce")
			if nonce != "" {
				resp.Body.Close()
				resp, err = executePAR(nonce)
				if err != nil {
					http.Redirect(w, r, "/login?error=par_retry_failed", http.StatusFound)
					return
				}
			}
		}
		defer resp.Body.Close()

		if resp.StatusCode != 201 && resp.StatusCode != 200 {
			b, _ := io.ReadAll(resp.Body)
			fmt.Printf("PAR Failed: %d %s\n", resp.StatusCode, string(b))
			http.Redirect(w, r, "/login?error=par_failed", http.StatusFound)
			return
		}

		var parRes struct {
			RequestURI string `json:"request_uri"`
			ExpiresIn  int    `json:"expires_in"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&parRes); err != nil {
			http.Redirect(w, r, "/login?error=par_decode_failed", http.StatusFound)
			return
		}

		dest := fmt.Sprintf("%s?client_id=%s&request_uri=%s",
			authMeta.AuthEndpoint, url.QueryEscape(clientID), url.QueryEscape(parRes.RequestURI))
		http.Redirect(w, r, dest, http.StatusFound)
		return
	}

	dest := authMeta.AuthEndpoint + "?" + v.Encode()
	http.Redirect(w, r, dest, http.StatusFound)
}

func (h *ATProtoHandler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		http.Redirect(w, r, "/login?error=missing_params", http.StatusFound)
		return
	}

	stateData := getATProtoState(state)
	if stateData == nil {
		http.Redirect(w, r, "/login?error=invalid_state", http.StatusFound)
		return
	}
	did := stateData.DID

	pds, err := getPDSEndpoint(did)
	if err != nil {
		fmt.Printf("Callback PDS error: %v\n", err)
		http.Redirect(w, r, "/login?error=pds_rediscovery_failed", http.StatusFound)
		return
	}
	meta, err := getAuthMetadata(pds)
	if err != nil {
		fmt.Printf("Callback Metadata error: %v\n", err)
		http.Redirect(w, r, "/login?error=meta_rediscovery_failed", http.StatusFound)
		return
	}

	redirectURI := fmt.Sprintf("%s/auth/atproto/callback", baseUrl(r))
	clientID := fmt.Sprintf("%s/client-metadata.json", baseUrl(r))
	assertion, err := makeClientAssertion(clientID, meta.Issuer)
	if err != nil {
		http.Redirect(w, r, "/login?error=assertion_generation_failed", http.StatusFound)
		return
	}

	v := url.Values{}
	v.Set("grant_type", "authorization_code")
	v.Set("code", code)
	v.Set("redirect_uri", redirectURI)
	v.Set("client_id", clientID)
	v.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	v.Set("client_assertion", assertion)

	if stateData != nil && stateData.CodeVerifier != "" {
		v.Set("code_verifier", stateData.CodeVerifier)
		deleteATProtoState(state)
	}

	executeTokenReq := func(nonce string) (*http.Response, error) {
		dpop, err := makeDPoPHeader("POST", meta.TokenEndpoint, nonce)
		if err != nil {
			return nil, err
		}
		req, _ := http.NewRequest("POST", meta.TokenEndpoint, strings.NewReader(v.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("DPoP", dpop)
		req.Header.Set("Accept", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		return client.Do(req)
	}

	resp, err := executeTokenReq("")
	if err != nil {
		fmt.Printf("Token Req Error: %v\n", err)
		http.Redirect(w, r, "/login?error=token_request_failed", http.StatusFound)
		return
	}

	if resp.StatusCode == 400 {
		nonce := resp.Header.Get("DPoP-Nonce")
		if nonce != "" {
			resp.Body.Close()
			resp, err = executeTokenReq(nonce)
			if err != nil {
				http.Redirect(w, r, "/login?error=token_retry_failed", http.StatusFound)
				return
			}
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		fmt.Printf("Token Req Failed: %d %s\n", resp.StatusCode, string(b))
		http.Redirect(w, r, "/login?error=token_exchange_failed", http.StatusFound)
		return
	}

	var tokenRes struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		Sub         string `json:"sub"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		http.Redirect(w, r, "/login?error=token_decode_failed", http.StatusFound)
		return
	}

	prof, err := fetchPublicProfile(tokenRes.Sub)

	handle := ""
	avatar := ""
	if prof != nil {
		handle = prof.Handle
		avatar = prof.Avatar
	}

	email := ""

	lastNonce := resp.Header.Get("DPoP-Nonce")

	fetchedEmail, err := fetchAtprotoEmail(pds, tokenRes.AccessToken, lastNonce)
	if err == nil && fetchedEmail != "" {
		email = fetchedEmail
	} else if err != nil {
		fmt.Printf("Email fetch warning: %v\n", err)
	}

	linkToUserID := stateData.LinkToUserID

	res, err := db.FindOrCreateUserFromAtproto(h.DB, tokenRes.Sub, handle, email, avatar, linkToUserID)
	if err != nil {
		fmt.Printf("DB Error: %v\n", err)
		http.Redirect(w, r, "/login?error=db_error", http.StatusFound)
		return
	}

	if res.Error != "" {
		if res.Error == "oauth-account-already-linked" {
			http.Redirect(w, r, "/dashboard/account?error=already-linked", http.StatusFound)
			return
		}
		http.Redirect(w, r, "/login?error="+res.Error, http.StatusFound)
		return
	}

	if res.User == nil {
		http.Redirect(w, r, "/login?error=user_creation_failed", http.StatusFound)
		return
	}

	if res.Linked {
		http.Redirect(w, r, "/dashboard/account", http.StatusFound)
		return
	}

	if err := middleware.LoginUser(w, r, res.User.ID); err != nil {
		fmt.Printf("Session Error: %v\n", err)
		http.Redirect(w, r, "/login?error=session_error", http.StatusFound)
		return
	}

	fmt.Printf("Login Success for User: %s\n", res.User.ID)
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

type PublicProfile struct {
	DID    string `json:"did"`
	Handle string `json:"handle"`
	Avatar string `json:"avatar"`
}

func fetchPublicProfile(did string) (*PublicProfile, error) {
	resp, err := http.Get(fmt.Sprintf("https://public.api.bsky.app/xrpc/app.bsky.actor.getProfile?actor=%s", did))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("profile fetch failed")
	}
	var p PublicProfile
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func getPrivateKey() (*ecdsa.PrivateKey, error) {
	keyPEM := os.Getenv("ATPROTO_PRIVATE_KEY_1")
	block, _ := pem.Decode([]byte(keyPEM))
	if block == nil {
		return nil, fmt.Errorf("invalid key")
	}
	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privKey, err = x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	}
	return privKey.(*ecdsa.PrivateKey), nil
}

func makeDPoPHeader(method, targetURL, nonce string) (string, error) {
	return makeDPoPHeaderWithAth(method, targetURL, nonce, "")
}

func makeDPoPHeaderWithAth(method, targetURL, nonce, accessToken string) (string, error) {
	key, err := getPrivateKey()
	if err != nil {
		return "", err
	}

	jwk := jose.JSONWebKey{
		Key:       &key.PublicKey,
		KeyID:     "key1",
		Algorithm: "ES256",
		Use:       "sig",
	}

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.ES256, Key: key}, &jose.SignerOptions{
		ExtraHeaders: map[jose.HeaderKey]interface{}{
			"typ": "dpop+jwt",
			"jwk": jwk,
		},
	})
	if err != nil {
		return "", err
	}

	claims := map[string]interface{}{
		"jti": fmt.Sprintf("%d", time.Now().UnixNano()),
		"htm": method,
		"htu": targetURL,
		"iat": time.Now().Unix(),
	}
	if nonce != "" {
		claims["nonce"] = nonce
	}

	if accessToken != "" {
		hash := sha256.Sum256([]byte(accessToken))
		claims["ath"] = base64.RawURLEncoding.EncodeToString(hash[:])
	}

	return jwt.Signed(signer).Claims(claims).CompactSerialize()
}

func makeClientAssertion(clientID, audience string) (string, error) {
	key, err := getPrivateKey()
	if err != nil {
		return "", err
	}

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.ES256, Key: key}, &jose.SignerOptions{
		ExtraHeaders: map[jose.HeaderKey]interface{}{
			"typ": "JWT",
			"kid": "key1",
		},
	})
	if err != nil {
		return "", err
	}

	claims := jwt.Claims{
		Issuer:   clientID,
		Subject:  clientID,
		Audience: jwt.Audience{audience},
		ID:       fmt.Sprintf("%d", time.Now().UnixNano()),
		IssuedAt: jwt.NewNumericDate(time.Now()),
		Expiry:   jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
	}

	return jwt.Signed(signer).Claims(claims).CompactSerialize()
}

func resolveHandle(handle string) (string, error) {
	fmt.Printf("Resolving handle: %s\n", handle)
	client := &http.Client{Timeout: 10 * time.Second}

	u := fmt.Sprintf("https://%s/.well-known/atproto-did", handle)
	fmt.Printf("GET %s\n", u)
	resp, err := client.Get(u)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		did := strings.TrimSpace(string(b))

		if strings.HasPrefix(did, "did:") {
			fmt.Printf("Resolved via well-known to: %s\n", did)
			return did, nil
		}
		fmt.Printf("Well-known returned invalid DID: %s\n", did[:min(50, len(did))])
	}
	if err != nil {
		fmt.Printf("Well-known error: %v\n", err)
	} else if resp != nil {
		fmt.Printf("Well-known status: %d\n", resp.StatusCode)
	}

	u = fmt.Sprintf("https://public.api.bsky.app/xrpc/com.atproto.identity.resolveHandle?handle=%s", handle)
	fmt.Printf("GET %s\n", u)
	resp, err = client.Get(u)
	if err != nil {
		fmt.Printf("XRPC error: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Printf("XRPC status: %d\n", resp.StatusCode)
		return "", fmt.Errorf("resolution failed")
	}

	var res struct {
		DID string `json:"did"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	fmt.Printf("Resolved via XRPC to: %s\n", res.DID)
	return res.DID, nil
}

func getPDSEndpoint(did string) (string, error) {
	fmt.Printf("Resolving PDS for DID: %s\n", did)

	resp, err := http.Get(fmt.Sprintf("https://plc.directory/%s", did))
	if err != nil {
		fmt.Printf("PLC HTTP error: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Printf("PLC Status: %d\n", resp.StatusCode)
		return "", fmt.Errorf("plc resolution failed")
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Printf("PLC Raw Response: %s\n", string(bodyBytes))

	var plc struct {
		Services []struct {
			Type     string `json:"type"`
			Endpoint string `json:"serviceEndpoint"`
		} `json:"service"`
	}
	if err := json.Unmarshal(bodyBytes, &plc); err != nil {
		fmt.Printf("PLC Decode error: %v\n", err)
		return "", err
	}
	fmt.Printf("PLC Services found: %d\n", len(plc.Services))

	for _, s := range plc.Services {
		fmt.Printf("Service Type: %s, Endpoint: %s\n", s.Type, s.Endpoint)
		if s.Type == "atproto_pds" || s.Type == "AtprotoPersonalDataServer" {
			return s.Endpoint, nil
		}
	}
	return "", fmt.Errorf("no pds found")
}

func getAuthMetadata(pds string) (*AuthMetadata, error) {
	pds = strings.TrimRight(pds, "/")

	authServer := pds
	if strings.Contains(pds, ".bsky.network") {
		authServer = "https://bsky.social"
	}

	resp, err := http.Get(fmt.Sprintf("%s/.well-known/oauth-authorization-server", authServer))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("auth metadata failed: status %d", resp.StatusCode)
	}

	var meta AuthMetadata
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func fetchAtprotoEmail(pds, accessToken, dpopNonce string) (string, error) {
	endpoint := strings.TrimRight(pds, "/") + "/xrpc/com.atproto.server.getSession"

	makeReq := func(nonce string) (*http.Response, error) {
		dpop, err := makeDPoPHeaderWithAth("GET", endpoint, nonce, accessToken)
		if err != nil {
			return nil, err
		}
		req, _ := http.NewRequest("GET", endpoint, nil)
		req.Header.Set("Authorization", "DPoP "+accessToken)
		req.Header.Set("DPoP", dpop)
		req.Header.Set("Accept", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		return client.Do(req)
	}

	resp, err := makeReq(dpopNonce)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == 400 {
		nonce := resp.Header.Get("DPoP-Nonce")
		if nonce != "" {
			resp.Body.Close()
			resp, err = makeReq(nonce)
			if err != nil {
				return "", err
			}
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", nil
	}

	var sess struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&sess); err != nil {
		return "", err
	}
	return sess.Email, nil
}
