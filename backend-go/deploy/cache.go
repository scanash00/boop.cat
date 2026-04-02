// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package deploy

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type BuildCache struct {
	CacheDir string
}

var lockfileNames = []string{
	"bun.lockb",
	"bun.lock",
	"pnpm-lock.yaml",
	"yarn.lock",
	"package-lock.json",
}

func (c *BuildCache) LockfileHash(buildDir string) string {
	for _, name := range lockfileNames {
		path := filepath.Join(buildDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		h := sha256.Sum256(data)
		return fmt.Sprintf("%x", h)
	}
	return ""
}

func (c *BuildCache) siteCacheDir(siteID string) string {
	return filepath.Join(c.CacheDir, siteID)
}

func (c *BuildCache) hashMarkerPath(siteID string) string {
	return filepath.Join(c.siteCacheDir(siteID), ".lockfile-hash")
}

func (c *BuildCache) RestoreNodeModules(siteID, buildDir, currentHash string, logger func(string)) bool {
	if currentHash == "" {
		return false
	}

	markerPath := c.hashMarkerPath(siteID)
	stored, err := os.ReadFile(markerPath)
	if err != nil {
		return false
	}

	if string(stored) != currentHash {
		if logger != nil {
			logger("Cache miss: lockfile has changed")
		}
		return false
	}

	cachedModules := filepath.Join(c.siteCacheDir(siteID), "node_modules")
	if !fileExists(cachedModules) {
		return false
	}

	dest := filepath.Join(buildDir, "node_modules")
	cmd := exec.Command("cp", "-a", cachedModules, dest)
	if err := cmd.Run(); err != nil {
		if logger != nil {
			logger(fmt.Sprintf("Cache restore failed: %v", err))
		}
		return false
	}

	return true
}

func (c *BuildCache) SaveNodeModules(siteID, buildDir, lockfileHash string, logger func(string)) {
	if lockfileHash == "" {
		return
	}

	srcModules := filepath.Join(buildDir, "node_modules")
	if !fileExists(srcModules) {
		return
	}

	siteCache := c.siteCacheDir(siteID)
	os.MkdirAll(siteCache, 0755)

	cachedModules := filepath.Join(siteCache, "node_modules")
	os.RemoveAll(cachedModules)

	cmd := exec.Command("cp", "-a", srcModules, cachedModules)
	if err := cmd.Run(); err != nil {
		if logger != nil {
			logger(fmt.Sprintf("Cache save failed: %v", err))
		}
		return
	}

	os.WriteFile(c.hashMarkerPath(siteID), []byte(lockfileHash), 0644)
}
