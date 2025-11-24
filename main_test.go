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
