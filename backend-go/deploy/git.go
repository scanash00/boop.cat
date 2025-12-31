// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package deploy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func GitClone(ctx context.Context, repoURL, branch, targetDir string, depth int, logger func(string)) error {
	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("failed to clear target dir: %w", err)
	}
	if err := ensureDir(filepath.Dir(targetDir)); err != nil {
		return fmt.Errorf("failed to create parent dir: %w", err)
	}

	args := []string{"clone", "--no-tags", "--depth", fmt.Sprintf("%d", depth)}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, repoURL, targetDir)

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return err
	}

	sanitize := func(s string) string {

		if strings.Contains(s, "@github.com") {
			parts := strings.Split(s, "@github.com")
			if len(parts) > 1 {

				lastSpace := strings.LastIndex(parts[0], " ")
				if lastSpace != -1 {
					return parts[0][:lastSpace+1] + "***" + "@github.com" + parts[1]
				}

				lastSlash := strings.LastIndex(parts[0], "//")
				if lastSlash != -1 {
					return parts[0][:lastSlash+2] + "***" + "@github.com" + parts[1]
				}
			}
		}
		return s
	}

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 && logger != nil {
				logger(sanitize(string(buf[:n])))
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if n > 0 && logger != nil {
				logger(sanitize(string(buf[:n])))
			}
			if err != nil {
				break
			}
		}
	}()

	return cmd.Wait()
}

func GitCheckout(targetDir, ref string) error {
	cmd := exec.Command("git", "checkout", ref)
	cmd.Dir = targetDir
	return cmd.Run()
}

type CommitInfo struct {
	SHA     string
	Message string
	Author  string
}

func GitCurrentHead(targetDir string) (*CommitInfo, error) {

	shaCmd := exec.Command("git", "rev-parse", "HEAD")
	shaCmd.Dir = targetDir
	shaOut, err := shaCmd.Output()
	if err != nil {
		return nil, err
	}

	msgCmd := exec.Command("git", "log", "-1", "--pretty=%s")
	msgCmd.Dir = targetDir
	msgOut, _ := msgCmd.Output()

	authCmd := exec.Command("git", "log", "-1", "--pretty=%an <%ae>")
	authCmd.Dir = targetDir
	authOut, _ := authCmd.Output()

	return &CommitInfo{
		SHA:     strings.TrimSpace(string(shaOut)),
		Message: strings.TrimSpace(string(msgOut)),
		Author:  strings.TrimSpace(string(authOut)),
	}, nil
}
