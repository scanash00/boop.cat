// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package lib

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type TurnstileResult struct {
	OK    bool
	Error string
}

func VerifyTurnstile(token, remoteIP string) TurnstileResult {
	secret := os.Getenv("TURNSTILE_SECRET_KEY")
	if secret == "" {

		return TurnstileResult{OK: true}
	}

	if token == "" {
		return TurnstileResult{OK: false, Error: "missing-token"}
	}

	data := url.Values{}
	data.Set("secret", secret)
	data.Set("response", token)
	if remoteIP != "" {
		data.Set("remoteip", remoteIP)
	}

	resp, err := http.Post(
		"https://challenges.cloudflare.com/turnstile/v0/siteverify",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return TurnstileResult{OK: false, Error: "verification-failed"}
	}
	defer resp.Body.Close()

	var result struct {
		Success    bool     `json:"success"`
		ErrorCodes []string `json:"error-codes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return TurnstileResult{OK: false, Error: "parse-failed"}
	}

	if !result.Success {
		errCode := "captcha-failed"
		if len(result.ErrorCodes) > 0 {
			errCode = result.ErrorCodes[0]
		}
		return TurnstileResult{OK: false, Error: errCode}
	}

	return TurnstileResult{OK: true}
}
