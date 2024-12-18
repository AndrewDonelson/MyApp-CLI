package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

/**
* Build production Binary
* go build -ldflags "-X main.version=$(date +%Y.%m.%d.%H%M%S)" -o myapp-cli.exe
*
* Example:
* go build -ldflags "-X main.version=2024.12.18.143000" -o myapp-cli.exe
*
* Note: Version format is YYYY.MM.DD.HHMMSS (24hr time)
 */

// Version will be set at build time
var version string

const (
	// Company Information
	CompanyName = "Nlaak Studios"
	WebsiteURL  = "https://nlaak.com"

	// Directory Settings
	projectsDir        = "C:\\Users\\andre\\NextJS-Projects"
	webappsDir         = "webapps"
	repoURL            = "https://github.com/AndrewDonelson/my-app"
	defaultProjectName = "my-new-app"
)

type CommandRunner struct {
	isWindows bool
}

func displayHeader() {
	if version == "" {
		version = "dev.build"
	}
	fmt.Printf("\n%s WebApp Utility v%s\n", CompanyName, version)
	fmt.Printf("----------------------------------------\n")
	fmt.Printf("Website: %s\n\n", WebsiteURL)
}

func NewCommandRunner() *CommandRunner {
	return &CommandRunner{
		isWindows: runtime.GOOS == "windows",
	}
}

func (cr *CommandRunner) execCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (cr *CommandRunner) getNPMCommand() string {
	if cr.isWindows {
		return "npm.cmd"
	}
	return "npm"
}

func (cr *CommandRunner) getGHCommand() string {
	if cr.isWindows {
		return "gh.exe"
	}
	return "gh"
}

func checkPrerequisites(cr *CommandRunner) error {
	// Check for git
	if err := cr.execCommand("git", "--version"); err != nil {
		return fmt.Errorf("git is not installed or not in PATH")
	}

	// Check for GitHub CLI
	if err := cr.execCommand(cr.getGHCommand(), "--version"); err != nil {
		return fmt.Errorf("GitHub CLI (gh) is not installed or not in PATH")
	}

	// Check for npm
	if err := cr.execCommand(cr.getNPMCommand(), "--version"); err != nil {
		return fmt.Errorf("npm is not installed or not in PATH")
	}

	return nil
}

func ensureWebappsDir() error {
	fullWebappsPath := filepath.Join(projectsDir, webappsDir)
	if err := os.MkdirAll(fullWebappsPath, 0755); err != nil {
		return fmt.Errorf("failed to create webapps directory: %v", err)
	}
	return nil
}

func isProjectExists(name string) bool {
	projectPath := filepath.Join(projectsDir, webappsDir, name)
	_, err := os.Stat(projectPath)
	return !os.IsNotExist(err)
}

func promptProjectName() string {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("Enter the name for your new WebApp (default: %s): ", defaultProjectName)
		name, err := reader.ReadString('\n')
		if err != nil {
			return defaultProjectName
		}

		name = strings.TrimSpace(name)
		if name == "" {
			name = defaultProjectName
		}

		if isProjectExists(name) {
			fmt.Printf("\nA project with the name '%s' already exists. Please choose a different name.\n\n", name)
			continue
		}

		// Basic name validation
		if strings.ContainsAny(name, "\\/:*?\"<>|") {
			fmt.Println("\nProject name contains invalid characters. Please use only letters, numbers, dashes, and underscores.\n")
			continue
		}

		return name
	}
}

func createProject(cr *CommandRunner, projectName string) error {
	if err := ensureWebappsDir(); err != nil {
		return err
	}

	// Get absolute path for the project within webapps directory
	fullWebappsPath := filepath.Join(projectsDir, webappsDir)
	projectPath := filepath.Join(fullWebappsPath, projectName)

	// Already checked in promptProjectName, but double-check
	if isProjectExists(projectName) {
		return fmt.Errorf("directory %s already exists", projectPath)
	}

	// First change to the webapps directory
	if err := os.Chdir(fullWebappsPath); err != nil {
		return fmt.Errorf("failed to change to webapps directory: %v", err)
	}

	fmt.Printf("Creating new WebApp: %s\n", projectName)

	// Clone the repository
	if err := cr.execCommand(cr.getGHCommand(), "repo", "clone", repoURL, projectName); err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
	}

	// Change to project directory
	if err := os.Chdir(projectPath); err != nil {
		return fmt.Errorf("failed to change to project directory: %v", err)
	}

	// Remove existing git directory
	if err := os.RemoveAll(".git"); err != nil {
		return fmt.Errorf("failed to remove existing .git directory: %v", err)
	}

	// Initialize new git repository
	fmt.Println("Initializing new git repository...")
	if err := cr.execCommand("git", "init"); err != nil {
		return fmt.Errorf("failed to initialize git repository: %v", err)
	}

	// Set up initial commit
	if err := cr.execCommand("git", "add", "."); err != nil {
		return fmt.Errorf("failed to stage files: %v", err)
	}

	if err := cr.execCommand("git", "commit", "-m", "Initial commit"); err != nil {
		return fmt.Errorf("failed to create initial commit: %v", err)
	}

	// Clean install node_modules
	fmt.Println("Cleaning existing node_modules...")
	if err := os.RemoveAll("node_modules"); err != nil {
		return fmt.Errorf("failed to remove node_modules: %v", err)
	}

	// Clean package-lock.json
	if err := os.Remove("package-lock.json"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove package-lock.json: %v", err)
	}

	// Install dependencies
	fmt.Println("Installing dependencies...")
	if err := cr.execCommand(cr.getNPMCommand(), "install"); err != nil {
		return fmt.Errorf("failed to install dependencies: %v", err)
	}

	// Modify package.json to remove predev script temporarily
	fmt.Println("Adjusting package.json for initial setup...")
	packageJSON, err := os.ReadFile("package.json")
	if err != nil {
		return fmt.Errorf("failed to read package.json: %v", err)
	}

	// Create backup of original package.json
	if err := os.WriteFile("package.json.backup", packageJSON, 0644); err != nil {
		return fmt.Errorf("failed to create package.json backup: %v", err)
	}

	// Replace predev script with simpler version
	packageJSONStr := string(packageJSON)
	packageJSONStr = strings.Replace(packageJSONStr,
		`"predev": "npx convex dev --until-success && node setup.mjs --once && npx convex dashboard"`,
		`"predev": "echo Skipping predev script for initial setup"`,
		1)

	if err := os.WriteFile("package.json", []byte(packageJSONStr), 0644); err != nil {
		return fmt.Errorf("failed to write modified package.json: %v", err)
	}

	fmt.Println("\nProject setup completed successfully!")
	return nil
}

func main() {
	// Display header
	displayHeader()

	cr := NewCommandRunner()

	// Check prerequisites
	if err := checkPrerequisites(cr); err != nil {
		log.Fatalf("Prerequisite check failed: %v", err)
	}

	// Ensure webapps directory exists
	if err := ensureWebappsDir(); err != nil {
		log.Fatalf("Failed to setup projects directory: %v", err)
	}

	// Get project name with validation
	projectName := promptProjectName()

	// Create the project
	if err := createProject(cr, projectName); err != nil {
		log.Fatalf("Failed to create project: %v", err)
	}

	// Instead of trying to change the directory, provide the command
	projectPath := filepath.Join(projectsDir, webappsDir, projectName)
	fmt.Println("\nTo continue setup, run these commands:")
	fmt.Printf("\n   cd %s\n", projectPath)
	fmt.Println("   mv package.json.backup package.json")
	fmt.Println("   npm run dev")
}
