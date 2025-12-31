// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package deploy

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func validateBuildCommand(cmd string) error {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return nil
	}

	dangerousMap := []string{
		"&", "|", ";", ">", "<", "`", "$(",
	}
	for _, char := range dangerousMap {
		if strings.Contains(cmd, char) {
			return fmt.Errorf("command contains forbidden character: %s", char)
		}
	}

	allowedPrefixes := []string{
		"npm ", "yarn ", "pnpm ", "bun ", "npx ", "node ",
	}
	isAllowed := false
	for _, p := range allowedPrefixes {
		if strings.HasPrefix(cmd, p) {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		return errors.New("command must start with npm, yarn, pnpm, bun, npx, or node")
	}

	forbiddenKeywords := []string{
		" start", " dev", " serve", " preview", " watch",
	}
	for _, kw := range forbiddenKeywords {
		if strings.Contains(cmd, kw) {
			return fmt.Errorf("command looks like a runtime server (contains '%s'), only build commands are allowed", strings.TrimSpace(kw))
		}
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

type BuildSystem struct {
	RootDir string
	Env     []string
	Logger  func(string)
}

func (b *BuildSystem) DetectPackageManager() string {
	if fileExists(filepath.Join(b.RootDir, "bun.lockb")) || fileExists(filepath.Join(b.RootDir, "bun.lock")) {
		return "bun"
	}
	if fileExists(filepath.Join(b.RootDir, "pnpm-lock.yaml")) {
		return "pnpm"
	}
	if fileExists(filepath.Join(b.RootDir, "yarn.lock")) {
		return "yarn"
	}
	if fileExists(filepath.Join(b.RootDir, "package-lock.json")) {
		return "npm"
	}

	return "npm"
}

func (b *BuildSystem) InstallArgs(pm string) []string {

	switch pm {
	case "npm":
		if fileExists(filepath.Join(b.RootDir, "package-lock.json")) {
			return []string{"ci", "--include=dev"}
		}
		return []string{"install", "--include=dev"}
	case "yarn":
		if fileExists(filepath.Join(b.RootDir, "yarn.lock")) {
			return []string{"install", "--frozen-lockfile", "--production=false"}
		}
		return []string{"install", "--production=false"}
	case "pnpm":
		return []string{"install", "--frozen-lockfile", "--production=false"}
	case "bun":
		return []string{"install"}
	}
	return []string{"install"}
}

func (b *BuildSystem) BuildArgs(pm string) []string {
	switch pm {
	case "yarn", "pnpm":
		return []string{"build"}
	case "deno":
		return []string{"task", "build"}
	default:
		return []string{"run", "build"}
	}
}

func (b *BuildSystem) RunCommand(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = b.RootDir
	cmd.Env = append(os.Environ(), b.Env...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 && b.Logger != nil {
				b.Logger(string(buf[:n]))
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
			if n > 0 && b.Logger != nil {
				b.Logger(string(buf[:n]))
			}
			if err != nil {
				break
			}
		}
	}()

	return cmd.Wait()
}

func (b *BuildSystem) Build(ctx context.Context, customCommand string) (string, error) {

	pm := b.DetectPackageManager()

	if fileExists(filepath.Join(b.RootDir, "package.json")) {
		installArgs := b.InstallArgs(pm)

		if b.Logger != nil {
			b.Logger(fmt.Sprintf("Installing dependencies with %s %v...\n", pm, installArgs))
		}

		if err := b.RunCommand(ctx, pm, installArgs...); err != nil {
			return "", fmt.Errorf("install failed: %w", err)
		}
	}

	if customCommand != "" {
		if err := validateBuildCommand(customCommand); err != nil {
			return "", fmt.Errorf("invalid build command: %w", err)
		}
		if b.Logger != nil {
			b.Logger(fmt.Sprintf("Running custom build command: %s\n", customCommand))
		}
		if err := b.RunCommand(ctx, "sh", "-c", customCommand); err != nil {
			return "", fmt.Errorf("build failed: %w", err)
		}
	} else if fileExists(filepath.Join(b.RootDir, "package.json")) {

		buildArgs := b.BuildArgs(pm)
		if b.Logger != nil {
			b.Logger(fmt.Sprintf("Building with %s %v...\n", pm, buildArgs))
		}
		if err := b.RunCommand(ctx, pm, buildArgs...); err != nil {
			return "", fmt.Errorf("build failed: %w", err)
		}
	}

	return b.DetectOutputDirectory()
}

func (b *BuildSystem) DetectOutputDirectory() (string, error) {
	candidates := []string{"dist", "build", "public", ".svelte-kit/output", "out", "_site"}
	for _, c := range candidates {
		path := filepath.Join(b.RootDir, c)

		if fileExists(path) {

			return c, nil
		}
	}

	if fileExists(filepath.Join(b.RootDir, "index.html")) {
		if b.Logger != nil {
			b.Logger("No build directory detected, but index.html found. Using root directory.")
		}
		return ".", nil
	}

	return "", fmt.Errorf("could not detect build output directory")
}
