package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	version = "1.0.0"
)

type Config struct {
	includeFiles   bool
	includeContent bool
	outputFile     string
	ignorePatterns []string
	maxDepth       int
	showHidden     bool
	targetDir      string
}

var config Config

var rootCmd = &cobra.Command{
	Use:   "foldermd [directory]",
	Short: "Generate README from folder structure",
	Long: `foldermd is a CLI tool that generates a README.md file from your current 
folder structure with optional file content inclusion.

The tool creates a beautifully formatted README with:
- Project structure tree visualization
- Optional file content with syntax highlighting  
- Smart filtering of common ignore patterns
- Customizable output options`,
	Example: `  # Generate README for current directory
  foldermd

  # Generate with files included
  foldermd --files

  # Generate with file contents and custom output name
  foldermd --content --output PROJECT.md

  # Generate for specific directory with depth limit
  foldermd /path/to/project --files --depth 3

  # Custom ignore patterns
  foldermd --files --ignore ".git,*.log,build,dist"`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Set target directory
		config.targetDir = "."
		if len(args) > 0 {
			config.targetDir = args[0]
		}

		// Validate target directory exists
		if _, err := os.Stat(config.targetDir); os.IsNotExist(err) {
			return fmt.Errorf("directory '%s' does not exist", config.targetDir)
		}

		// Parse ignore patterns
		ignoreList, _ := cmd.Flags().GetString("ignore")
		if ignoreList != "" {
			config.ignorePatterns = strings.Split(ignoreList, ",")
			// Trim whitespace from patterns
			for i, pattern := range config.ignorePatterns {
				config.ignorePatterns[i] = strings.TrimSpace(pattern)
			}
		}

		// If content is requested, automatically include files
		if config.includeContent {
			config.includeFiles = true
		}

		return generateReadme(config)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  "Print the version number of foldermd",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("foldermd v%s\n", version)
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a sample .foldermd.ignore file",
	Long:  "Create a .foldermd.ignore file in the current directory with common ignore patterns",
	RunE: func(cmd *cobra.Command, args []string) error {
		return createIgnoreFile()
	},
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&config.includeFiles, "files", "f", false, "Include files in the tree structure")
	rootCmd.PersistentFlags().BoolVarP(&config.includeContent, "content", "c", false, "Include file contents with syntax highlighting (implies --files)")
	rootCmd.PersistentFlags().StringVarP(&config.outputFile, "output", "o", "README.md", "Output README file name")
	rootCmd.PersistentFlags().StringP("ignore", "i", ".git,.DS_Store,node_modules,*.log", "Comma-separated patterns to ignore")
	rootCmd.PersistentFlags().IntVarP(&config.maxDepth, "depth", "d", -1, "Maximum directory depth to traverse (-1 for unlimited)")
	rootCmd.PersistentFlags().BoolVar(&config.showHidden, "hidden", false, "Include hidden files and directories")

	// Flag descriptions and examples
	rootCmd.Flags().SetInterspersed(false)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func generateReadme(config Config) error {
	fmt.Printf("🚀 Generating README for: %s\n", config.targetDir)
	
	file, err := os.Create(config.outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file '%s': %w", config.outputFile, err)
	}
	defer file.Close()

	// Get project name from directory
	projectName := filepath.Base(config.targetDir)
	if projectName == "." {
		cwd, _ := os.Getwd()
		projectName = filepath.Base(cwd)
	}

	// Check if there's a .foldermd.ignore file
	ignoreFile := filepath.Join(config.targetDir, ".foldermd.ignore")
	if _, err := os.Stat(ignoreFile); err == nil {
		additionalPatterns, err := readIgnoreFile(ignoreFile)
		if err == nil {
			config.ignorePatterns = append(config.ignorePatterns, additionalPatterns...)
			fmt.Printf("📋 Loaded ignore patterns from .foldermd.ignore\n")
		}
	}

	// Write header with more details
	fmt.Fprintf(file, "# %s\n\n", projectName)
	fmt.Fprintf(file, "> Generated with [foldermd](https://github.com/yourusername/foldermd) on %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	
	// Add basic project info
	writeProjectInfo(file, config.targetDir)
	
	fmt.Fprintf(file, "## 📁 Project Structure\n\n")
	
	// Add legend if files are included
	if config.includeFiles {
		fmt.Fprintf(file, "```\n")
		fmt.Fprintf(file, "Legend: 📁 Directory | 📄 File\n")
		fmt.Fprintf(file, "```\n\n")
	}
	
	fmt.Fprintf(file, "```\n")

	// Generate tree structure
	fmt.Printf("🌳 Building project tree...\n")
	if err := writeTree(file, config.targetDir, "", config, 0); err != nil {
		return fmt.Errorf("failed to generate tree: %w", err)
	}

	fmt.Fprintf(file, "```\n\n")

	// Add file contents if requested
	if config.includeContent {
		fmt.Printf("📄 Including file contents...\n")
		fmt.Fprintf(file, "## 📄 File Contents\n\n")
		if err := writeFileContents(file, config.targetDir, config, 0); err != nil {
			return fmt.Errorf("failed to write file contents: %w", err)
		}
	}

	// Add footer
	writeFooter(file, config)

	fmt.Printf("✅ README generated successfully: %s\n", config.outputFile)
	return nil
}

func writeProjectInfo(file *os.File, dir string) {
	// Count files and directories
	var fileCount, dirCount int
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			dirCount++
		} else {
			fileCount++
		}
		return nil
	})

	fmt.Fprintf(file, "## 📊 Project Overview\n\n")
	fmt.Fprintf(file, "- **Total Files:** %d\n", fileCount)
	fmt.Fprintf(file, "- **Total Directories:** %d\n", dirCount-1) // Subtract root directory
	fmt.Fprintf(file, "- **Project Root:** `%s`\n\n", dir)
}

func writeTree(file *os.File, dir string, prefix string, config Config, depth int) error {
	if config.maxDepth >= 0 && depth > config.maxDepth {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Filter and sort entries
	var filteredEntries []fs.DirEntry
	for _, entry := range entries {
		if shouldIgnore(entry.Name(), config) {
			continue
		}
		if !config.showHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if !config.includeFiles && !entry.IsDir() {
			continue
		}
		filteredEntries = append(filteredEntries, entry)
	}

	// Sort: directories first, then files
	sort.Slice(filteredEntries, func(i, j int) bool {
		if filteredEntries[i].IsDir() != filteredEntries[j].IsDir() {
			return filteredEntries[i].IsDir()
		}
		return filteredEntries[i].Name() < filteredEntries[j].Name()
	})

	for i, entry := range filteredEntries {
		isLast := i == len(filteredEntries)-1
		
		// Choose the appropriate tree characters
		var connector, nextPrefix string
		if isLast {
			connector = "└── "
			nextPrefix = prefix + "    "
		} else {
			connector = "├── "
			nextPrefix = prefix + "│   "
		}

		// Write the entry with emoji if content mode
		name := entry.Name()
		if config.includeContent {
			if entry.IsDir() {
				name = "📁 " + name + "/"
			} else {
				name = "📄 " + name
			}
		} else if entry.IsDir() {
			name += "/"
		}
		
		fmt.Fprintf(file, "%s%s%s\n", prefix, connector, name)

		// Recurse into directories
		if entry.IsDir() {
			subDir := filepath.Join(dir, entry.Name())
			if err := writeTree(file, subDir, nextPrefix, config, depth+1); err != nil {
				return err
			}
		}
	}

	return nil
}

func writeFileContents(file *os.File, dir string, config Config, depth int) error {
	if config.maxDepth >= 0 && depth > config.maxDepth {
		return nil
	}

	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip if beyond max depth
		relPath, _ := filepath.Rel(dir, path)
		currentDepth := len(strings.Split(relPath, string(filepath.Separator))) - 1
		if config.maxDepth >= 0 && currentDepth > config.maxDepth {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Skip directories and ignored files
		if d.IsDir() || shouldIgnore(d.Name(), config) {
			return nil
		}

		// Skip hidden files if not requested
		if !config.showHidden && strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		// Get relative path for display
		relPath, _ = filepath.Rel(dir, path)
		if relPath == "." {
			relPath = d.Name()
		}

		// Get file info
		info, err := d.Info()
		if err != nil {
			return nil
		}

		// Check file size (skip very large files)
		if info.Size() > 1024*1024 { // 1MB limit
			fmt.Fprintf(file, "### 📄 %s\n\n*File too large to display (%s)*\n\n", relPath, formatFileSize(info.Size()))
			return nil
		}

		// Check if file is text-based
		if !isTextFile(path) {
			fmt.Fprintf(file, "### 📄 %s\n\n*Binary file (%s)*\n\n", relPath, formatFileSize(info.Size()))
			return nil
		}

		// Read and write file contents
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(file, "### 📄 %s\n\n*Error reading file: %v*\n\n", relPath, err)
			return nil
		}

		// Detect file type for syntax highlighting
		ext := strings.ToLower(filepath.Ext(path))
		lang := getLanguageFromExtension(ext)

		fmt.Fprintf(file, "### 📄 %s\n\n", relPath)
		fmt.Fprintf(file, "*Size: %s | Language: %s*\n\n", formatFileSize(info.Size()), lang)
		fmt.Fprintf(file, "```%s\n%s\n```\n\n", lang, string(content))

		return nil
	})
}

func writeFooter(file *os.File, config Config) {
	fmt.Fprintf(file, "---\n\n")
	fmt.Fprintf(file, "## 🛠️ Generated with foldermd\n\n")
	fmt.Fprintf(file, "**Configuration used:**\n")
	fmt.Fprintf(file, "- Include files: `%t`\n", config.includeFiles)
	fmt.Fprintf(file, "- Include content: `%t`\n", config.includeContent)
	fmt.Fprintf(file, "- Max depth: `%d`\n", config.maxDepth)
	fmt.Fprintf(file, "- Show hidden: `%t`\n", config.showHidden)
	fmt.Fprintf(file, "- Ignore patterns: `%s`\n", strings.Join(config.ignorePatterns, ", "))
	fmt.Fprintf(file, "\n*This README was automatically generated. Consider customizing it for your project!*\n")
}

func shouldIgnore(name string, config Config) bool {
	for _, pattern := range config.ignorePatterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == name {
			return true
		}
		// Simple glob matching for * patterns
		if strings.Contains(pattern, "*") {
			matched, _ := filepath.Match(pattern, name)
			if matched {
				return true
			}
		}
	}
	return false
}

func isTextFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read first 512 bytes to check for binary content
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return false
	}

	// Check for null bytes (common in binary files)
	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			return false
		}
	}

	return true
}

func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func getLanguageFromExtension(ext string) string {
	languages := map[string]string{
		".go":         "go",
		".py":         "python",
		".js":         "javascript",
		".ts":         "typescript",
		".jsx":        "jsx",
		".tsx":        "tsx",
		".java":       "java",
		".c":          "c",
		".cpp":        "cpp",
		".cc":         "cpp",
		".cxx":        "cpp",
		".h":          "c",
		".hpp":        "cpp",
		".hxx":        "cpp",
		".rs":         "rust",
		".php":        "php",
		".rb":         "ruby",
		".sh":         "bash",
		".bash":       "bash",
		".zsh":        "zsh",
		".fish":       "fish",
		".ps1":        "powershell",
		".html":       "html",
		".htm":        "html",
		".css":        "css",
		".scss":       "scss",
		".sass":       "sass",
		".less":       "less",
		".xml":        "xml",
		".json":       "json",
		".yaml":       "yaml",
		".yml":        "yaml",
		".toml":       "toml",
		".ini":        "ini",
		".cfg":        "ini",
		".conf":       "ini",
		".md":         "markdown",
		".txt":        "text",
		".sql":        "sql",
		".r":          "r",
		".m":          "matlab",
		".swift":      "swift",
		".kt":         "kotlin",
		".kts":        "kotlin",
		".scala":      "scala",
		".clj":        "clojure",
		".cljs":       "clojure",
		".hs":         "haskell",
		".elm":        "elm",
		".ex":         "elixir",
		".exs":        "elixir",
		".erl":        "erlang",
		".dart":       "dart",
		".lua":        "lua",
		".pl":         "perl",
		".vim":        "vim",
		".dockerfile": "dockerfile",
		".gitignore":  "gitignore",
		".env":        "bash",
		".makefile":   "makefile",
		".cmake":      "cmake",
	}

	if lang, exists := languages[ext]; exists {
		return lang
	}
	return "text"
}

func readIgnoreFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}

	return patterns, scanner.Err()
}

func createIgnoreFile() error {
	filename := ".foldermd.ignore"
	if _, err := os.Stat(filename); err == nil {
		return fmt.Errorf("file %s already exists", filename)
	}

	content := `# foldermd ignore patterns
# Lines starting with # are comments
# Use glob patterns to match files and directories

# Version control
.git
.svn
.hg

# Dependencies
node_modules
vendor
__pycache__
.venv
venv

# Build outputs
build
dist
out
target
bin
obj

# IDE and editor files
.vscode
.idea
*.swp
*.swo
*~

# OS generated files
.DS_Store
Thumbs.db
Desktop.ini

# Logs
*.log
logs

# Temporary files
tmp
temp
*.tmp
*.temp

# Archives
*.zip
*.tar.gz
*.rar
*.7z`

	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", filename, err)
	}

	fmt.Printf("✅ Created %s with common ignore patterns\n", filename)
	fmt.Printf("💡 Edit this file to customize ignore patterns for your project\n")
	return nil
}