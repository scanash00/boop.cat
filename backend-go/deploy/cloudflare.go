// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package deploy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type CloudflareClient struct {
	AccountID   string
	NamespaceID string
	APIToken    string
	Client      *http.Client
}

func NewCloudflareClient(accountID, namespaceID, token string) *CloudflareClient {
	return &CloudflareClient{
		AccountID:   accountID,
		NamespaceID: namespaceID,
		APIToken:    token,
		Client:      &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *CloudflareClient) Do(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	url := "https://api.cloudflare.com/client/v4" + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	req.Header.Set("Content-Type", "application/json")

	return c.Client.Do(req)
}

func (c *CloudflareClient) KVPut(key, value string) error {
	encodedKey := url.PathEscape(key)
	url := fmt.Sprintf("/accounts/%s/storage/kv/namespaces/%s/values/%s",
		c.AccountID, c.NamespaceID, encodedKey)

	fullURL := "https://api.cloudflare.com/client/v4" + url
	req, _ := http.NewRequest("PUT", fullURL, bytes.NewBuffer([]byte(value)))
	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("kv_put failed: %s", resp.Status)
	}
	return nil
}

func (c *CloudflareClient) KVGet(key string) (string, error) {
	encodedKey := url.PathEscape(key)
	url := fmt.Sprintf("/accounts/%s/storage/kv/namespaces/%s/values/%s",
		c.AccountID, c.NamespaceID, encodedKey)

	fullURL := "https://api.cloudflare.com/client/v4" + url
	req, _ := http.NewRequest("GET", fullURL, nil)
	req.Header.Set("Authorization", "Bearer "+c.APIToken)

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", nil
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("kv_get failed: %s", resp.Status)
	}

	b, _ := io.ReadAll(resp.Body)
	return string(b), nil
}

func (c *CloudflareClient) KVDelete(key string) error {
	encodedKey := url.PathEscape(key)
	url := fmt.Sprintf("/accounts/%s/storage/kv/namespaces/%s/values/%s",
		c.AccountID, c.NamespaceID, encodedKey)

	fullURL := "https://api.cloudflare.com/client/v4" + url
	req, _ := http.NewRequest("DELETE", fullURL, nil)
	req.Header.Set("Authorization", "Bearer "+c.APIToken)

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

type Response struct {
	Result  json.RawMessage `json:"result"`
	Success bool            `json:"success"`
	Errors  []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (c *CloudflareClient) GetZoneID(domain string) (string, error) {
	resp, err := c.Do("GET", "/zones?name="+domain, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res struct {
		Result []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&res)

	if len(res.Result) == 0 {
		return "", fmt.Errorf("zone not found")
	}
	return res.Result[0].ID, nil
}

func (c *CloudflareClient) CreateCustomHostname(zoneID, hostname string) (json.RawMessage, error) {
	body := map[string]interface{}{
		"hostname": hostname,
		"ssl": map[string]interface{}{
			"method": "http",
			"type":   "dv",
			"settings": map[string]interface{}{
				"min_tls_version": "1.2",
				"http2":           "on",
			},
		},
	}

	resp, err := c.Do("POST", fmt.Sprintf("/zones/%s/custom_hostnames", zoneID), body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res Response
	json.NewDecoder(resp.Body).Decode(&res)

	if !res.Success {
		msg := "unknown"
		if len(res.Errors) > 0 {
			msg = res.Errors[0].Message
		}
		return nil, fmt.Errorf("create_custom_hostname failed: %s", msg)
	}
	return res.Result, nil
}

func (c *CloudflareClient) GetCustomHostname(zoneID, id string) (json.RawMessage, error) {
	resp, err := c.Do("GET", fmt.Sprintf("/zones/%s/custom_hostnames/%s", zoneID, id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res Response
	json.NewDecoder(resp.Body).Decode(&res)
	return res.Result, nil
}

func (c *CloudflareClient) GetCustomHostnameIDByName(zoneID, hostname string) (string, error) {
	resp, err := c.Do("GET", fmt.Sprintf("/zones/%s/custom_hostnames?hostname=%s", zoneID, url.QueryEscape(hostname)), nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res struct {
		Result []struct {
			ID string `json:"id"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	if len(res.Result) == 0 {
		return "", fmt.Errorf("hostname not found")
	}
	return res.Result[0].ID, nil
}

func (c *CloudflareClient) DeleteCustomHostname(zoneID, id string) error {
	resp, err := c.Do("DELETE", fmt.Sprintf("/zones/%s/custom_hostnames/%s", zoneID, id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *CloudflareClient) UpdateFallbackOrigin(zoneID, origin string) error {
	body := map[string]string{"origin": origin}
	resp, err := c.Do("PUT", fmt.Sprintf("/zones/%s/custom_hostnames/fallback_origin", zoneID), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

type DNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
}

func (c *CloudflareClient) GetDNSRecords(zoneID, name string) ([]DNSRecord, error) {
	path := fmt.Sprintf("/zones/%s/dns_records", zoneID)
	if name != "" {
		path += "?name=" + url.QueryEscape(name)
	}
	resp, err := c.Do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res struct {
		Result []DNSRecord `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	return res.Result, nil
}

func (c *CloudflareClient) CreateDNSRecord(zoneID string, record DNSRecord) error {
	path := fmt.Sprintf("/zones/%s/dns_records", zoneID)
	resp, err := c.Do("POST", path, record)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("create_dns failed: %s", resp.Status)
	}
	return nil
}

func (c *CloudflareClient) UpdateDNSRecord(zoneID, recordID string, record DNSRecord) error {
	path := fmt.Sprintf("/zones/%s/dns_records/%s", zoneID, recordID)
	resp, err := c.Do("PATCH", path, record)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("update_dns failed: %s", resp.Status)
	}
	return nil
}

func (c *CloudflareClient) EnsureRouting(subdomain, siteID, deployID string) error {
	if subdomain != "" {
		if err := c.KVPut("host:"+subdomain, siteID); err != nil {
			return err
		}
	}
	if deployID != "" {
		if err := c.KVPut("current:"+siteID, deployID); err != nil {
			return err
		}
	}
	return nil
}

func (c *CloudflareClient) RemoveRouting(subdomain, siteID, domain string) error {
	if domain != "" {

		if err := c.KVDelete("host:" + domain); err != nil {
			return err
		}
	}
	if subdomain != "" {
		if err := c.KVDelete("host:" + subdomain); err != nil {
			return err
		}
	}
	if siteID != "" {
		if err := c.KVDelete("current:" + siteID); err != nil {
			return err
		}
	}
	return nil
}
