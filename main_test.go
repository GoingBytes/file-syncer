package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid push config",
			config: Config{
				Mode:       ModePush,
				FolderPath: "/tmp/test",
				RepoURL:    "https://github.com/user/repo.git",
				Branch:     "main",
			},
			wantErr: false,
		},
		{
			name: "valid pull config",
			config: Config{
				Mode:       ModePull,
				FolderPath: "/tmp/test",
				RepoURL:    "https://github.com/user/repo.git",
				Branch:     "main",
			},
			wantErr: false,
		},
		{
			name: "invalid mode",
			config: Config{
				Mode:       "invalid",
				FolderPath: "/tmp/test",
				RepoURL:    "https://github.com/user/repo.git",
			},
			wantErr: true,
		},
		{
			name: "missing folder path",
			config: Config{
				Mode:    ModePush,
				RepoURL: "https://github.com/user/repo.git",
			},
			wantErr: true,
		},
		{
			name: "missing repo URL",
			config: Config{
				Mode:       ModePush,
				FolderPath: "/tmp/test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSyncFiles(t *testing.T) {
	// Create source directory with test files
	srcDir, err := os.MkdirTemp("", "sync-src-*")
	if err != nil {
		t.Fatalf("failed to create temp source dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	// Create destination directory
	dstDir, err := os.MkdirTemp("", "sync-dst-*")
	if err != nil {
		t.Fatalf("failed to create temp destination dir: %v", err)
	}
	defer os.RemoveAll(dstDir)

	// Create test files in source
	testFile1 := filepath.Join(srcDir, "test1.txt")
	if err := os.WriteFile(testFile1, []byte("test content 1"), 0644); err != nil {
		t.Fatalf("failed to create test file 1: %v", err)
	}

	testDir := filepath.Join(srcDir, "subdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	testFile2 := filepath.Join(testDir, "test2.txt")
	if err := os.WriteFile(testFile2, []byte("test content 2"), 0644); err != nil {
		t.Fatalf("failed to create test file 2: %v", err)
	}

	// Sync files
	if err := syncFiles(srcDir, dstDir); err != nil {
		t.Fatalf("syncFiles() failed: %v", err)
	}

	// Verify files were synced
	dstFile1 := filepath.Join(dstDir, "test1.txt")
	content1, err := os.ReadFile(dstFile1)
	if err != nil {
		t.Errorf("failed to read destination file 1: %v", err)
	}
	if string(content1) != "test content 1" {
		t.Errorf("file 1 content mismatch: got %q, want %q", string(content1), "test content 1")
	}

	dstFile2 := filepath.Join(dstDir, "subdir", "test2.txt")
	content2, err := os.ReadFile(dstFile2)
	if err != nil {
		t.Errorf("failed to read destination file 2: %v", err)
	}
	if string(content2) != "test content 2" {
		t.Errorf("file 2 content mismatch: got %q, want %q", string(content2), "test content 2")
	}
}

func TestSyncFilesSkipsGitDirectory(t *testing.T) {
	// Create source directory with .git directory
	srcDir, err := os.MkdirTemp("", "sync-src-git-*")
	if err != nil {
		t.Fatalf("failed to create temp source dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	// Create destination directory
	dstDir, err := os.MkdirTemp("", "sync-dst-git-*")
	if err != nil {
		t.Fatalf("failed to create temp destination dir: %v", err)
	}
	defer os.RemoveAll(dstDir)

	// Create .git directory in source
	gitDir := filepath.Join(srcDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	gitFile := filepath.Join(gitDir, "config")
	if err := os.WriteFile(gitFile, []byte("git config"), 0644); err != nil {
		t.Fatalf("failed to create git config file: %v", err)
	}

	// Create regular file
	testFile := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Sync files
	if err := syncFiles(srcDir, dstDir); err != nil {
		t.Fatalf("syncFiles() failed: %v", err)
	}

	// Verify .git directory was not synced
	dstGitDir := filepath.Join(dstDir, ".git")
	if _, err := os.Stat(dstGitDir); !os.IsNotExist(err) {
		t.Errorf(".git directory should not be synced, but it exists")
	}

	// Verify regular file was synced
	dstFile := filepath.Join(dstDir, "test.txt")
	if _, err := os.Stat(dstFile); err != nil {
		t.Errorf("regular file should be synced: %v", err)
	}
}

func TestCopyFile(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "copy-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create source file
	srcFile := filepath.Join(tempDir, "source.txt")
	testContent := "test file content"
	if err := os.WriteFile(srcFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Copy file
	dstFile := filepath.Join(tempDir, "destination.txt")
	if err := copyFile(srcFile, dstFile, 0644); err != nil {
		t.Fatalf("copyFile() failed: %v", err)
	}

	// Verify destination file
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("content mismatch: got %q, want %q", string(content), testContent)
	}
}

func TestValidateConfigWithSSHKey(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config with SSH key",
			config: Config{
				Mode:       ModePush,
				FolderPath: "/tmp/test",
				RepoURL:    "git@github.com:user/repo.git",
				Branch:     "main",
				SSHKeyPath: "/home/user/.ssh/id_rsa",
			},
			wantErr: false,
		},
		{
			name: "valid config without SSH key",
			config: Config{
				Mode:       ModePush,
				FolderPath: "/tmp/test",
				RepoURL:    "https://github.com/user/repo.git",
				Branch:     "main",
				SSHKeyPath: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEscapeShellArg(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple path",
			input: "/home/user/.ssh/id_rsa",
			want:  "/home/user/.ssh/id_rsa",
		},
		{
			name:  "path with spaces",
			input: "/home/user/my files/.ssh/id_rsa",
			want:  "/home/user/my\\ files/.ssh/id_rsa",
		},
		{
			name:  "path with single quote",
			input: "/home/user's/.ssh/id_rsa",
			want:  "/home/user\\'s/.ssh/id_rsa",
		},
		{
			name:  "path with special chars",
			input: "/home/user/.ssh/key$file",
			want:  "/home/user/.ssh/key\\$file",
		},
		{
			name:  "path with multiple special chars",
			input: "/home/user name/.ssh/key file (1).pem",
			want:  "/home/user\\ name/.ssh/key\\ file\\ \\(1\\).pem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeShellArg(tt.input)
			if got != tt.want {
				t.Errorf("escapeShellArg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildGitSSHCommand(t *testing.T) {
	tests := []struct {
		name       string
		sshKeyPath string
		want       string
	}{
		{
			name:       "simple path",
			sshKeyPath: "/home/user/.ssh/id_rsa",
			want:       "ssh -i /home/user/.ssh/id_rsa -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new",
		},
		{
			name:       "path with spaces",
			sshKeyPath: "/home/user/my files/.ssh/id_rsa",
			want:       "ssh -i /home/user/my\\ files/.ssh/id_rsa -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new",
		},
		{
			name:       "path with special chars",
			sshKeyPath: "/home/user's key/.ssh/deploy (prod).pem",
			want:       "ssh -i /home/user\\'s\\ key/.ssh/deploy\\ \\(prod\\).pem -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildGitSSHCommand(tt.sshKeyPath)
			if got != tt.want {
				t.Errorf("buildGitSSHCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseGitStatus(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   FileChangeStats
	}{
		{
			name:   "single added file",
			output: "A  newfile.txt",
			want: FileChangeStats{
				Added:    []string{"newfile.txt"},
				Modified: []string{},
				Deleted:  []string{},
			},
		},
		{
			name:   "single modified file",
			output: "M  modified.txt",
			want: FileChangeStats{
				Added:    []string{},
				Modified: []string{"modified.txt"},
				Deleted:  []string{},
			},
		},
		{
			name:   "single deleted file",
			output: "D  deleted.txt",
			want: FileChangeStats{
				Added:    []string{},
				Modified: []string{},
				Deleted:  []string{"deleted.txt"},
			},
		},
		{
			name:   "mixed changes",
			output: "A  added.txt\nM  modified.txt\nD  deleted.txt",
			want: FileChangeStats{
				Added:    []string{"added.txt"},
				Modified: []string{"modified.txt"},
				Deleted:  []string{"deleted.txt"},
			},
		},
		{
			name:   "untracked file",
			output: "?? untracked.txt",
			want: FileChangeStats{
				Added:    []string{"untracked.txt"},
				Modified: []string{},
				Deleted:  []string{},
			},
		},
		{
			name:   "multiple files of same type",
			output: "A  file1.txt\nA  file2.txt\nM  file3.txt",
			want: FileChangeStats{
				Added:    []string{"file1.txt", "file2.txt"},
				Modified: []string{"file3.txt"},
				Deleted:  []string{},
			},
		},
		{
			name:   "modified with space prefix",
			output: " M modified.txt",
			want: FileChangeStats{
				Added:    []string{},
				Modified: []string{"modified.txt"},
				Deleted:  []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseGitStatus(tt.output)
			if len(got.Added) != len(tt.want.Added) {
				t.Errorf("parseGitStatus() added count = %v, want %v", len(got.Added), len(tt.want.Added))
			}
			if len(got.Modified) != len(tt.want.Modified) {
				t.Errorf("parseGitStatus() modified count = %v, want %v", len(got.Modified), len(tt.want.Modified))
			}
			if len(got.Deleted) != len(tt.want.Deleted) {
				t.Errorf("parseGitStatus() deleted count = %v, want %v", len(got.Deleted), len(tt.want.Deleted))
			}
			
			// Check individual files
			for i, file := range tt.want.Added {
				if i >= len(got.Added) || got.Added[i] != file {
					t.Errorf("parseGitStatus() added[%d] = %v, want %v", i, got.Added, tt.want.Added)
					break
				}
			}
			for i, file := range tt.want.Modified {
				if i >= len(got.Modified) || got.Modified[i] != file {
					t.Errorf("parseGitStatus() modified[%d] = %v, want %v", i, got.Modified, tt.want.Modified)
					break
				}
			}
			for i, file := range tt.want.Deleted {
				if i >= len(got.Deleted) || got.Deleted[i] != file {
					t.Errorf("parseGitStatus() deleted[%d] = %v, want %v", i, got.Deleted, tt.want.Deleted)
					break
				}
			}
		})
	}
}

func TestGenerateCommitMessage(t *testing.T) {
	tests := []struct {
		name        string
		stats       FileChangeStats
		wantSubject string
		wantBody    string
	}{
		{
			name: "single added file",
			stats: FileChangeStats{
				Added:    []string{"file.txt"},
				Modified: []string{},
				Deleted:  []string{},
			},
			wantSubject: "Sync 1 file (1 added)",
			wantBody:    "Added files:\n  + file.txt",
		},
		{
			name: "single modified file",
			stats: FileChangeStats{
				Added:    []string{},
				Modified: []string{"file.txt"},
				Deleted:  []string{},
			},
			wantSubject: "Sync 1 file (1 modified)",
			wantBody:    "Modified files:\n  ~ file.txt",
		},
		{
			name: "single deleted file",
			stats: FileChangeStats{
				Added:    []string{},
				Modified: []string{},
				Deleted:  []string{"file.txt"},
			},
			wantSubject: "Sync 1 file (1 deleted)",
			wantBody:    "Deleted files:\n  - file.txt",
		},
		{
			name: "multiple files mixed",
			stats: FileChangeStats{
				Added:    []string{"new1.txt", "new2.txt"},
				Modified: []string{"mod.txt"},
				Deleted:  []string{"old.txt"},
			},
			wantSubject: "Sync 4 files (2 added, 1 modified, 1 deleted)",
			wantBody:    "Added files:\n  + new1.txt\n  + new2.txt\n\nModified files:\n  ~ mod.txt\n\nDeleted files:\n  - old.txt",
		},
		{
			name: "only modified files",
			stats: FileChangeStats{
				Added:    []string{},
				Modified: []string{"file1.txt", "file2.txt"},
				Deleted:  []string{},
			},
			wantSubject: "Sync 2 files (2 modified)",
			wantBody:    "Modified files:\n  ~ file1.txt\n  ~ file2.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSubject, gotBody := generateCommitMessage(tt.stats)
			if gotSubject != tt.wantSubject {
				t.Errorf("generateCommitMessage() subject = %v, want %v", gotSubject, tt.wantSubject)
			}
			if gotBody != tt.wantBody {
				t.Errorf("generateCommitMessage() body = %v, want %v", gotBody, tt.wantBody)
			}
		})
	}
}
