// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nrednav/cuid2"
)

type PreviewResult struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	DefaultBranch string   `json:"defaultBranch"`
	RootFiles     []string `json:"rootFiles"`
	Subdirs       []string `json:"subdirs"`
	IsPrivate     bool     `json:"isPrivate"`
}

func (e *Engine) PreviewGitRepo(gitURL string) (*PreviewResult, error) {

	cmd := exec.Command("git", "ls-remote", gitURL, "HEAD")

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("GIT_CLONE_FAILED: Repo not found or private")
	}

	tmpDir := filepath.Join(os.TempDir(), "fsd-preview-"+cuid2.Generate())
	defer os.RemoveAll(tmpDir)

	cloneCmd := exec.Command("git", "clone", "--depth", "1", gitURL, tmpDir)
	if out, err := cloneCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("GIT_CLONE_FAILED: %v - %s", err, string(out))
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return nil, err
	}

	var rootFiles []string
	var subdirs []string

	for _, ent := range entries {
		name := ent.Name()
		if name == ".git" {
			continue
		}
		if ent.IsDir() {
			subdirs = append(subdirs, name)
		} else {
			rootFiles = append(rootFiles, name)
		}
	}

	parts := strings.Split(strings.TrimSuffix(gitURL, ".git"), "/")
	name := ""
	if len(parts) > 0 {
		name = parts[len(parts)-1]
	}

	return &PreviewResult{
		Name:          name,
		DefaultBranch: "main",
		RootFiles:     rootFiles,
		Subdirs:       subdirs,
		IsPrivate:     false,
	}, nil
}
