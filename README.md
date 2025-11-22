# file-syncer

A Go application to synchronize files in a folder with a GitHub repository.

## Features

- **Push Mode**: Synchronize local files to a GitHub repository
- **Pull Mode**: Synchronize files from a GitHub repository to a local folder
- Supports custom branch selection
- Automatic handling of repository cloning and commits
- Skips `.git` directory during synchronization

## Installation

### Prerequisites

- Go 1.16 or later
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

## Testing

Run the test suite:

```bash
go test -v ./...
```

## Notes

- The application uses temporary directories for git operations, which are cleaned up automatically
- Git credentials should be configured on your system (via SSH keys or credential helper)
- The `.git` directory is always excluded from synchronization
- For push mode, if there are no changes, no commit or push will be performed

## License

MIT
