package deploy

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nrednav/cuid2"

	"boop-cat/db"
	"boop-cat/lib"
)

type Engine struct {
	DB             *sql.DB
	WorkDir        string
	B2KeyID        string
	B2AppKey       string
	B2BucketID     string
	CFToken        string
	CFAccountID    string
	CFNamespaceID  string
	deploymentsMux sync.Mutex
	deployments    map[string]context.CancelFunc
}

func NewEngine(database *sql.DB, b2KeyID, b2AppKey, b2BucketID, cfToken, cfAccount, cfNamespace string) *Engine {

	workDir := filepath.Join(os.TempDir(), "fsd-builds")
	os.MkdirAll(workDir, 0755)

	return &Engine{
		DB:            database,
		WorkDir:       workDir,
		B2KeyID:       b2KeyID,
		B2AppKey:      b2AppKey,
		B2BucketID:    b2BucketID,
		CFToken:       cfToken,
		CFAccountID:   cfAccount,
		CFNamespaceID: cfNamespace,
		deployments:   make(map[string]context.CancelFunc),
	}
}

func (e *Engine) DeploySite(siteID, userID string, logStream chan<- string) (*db.Deployment, error) {

	var commitSha, commitMessage, commitAuthor, commitAvatar *string

	site, err := db.GetSiteByID(e.DB, userID, siteID)
	if err == nil && site.GitURL.Valid && strings.Contains(site.GitURL.String, "github.com") {

		ghToken, _ := db.GetGitHubToken(e.DB, userID)

		parts := strings.Split(strings.TrimPrefix(site.GitURL.String, "https://github.com/"), "/")
		if len(parts) >= 2 {
			owner := parts[0]
			repo := strings.TrimSuffix(parts[1], ".git")
			branch := "main"
			if site.GitBranch.Valid {
				branch = site.GitBranch.String
			}

			apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", owner, repo, branch)
			req, _ := http.NewRequest("GET", apiURL, nil)
			if ghToken != "" {
				req.Header.Set("Authorization", "Bearer "+ghToken)
			}

			client := &http.Client{Timeout: 2 * time.Second}
			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == 200 {
					var res struct {
						SHA    string `json:"sha"`
						Commit struct {
							Message string `json:"message"`
							Author  struct {
								Name string `json:"name"`
							} `json:"author"`
						} `json:"commit"`
						Author struct {
							AvatarURL string `json:"avatar_url"`
						} `json:"author"`
						Committer struct {
							AvatarURL string `json:"avatar_url"`
						} `json:"committer"`
					}

					if err := json.NewDecoder(resp.Body).Decode(&res); err == nil {

						sha := res.SHA
						commitSha = &sha

						msg := res.Commit.Message
						commitMessage = &msg

						auth := res.Commit.Author.Name
						commitAuthor = &auth

						if res.Author.AvatarURL != "" {
							av := res.Author.AvatarURL
							commitAvatar = &av
						} else if res.Committer.AvatarURL != "" {
							av := res.Committer.AvatarURL
							commitAvatar = &av
						}
					}
				}
			}
		}
	}

	deployID := cuid2.Generate()
	err = db.CreateDeployment(e.DB, deployID, userID, siteID, "building", commitSha, commitMessage, commitAuthor, commitAvatar)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment record: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	e.deploymentsMux.Lock()
	e.deployments[deployID] = cancel
	e.deploymentsMux.Unlock()

	go func() {
		defer func() {
			e.deploymentsMux.Lock()
			delete(e.deployments, deployID)
			e.deploymentsMux.Unlock()
			cancel()
			if logStream != nil {
				close(logStream)
			}
		}()

		logsDir := filepath.Join(e.WorkDir, "logs")
		os.MkdirAll(logsDir, 0755)
		logsPath := filepath.Join(logsDir, deployID+".log")
		logFile, _ := os.Create(logsPath)

		db.UpdateDeploymentLogs(e.DB, deployID, logsPath)

		logger := func(msg string) {
			log.Printf("[Deploy %s] %s", deployID, msg)
			if logFile != nil {
				logFile.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), msg))
			}
			if logStream != nil {
				logStream <- msg
			}
		}

		err := e.runPipeline(ctx, siteID, userID, deployID, logger)
		if logFile != nil {
			logFile.Close()
		}
		if err != nil {
			logger(fmt.Sprintf("Deployment failed: %v", err))
			if ctx.Err() == context.Canceled {
				db.UpdateDeploymentStatus(e.DB, deployID, "canceled", "")
			} else {
				db.UpdateDeploymentStatus(e.DB, deployID, "failed", "")
			}
		} else {
			logger("Deployment successful")
		}
	}()

	return db.GetDeploymentByID(e.DB, deployID)
}

func (e *Engine) CancelDeployment(deployID string) error {
	e.deploymentsMux.Lock()
	cancel, ok := e.deployments[deployID]
	e.deploymentsMux.Unlock()

	if !ok {
		return fmt.Errorf("deployment not found or not running")
	}

	cancel()
	return nil
}

func (e *Engine) runPipeline(ctx context.Context, siteID, userID, deployID string, logger func(string)) error {

	site, err := db.GetSiteByID(e.DB, userID, siteID)
	if err != nil {
		return err
	}

	logger(fmt.Sprintf("Starting deployment for site %s (%s)", site.Name, site.ID))

	if ctx.Err() != nil {
		return ctx.Err()
	}

	buildDir := filepath.Join(e.WorkDir, deployID)

	logger("Cloning repository...")
	if !site.GitURL.Valid {
		return fmt.Errorf("site has no git url")
	}
	repoURL := site.GitURL.String

	ghToken, err := db.GetGitHubToken(e.DB, userID)
	if err == nil && ghToken != "" && strings.Contains(repoURL, "github.com") {

		if strings.HasPrefix(repoURL, "https://github.com/") {
			logger("Injecting GitHub authentication token...")

			repoURL = strings.Replace(repoURL, "https://github.com/", fmt.Sprintf("https://oauth2:%s@github.com/", ghToken), 1)
		}
	} else if err != nil {
		logger(fmt.Sprintf("Warning: Failed to check for GitHub token: %v", err))
	} else if ghToken == "" {
		logger("No GitHub token found for user. Private repos may fail.")
	}

	branch := "main"
	if site.GitBranch.Valid {
		branch = site.GitBranch.String
	}

	err = GitClone(ctx, repoURL, branch, buildDir, 1, logger)
	if err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}
	head, err := GitCurrentHead(buildDir)
	if err == nil {

		avatarURL := ""
		if strings.Contains(site.GitURL.String, "github.com") {

			parts := strings.Split(strings.TrimPrefix(site.GitURL.String, "https://github.com/"), "/")
			if len(parts) >= 2 {
				owner := parts[0]
				repo := strings.TrimSuffix(parts[1], ".git")

				apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", owner, repo, head.SHA)
				req, _ := http.NewRequest("GET", apiURL, nil)
				if ghToken != "" {
					req.Header.Set("Authorization", "Bearer "+ghToken)
				}

				client := &http.Client{Timeout: 5 * time.Second}
				resp, err := client.Do(req)
				if err == nil {
					defer resp.Body.Close()
					var res struct {
						Author struct {
							AvatarURL string `json:"avatar_url"`
						} `json:"author"`
						Committer struct {
							AvatarURL string `json:"avatar_url"`
						} `json:"committer"`
					}
					if err := json.NewDecoder(resp.Body).Decode(&res); err == nil {
						if res.Author.AvatarURL != "" {
							avatarURL = res.Author.AvatarURL
						} else if res.Committer.AvatarURL != "" {
							avatarURL = res.Committer.AvatarURL
						}
					}
				}
			}
		}

		e.DB.Exec(`UPDATE deployments SET commitSha=?, commitMessage=?, commitAuthor=?, commitAvatar=? WHERE id=?`,
			head.SHA, head.Message, head.Author, avatarURL, deployID)
	}

	logger("Building project...")

	envVars := []string{}
	if site.EnvText.Valid && site.EnvText.String != "" {
		decryptedEnv := lib.Decrypt(site.EnvText.String)
		envVars = parseEnvText(decryptedEnv)
	}

	bs := &BuildSystem{
		RootDir: buildDir,
		Env:     envVars,
		Logger:  logger,
	}

	customBuildCmd := ""
	if site.BuildCommand.Valid {
		customBuildCmd = site.BuildCommand.String
	}

	outputDirName, err := bs.Build(ctx, customBuildCmd)
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fullOutputDir := filepath.Join(buildDir, outputDirName)
	if !fileExists(fullOutputDir) {
		return fmt.Errorf("output directory %s not found", outputDirName)
	}

	logger("Build complete. Starting upload...")
	db.UpdateDeploymentStatus(e.DB, deployID, "running", "")

	logger("Uploading to storage...")
	b2 := NewB2Client(e.B2KeyID, e.B2AppKey, e.B2BucketID)

	prefix := fmt.Sprintf("sites/%s/%s", siteID, deployID)

	files, _ := ListFilesRecursive(fullOutputDir)
	logger(fmt.Sprintf("Found %d files to upload", len(files)))

	const maxConcurrency = 20
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var uploadErr error
	var errMutex sync.Mutex

	for _, fPath := range files {
		if ctx.Err() != nil {
			break
		}

		wg.Add(1)
		semaphore <- struct{}{}

		go func(path string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			errMutex.Lock()
			if uploadErr != nil {
				errMutex.Unlock()
				return
			}
			errMutex.Unlock()

			relPath, _ := filepath.Rel(fullOutputDir, path)
			key := fmt.Sprintf("%s/%s", prefix, relPath)
			key = filepath.ToSlash(key)

			content, err := ioutil.ReadFile(path)
			if err != nil {
				errMutex.Lock()
				if uploadErr == nil {
					uploadErr = fmt.Errorf("read failed for %s: %w", relPath, err)
				}
				errMutex.Unlock()
				return
			}

			contentType := "application/octet-stream"
			if strings.HasSuffix(key, ".html") {
				contentType = "text/html"
			}
			if strings.HasSuffix(key, ".css") {
				contentType = "text/css"
			}
			if strings.HasSuffix(key, ".js") {
				contentType = "application/javascript"
			}
			if strings.HasSuffix(key, ".json") {
				contentType = "application/json"
			}
			if strings.HasSuffix(key, ".png") {
				contentType = "image/png"
			}
			if strings.HasSuffix(key, ".jpg") {
				contentType = "image/jpeg"
			}
			if strings.HasSuffix(key, ".svg") {
				contentType = "image/svg+xml"
			}

			err = b2.UploadFile(key, content, contentType)
			if err != nil {
				errMutex.Lock()
				if uploadErr == nil {
					uploadErr = fmt.Errorf("upload failed for %s: %w", relPath, err)
				}
				errMutex.Unlock()
			}
		}(fPath)
	}

	wg.Wait()

	if uploadErr != nil {
		return uploadErr
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	logger("Upload complete")

	logger("Updating routing...")
	cf := NewCloudflareClient(e.CFAccountID, e.CFNamespaceID, e.CFToken)
	rootDomain := os.Getenv("FSD_EDGE_ROOT_DOMAIN")

	if site.Domain != "" {

		routingKey := site.Domain

		if rootDomain != "" && strings.HasSuffix(site.Domain, "."+rootDomain) {
			routingKey = strings.TrimSuffix(site.Domain, "."+rootDomain)
		} else if rootDomain != "" && site.Domain == rootDomain {

			routingKey = "@"
		}

		err = cf.EnsureRouting(routingKey, siteID, deployID)
		if err != nil {
			return fmt.Errorf("routing update failed: %w", err)
		}
	}

	customDomains, _ := db.ListCustomDomains(e.DB, siteID)
	for _, cd := range customDomains {

		hostname := strings.ToLower(strings.TrimSpace(cd.Hostname))
		hostname = strings.TrimPrefix(hostname, "http://")
		hostname = strings.TrimPrefix(hostname, "https://")

		logger(fmt.Sprintf("Updating routing for custom domain: %s", hostname))
		err = cf.EnsureRouting(hostname, siteID, deployID)
		if err != nil {
			logger(fmt.Sprintf("Failed to update routing for %s: %v", hostname, err))
		}
	}

	if site.Domain == "" && len(customDomains) == 0 {

		err = cf.EnsureRouting("", siteID, deployID)
		if err != nil {
			return fmt.Errorf("routing update failed: %w", err)
		}
	}

	if rootDomain == "" {
		rootDomain = os.Getenv("FSD_EDGE_ROOT_DOMAIN")
	}
	if rootDomain == "" {
		rootDomain = "boop.cat"
	}

	finalURL := ""

	if len(customDomains) > 0 {
		for _, cd := range customDomains {

			if cd.Status == "active" || cd.Status == "live" {
				finalURL = fmt.Sprintf("https://%s", cd.Hostname)
				break
			}
		}

		if finalURL == "" && len(customDomains) > 0 {
			finalURL = fmt.Sprintf("https://%s", customDomains[0].Hostname)
		}
	}

	if finalURL == "" && site.Domain != "" {
		if strings.HasSuffix(site.Domain, "."+rootDomain) || site.Domain == rootDomain {
			finalURL = fmt.Sprintf("https://%s", site.Domain)
		} else {
			finalURL = fmt.Sprintf("https://%s.%s", site.Domain, rootDomain)
		}
	}

	db.UpdateDeploymentStatus(e.DB, deployID, "running", finalURL)
	db.UpdateSiteCurrentDeployment(e.DB, siteID, deployID)

	if err := db.StopOtherDeployments(e.DB, siteID, deployID); err != nil {
		logger(fmt.Sprintf("Warning: Failed to stop other deployments: %v", err))
	}

	logger("Deployment successful!")
	return nil
}

func ListFilesRecursive(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
		}
		if !info.IsDir() {

			if !strings.Contains(path, "/.git/") && !strings.Contains(path, "\\.git\\") {
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}

func parseEnvText(envText string) []string {
	var result []string
	lines := strings.Split(envText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])

			if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
				value = value[1 : len(value)-1]
			}
			result = append(result, fmt.Sprintf("%s=%s", key, value))
		}
	}
	return result
}

func (e *Engine) CleanupSite(siteID string, userID string) error {

	cf := NewCloudflareClient(e.CFAccountID, e.CFNamespaceID, e.CFToken)

	site, err := db.GetSiteByID(e.DB, userID, siteID)
	if err == nil && site != nil {

		rootDomain := os.Getenv("FSD_EDGE_ROOT_DOMAIN")
		routingKey := site.Domain
		if rootDomain != "" && strings.HasSuffix(site.Domain, "."+rootDomain) {
			routingKey = strings.TrimSuffix(site.Domain, "."+rootDomain)
		}
		cf.RemoveRouting(routingKey, site.ID, site.Domain)
	}

	customDomains, _ := db.ListCustomDomains(e.DB, siteID)
	for _, cd := range customDomains {
		cf.RemoveRouting("", siteID, cd.Hostname)

	}

	b2 := NewB2Client(e.B2KeyID, e.B2AppKey, e.B2BucketID)

	prefix := fmt.Sprintf("sites/%s/", siteID)

	err = b2.DeleteFilesWithPrefix(prefix)
	if err != nil {
		return fmt.Errorf("failed to delete files: %w", err)
	}

	return nil
}
