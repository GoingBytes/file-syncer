# file-syncer

A Go application to synchronize files in a folder with a GitHub repository.

## Features

- **Push Mode**: Synchronize local files to a GitHub repository
- **Pull Mode**: Synchronize files from a GitHub repository to a local folder
- **Private Repository Support**: Works with private GitHub repositories using system git credentials
- **Structured Logging**: JSON-formatted logs with automatic rotation
  - Maximum log file size: 10MB
  - Maximum log age: 28 days
  - Maximum number of backup files: 3
  - Automatic compression of rotated logs
  - Logs to both stdout and `file-syncer.log`
- Supports custom branch selection
- Automatic handling of repository cloning and commits
- Skips `.git` directory during synchronization

## Installation

### Prerequisites

- Go 1.25.4 or later
- Git

### Build from Source

```bash
git clone https://github.com/rikkicom/file-syncer.git
cd file-syncer
go build -o file-syncer .
```

## Usage

### Command-Line Options

```
-mode string
    Operation mode: 'push' or 'pull' (required)
-folder string
    Path to the folder to sync (required)
-repo string
    GitHub repository URL (required)
-branch string
    Git branch to use (default: "main")
```

### Push Mode

Push local files from a folder to a GitHub repository:

```bash
./file-syncer -mode push -folder ./myfiles -repo https://github.com/user/repo.git
```

With a specific branch:

```bash
./file-syncer -mode push -folder ./myfiles -repo https://github.com/user/repo.git -branch develop
```

### Pull Mode

Pull files from a GitHub repository to a local folder:

```bash
./file-syncer -mode pull -folder ./myfiles -repo https://github.com/user/repo.git
```

With a specific branch:

```bash
./file-syncer -mode pull -folder ./myfiles -repo https://github.com/user/repo.git -branch develop
```

## How It Works

### Push Mode

1. Clones the specified repository to a temporary directory
2. Syncs all files from your local folder to the cloned repository (excluding `.git`)
3. Commits the changes with message "Sync files from local folder"
4. Pushes the changes to the remote repository

### Pull Mode

1. Clones the specified repository to a temporary directory
2. Syncs all files from the repository to your local folder (excluding `.git`)
3. Creates the destination folder if it doesn't exist

## Examples

### Example 1: Backing up local files to GitHub

```bash
# First time: push files to a new repository
./file-syncer -mode push -folder ~/documents -repo https://github.com/yourusername/my-backup.git

# Later: push updates
./file-syncer -mode push -folder ~/documents -repo https://github.com/yourusername/my-backup.git
```

### Example 2: Syncing files between machines

On machine 1:
```bash
./file-syncer -mode push -folder ~/projects/shared -repo https://github.com/yourusername/shared-files.git
```

On machine 2:
```bash
./file-syncer -mode pull -folder ~/projects/shared -repo https://github.com/yourusername/shared-files.git
```

## Private Repository Authentication

The application supports both public and private GitHub repositories. For private repositories, ensure your system is configured with appropriate git credentials:

### SSH Keys (Recommended)

```bash
# Use SSH URL format
./file-syncer -mode push -folder ./myfiles -repo git@github.com:yourusername/private-repo.git
```

### HTTPS with Credential Helper

```bash
# Configure git credential helper (one-time setup)
git config --global credential.helper store

# Or use GitHub CLI for authentication
gh auth login

# Then use HTTPS URL
./file-syncer -mode push -folder ./myfiles -repo https://github.com/yourusername/private-repo.git
```

### Personal Access Token

For HTTPS URLs, you can embed credentials or use a credential helper. The application inherits all git configuration from your system.

## Logging

The application uses structured JSON logging with automatic rotation:

- **Log file**: `file-syncer.log` (created in the current directory)
- **Max size**: 10MB per file
- **Max age**: 28 days
- **Retention**: 3 backup files
- **Compression**: Old logs are automatically gzipped
- **Output**: Logs are written to both stdout and the log file

Example log entry:
```json
{"time":"2025-11-22T19:35:58.101Z","level":"INFO","msg":"File Syncer started","mode":"push","folder":"/path/to/folder","repository":"https://github.com/user/repo.git","branch":"main"}
```

## Testing

Run the test suite:

```bash
go test -v ./...
```

## Notes

- The application uses temporary directories for git operations, which are cleaned up automatically
- Git credentials are inherited from your system's git configuration
- The `.git` directory is always excluded from synchronization
- For push mode, if there are no changes, no commit or push will be performed
- All operations are logged with structured JSON format for easy parsing and monitoring
