package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	ModePush = "push"
	ModePull = "pull"
)

type Config struct {
	Mode       string
	FolderPath string
	RepoURL    string
	Branch     string
}

func main() {
	config := parseFlags()

	if err := validateConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	if err := run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() Config {
	config := Config{}

	flag.StringVar(&config.Mode, "mode", "", "Operation mode: 'push' or 'pull'")
	flag.StringVar(&config.FolderPath, "folder", "", "Path to the folder to sync")
	flag.StringVar(&config.RepoURL, "repo", "", "GitHub repository URL")
	flag.StringVar(&config.Branch, "branch", "main", "Git branch to use (default: main)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  Push files to repository:\n")
		fmt.Fprintf(os.Stderr, "    %s -mode push -folder ./myfiles -repo https://github.com/user/repo.git\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Pull files from repository:\n")
		fmt.Fprintf(os.Stderr, "    %s -mode pull -folder ./myfiles -repo https://github.com/user/repo.git\n\n", os.Args[0])
	}

	flag.Parse()
	return config
}

func validateConfig(config Config) error {
	if config.Mode != ModePush && config.Mode != ModePull {
		return fmt.Errorf("mode must be either 'push' or 'pull'")
	}

	if config.FolderPath == "" {
		return fmt.Errorf("folder path is required")
	}

	if config.RepoURL == "" {
		return fmt.Errorf("repository URL is required")
	}

	return nil
}

func run(config Config) error {
	fmt.Printf("File Syncer - Mode: %s\n", config.Mode)
	fmt.Printf("Folder: %s\n", config.FolderPath)
	fmt.Printf("Repository: %s\n", config.RepoURL)
	fmt.Printf("Branch: %s\n\n", config.Branch)

	if config.Mode == ModePush {
		return pushFiles(config)
	}
	return pullFiles(config)
}

func pushFiles(config Config) error {
	fmt.Println("Starting push operation...")

	// Create absolute path for folder
	absPath, err := filepath.Abs(config.FolderPath)
	if err != nil {
		return fmt.Errorf("failed to resolve folder path: %w", err)
	}

	// Check if folder exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("folder does not exist: %s", absPath)
	}

	// Create temporary directory for git operations
	tempDir, err := os.MkdirTemp("", "file-syncer-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone the repository
	fmt.Println("Cloning repository...")
	if err := runCommand(tempDir, "git", "clone", "--branch", config.Branch, config.RepoURL, "."); err != nil {
		// Try cloning without branch if it doesn't exist
		fmt.Printf("Branch '%s' not found, cloning default branch...\n", config.Branch)
		if err := runCommand(tempDir, "git", "clone", config.RepoURL, "."); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
		// Create and checkout the branch
		if err := runCommand(tempDir, "git", "checkout", "-b", config.Branch); err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}
	}

	// Sync files from source folder to repo
	fmt.Println("Syncing files...")
	if err := syncFiles(absPath, tempDir); err != nil {
		return fmt.Errorf("failed to sync files: %w", err)
	}

	// Check if there are changes
	output, err := runCommandOutput(tempDir, "git", "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	if strings.TrimSpace(output) == "" {
		fmt.Println("No changes to push.")
		return nil
	}

	// Add all changes
	fmt.Println("Adding changes...")
	if err := runCommand(tempDir, "git", "add", "-A"); err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	// Commit changes
	fmt.Println("Committing changes...")
	if err := runCommand(tempDir, "git", "commit", "-m", "Sync files from local folder"); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Push to remote
	fmt.Println("Pushing to remote...")
	if err := runCommand(tempDir, "git", "push", "origin", config.Branch); err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}

	fmt.Println("Push completed successfully!")
	return nil
}

func pullFiles(config Config) error {
	fmt.Println("Starting pull operation...")

	// Create absolute path for folder
	absPath, err := filepath.Abs(config.FolderPath)
	if err != nil {
		return fmt.Errorf("failed to resolve folder path: %w", err)
	}

	// Create folder if it doesn't exist
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}

	// Create temporary directory for git operations
	tempDir, err := os.MkdirTemp("", "file-syncer-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone the repository
	fmt.Println("Cloning repository...")
	if err := runCommand(tempDir, "git", "clone", "--branch", config.Branch, config.RepoURL, "."); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Sync files from repo to destination folder
	fmt.Println("Syncing files...")
	if err := syncFiles(tempDir, absPath); err != nil {
		return fmt.Errorf("failed to sync files: %w", err)
	}

	fmt.Println("Pull completed successfully!")
	return nil
}

func syncFiles(srcDir, dstDir string) error {
	// Walk through source directory
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Skip .git directory
		if strings.HasPrefix(relPath, ".git") || relPath == ".git" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		dstPath := filepath.Join(dstDir, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		return copyFile(path, dstPath, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy contents
	_, err = io.Copy(dstFile, srcFile)
	return err
}

func runCommand(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCommandOutput(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return string(output), err
}
