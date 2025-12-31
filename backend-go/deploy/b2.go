package deploy

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type B2Client struct {
	KeyID       string
	AppKey      string
	BucketID    string
	AuthToken   string
	APIURL      string
	DownloadURL string
	Client      *http.Client
	mu          sync.Mutex
}

func NewB2Client(keyID, appKey, bucketID string) *B2Client {
	return &B2Client{
		KeyID:    keyID,
		AppKey:   appKey,
		BucketID: bucketID,
		Client:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *B2Client) Authorize() error {
	req, err := http.NewRequest("GET", "https://api.backblazeb2.com/b2api/v2/b2_authorize_account", nil)
	if err != nil {
		return err
	}
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.KeyID, c.AppKey)))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("authorize failed: %d", resp.StatusCode)
	}

	var res struct {
		AuthorizationToken      string `json:"authorizationToken"`
		APIURL                  string `json:"apiUrl"`
		DownloadURL             string `json:"downloadUrl"`
		AbsoluteMinimumPartSize int    `json:"absoluteMinimumPartSize"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}

	c.mu.Lock()
	c.AuthToken = res.AuthorizationToken
	c.APIURL = res.APIURL
	c.DownloadURL = res.DownloadURL
	c.mu.Unlock()

	return nil
}

func (c *B2Client) GetUploadURL() (string, string, error) {
	c.mu.Lock()
	apiURL := c.APIURL
	authToken := c.AuthToken
	c.mu.Unlock()

	if apiURL == "" {
		return "", "", fmt.Errorf("not authorized")
	}

	body, _ := json.Marshal(map[string]string{
		"bucketId": c.BucketID,
	})

	req, _ := http.NewRequest("POST", apiURL+"/b2api/v2/b2_get_upload_url", bytes.NewBuffer(body))
	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("get_upload_url failed: %d", resp.StatusCode)
	}

	var res struct {
		UploadURL          string `json:"uploadUrl"`
		AuthorizationToken string `json:"authorizationToken"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", "", err
	}

	return res.UploadURL, res.AuthorizationToken, nil
}

func (c *B2Client) UploadFile(fileName string, content []byte, contentType string) error {

	c.mu.Lock()
	apiURL := c.APIURL
	c.mu.Unlock()

	if apiURL == "" {
		if err := c.Authorize(); err != nil {
			return err
		}
	}

	uploadURL, uploadToken, err := c.GetUploadURL()
	if err != nil {

		if err := c.Authorize(); err != nil {
			return err
		}
		uploadURL, uploadToken, err = c.GetUploadURL()
		if err != nil {
			return err
		}
	}

	hash := sha1.Sum(content)
	sha1Str := hex.EncodeToString(hash[:])

	encodedName := url.PathEscape(fileName)

	req, _ := http.NewRequest("POST", uploadURL, bytes.NewBuffer(content))
	req.Header.Set("Authorization", uploadToken)
	req.Header.Set("X-Bz-File-Name", encodedName)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Bz-Content-Sha1", sha1Str)

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload_file failed: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *B2Client) ListFileNames(prefix string, startFileName string, maxFileCount int) ([]string, string, error) {
	c.mu.Lock()
	apiURL := c.APIURL
	authToken := c.AuthToken
	c.mu.Unlock()

	if apiURL == "" {
		if err := c.Authorize(); err != nil {
			return nil, "", err
		}
		c.mu.Lock()
		apiURL = c.APIURL
		authToken = c.AuthToken
		c.mu.Unlock()
	}

	bodyMap := map[string]interface{}{
		"bucketId":     c.BucketID,
		"maxFileCount": maxFileCount,
	}
	if prefix != "" {
		bodyMap["prefix"] = prefix
	}
	if startFileName != "" {
		bodyMap["startFileName"] = startFileName
	}

	body, _ := json.Marshal(bodyMap)

	req, _ := http.NewRequest("POST", apiURL+"/b2api/v2/b2_list_file_names", bytes.NewBuffer(body))
	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("list_file_names failed: %d", resp.StatusCode)
	}

	var res struct {
		Files []struct {
			FileName string `json:"fileName"`
			FileID   string `json:"fileId"`
		} `json:"files"`
		NextFileName *string `json:"nextFileName"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, "", err
	}

	var names []string
	for _, f := range res.Files {
		names = append(names, f.FileName)
	}

	next := ""
	if res.NextFileName != nil {
		next = *res.NextFileName
	}

	return names, next, nil
}

func (c *B2Client) DeleteFileVersion(fileName, fileID string) error {
	c.mu.Lock()
	apiURL := c.APIURL
	authToken := c.AuthToken
	c.mu.Unlock()

	if apiURL == "" {
		if err := c.Authorize(); err != nil {
			return err
		}
		c.mu.Lock()
		apiURL = c.APIURL
		authToken = c.AuthToken
		c.mu.Unlock()
	}

	body, _ := json.Marshal(map[string]string{
		"fileName": fileName,
		"fileId":   fileID,
	})

	req, _ := http.NewRequest("POST", apiURL+"/b2api/v2/b2_delete_file_version", bytes.NewBuffer(body))
	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("delete_file_version failed: %d", resp.StatusCode)
	}

	return nil
}

func (c *B2Client) DeleteFilesWithPrefix(prefix string) error {
	for {

		c.mu.Lock()
		apiURL := c.APIURL
		authToken := c.AuthToken
		c.mu.Unlock()

		if apiURL == "" {
			if err := c.Authorize(); err != nil {
				return err
			}
			c.mu.Lock()
			apiURL = c.APIURL
			authToken = c.AuthToken
			c.mu.Unlock()
		}

		bodyMap := map[string]interface{}{
			"bucketId":     c.BucketID,
			"maxFileCount": 100,
			"prefix":       prefix,
		}
		body, _ := json.Marshal(bodyMap)

		req, _ := http.NewRequest("POST", apiURL+"/b2api/v2/b2_list_file_names", bytes.NewBuffer(body))
		req.Header.Set("Authorization", authToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.Client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("list_file_names failed: %d", resp.StatusCode)
		}

		var res struct {
			Files []struct {
				FileName string `json:"fileName"`
				FileID   string `json:"fileId"`
			} `json:"files"`
			NextFileName *string `json:"nextFileName"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			return err
		}

		if len(res.Files) == 0 {
			break
		}

		for _, f := range res.Files {
			if err := c.DeleteFileVersion(f.FileName, f.FileID); err != nil {

				fmt.Printf("Failed to delete %s: %v\n", f.FileName, err)
			}
		}

		if res.NextFileName == nil {
			break
		}
	}
	return nil
}
