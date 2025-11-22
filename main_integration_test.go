//go:build integration

package main

import (
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPushIntegrationPushesFilesToRemote(t *testing.T) {
	requireGit(t)
	useTestLogger(t)
	setGitIdentityEnv(t)

	remote := createRemoteRepoWithContent(t, map[string]string{
		"seed.txt": "initial content",
	})

	sourceDir := t.TempDir()
	writeTestFile(t, sourceDir, "new-file.txt", "integration content")

	config := Config{
		Mode:       ModePush,
		FolderPath: sourceDir,
		RepoURL:    remote,
		Branch:     "main",
	}

	if err := run(config); err != nil {
		t.Fatalf("run() push failed: %v", err)
	}

	verificationDir := t.TempDir()
	runGit(t, verificationDir, "clone", "--branch", "main", remote, ".")

	content, err := os.ReadFile(filepath.Join(verificationDir, "new-file.txt"))
	if err != nil {
		t.Fatalf("failed to read synced file: %v", err)
	}

	if string(content) != "integration content" {
		t.Fatalf("unexpected file content: %q", string(content))
	}
}

func TestPullIntegrationPullsFilesFromRemote(t *testing.T) {
	requireGit(t)
	useTestLogger(t)
	setGitIdentityEnv(t)

	remote := createRemoteRepoWithContent(t, map[string]string{
		"pull-dir/file.txt": "pulled content",
	})

	destinationDir := t.TempDir()

	config := Config{
		Mode:       ModePull,
		FolderPath: destinationDir,
		RepoURL:    remote,
		Branch:     "main",
	}

	if err := run(config); err != nil {
		t.Fatalf("run() pull failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destinationDir, "pull-dir", "file.txt"))
	if err != nil {
		t.Fatalf("failed to read pulled file: %v", err)
	}

	if string(content) != "pulled content" {
		t.Fatalf("unexpected pulled content: %q", string(content))
	}

	if _, err := os.Stat(filepath.Join(destinationDir, ".git")); !os.IsNotExist(err) {
		t.Fatalf(".git directory should not be present in destination")
	}
}

func createRemoteRepoWithContent(t *testing.T, files map[string]string) string {
	t.Helper()

	baseDir := t.TempDir()
	remotePath := filepath.Join(baseDir, "remote.git")
	runGit(t, baseDir, "init", "--bare", remotePath)

	workingDir := t.TempDir()
	runGit(t, workingDir, "init")
	runGit(t, workingDir, "config", "user.email", "file-syncer@example.com")
	runGit(t, workingDir, "config", "user.name", "file-syncer")

	for path, content := range files {
		writeTestFile(t, workingDir, path, content)
	}

	runGit(t, workingDir, "add", ".")
	runGit(t, workingDir, "commit", "-m", "seed")
	runGit(t, workingDir, "branch", "-M", "main")
	runGit(t, workingDir, "remote", "add", "origin", remotePath)
	runGit(t, workingDir, "push", "-u", "origin", "main")
	runGit(t, remotePath, "symbolic-ref", "HEAD", "refs/heads/main")

	return remotePath
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\nOutput: %s", strings.Join(args, " "), err, string(output))
	}
}

func writeTestFile(t *testing.T, baseDir, relativePath, content string) {
	t.Helper()

	fullPath := filepath.Join(baseDir, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("failed to create directories for %s: %v", relativePath, err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", relativePath, err)
	}
}

func useTestLogger(t *testing.T) {
	t.Helper()
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))
}

func setGitIdentityEnv(t *testing.T) {
	t.Helper()

	t.Setenv("GIT_AUTHOR_NAME", "file-syncer")
	t.Setenv("GIT_AUTHOR_EMAIL", "file-syncer@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "file-syncer")
	t.Setenv("GIT_COMMITTER_EMAIL", "file-syncer@example.com")
}

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}
}
