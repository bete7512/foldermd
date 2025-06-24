# foldermd

> 🚀 A powerful CLI tool that generates beautiful README files from your project's folder structure

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/bete7512/foldermd)](https://github.com/bete7512/foldermd/releases)

Transform any project directory into a comprehensive, well-structured README with just one command. Perfect for documenting codebases, project structures, and creating professional documentation.

## ✨ Features

- 🌳 **Beautiful tree structure visualization** - Clean, hierarchical project layout
- 📄 **File content inclusion** - Include actual file contents with syntax highlighting
- 🎯 **Smart filtering** - Automatically ignores common files (`.git`, `node_modules`, etc.)
- 📋 **Custom ignore patterns** - Use `.foldermd.ignore` file or command flags
- 🔍 **Hidden file support** - Optionally include hidden files and directories
- 📏 **Depth control** - Limit traversal depth for large projects
- 📊 **Project statistics** - Automatic file and directory counts
- 🎨 **Rich formatting** - Professional markdown output with emojis and sections
- 🛠️ **Flexible configuration** - Extensive CLI options for customization

## 🚀 Installation

### Option 1: Install from GitHub (Recommended)

#### Using Go Install (Go 1.16+)
```bash
go install github.com/bete7512/foldermd@latest
```

#### Download Pre-built Binaries
```bash
# Download latest release for your platform
curl -L https://github.com/bete7512/foldermd/releases/latest/download/foldermd-linux-amd64 -o foldermd
chmod +x foldermd
sudo mv foldermd /usr/local/bin/
```

#### Available Platforms
- Linux (amd64, arm64)
- macOS (amd64, arm64/M1)
- Windows (amd64)

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/bete7512/foldermd.git
cd foldermd

# Install dependencies
go mod tidy

# Build the binary
go build -o foldermd

# Install globally (optional)
sudo mv foldermd /usr/local/bin/
```

### Option 3: Using Homebrew (macOS/Linux)

```bash
# Add tap (coming soon)
brew tap bete7512/tools
brew install foldermd
```

## 📖 Quick Start

```bash
# Generate README for current directory
foldermd

# Include files in the tree structure
foldermd --files

# Include file contents with syntax highlighting
foldermd --content

# Generate for specific directory with custom output
foldermd /path/to/project --output PROJECT.md
```

## 🎯 Usage

### Basic Commands

```bash
foldermd [directory] [flags]
```

### Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--files` | `-f` | `false` | Include files in tree structure |
| `--content` | `-c` | `false` | Include file contents (implies `--files`) |
| `--output` | `-o` | `README.md` | Output file name |
| `--ignore` | `-i` | `.git,.DS_Store,node_modules,*.log` | Comma-separated ignore patterns |
| `--depth` | `-d` | `-1` | Maximum depth (-1 for unlimited) |
| `--hidden` | | `false` | Include hidden files and directories |

### Subcommands

- `foldermd version` - Show version information
- `foldermd init` - Create a `.foldermd.ignore` file
- `foldermd --help` - Show detailed help

### Advanced Exam