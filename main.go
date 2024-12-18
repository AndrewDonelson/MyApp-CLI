package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	repoURL          = "https://github.com/AndrewDonelson/my-app"
	defaultProjectName = "my-new-app"
)

type CommandRunner struct {
	isWindows bool
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

func checkSetupScript(projectPath string) bool {
	envPath := filepath.Join(projectPath, ".env.local")
	file, err := os.Open(envPath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "SETUP_SCRIPT_RAN=1") {
			return true
		}
	}
	return false
}

func promptProjectName() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter the name for your new WebApp (default: %s): ", defaultProjectName)
	name, err := reader.ReadString('\n')
	if err != nil {
		return defaultProjectName
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return defaultProjectName
	}
	return name
}

func createProject(cr *CommandRunner, projectName string) error {
	// Get absolute path for the project
	projectPath, err := filepath.Abs(projectName)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Check if project directory already exists
	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		return fmt.Errorf("directory %s already exists", projectPath)
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

	// Install dependencies
	fmt.Println("Installing dependencies...")
	if err := cr.execCommand(cr.getNPMCommand(), "install"); err != nil {
		return fmt.Errorf("failed to install dependencies: %v", err)
	}

	return nil
}

func startDevServer(cr *CommandRunner) error {
	return cr.execCommand(cr.getNPMCommand(), "run", "dev")
}

func main() {
	// Parse command line flags
	skipSetup := flag.Bool("skip-setup", false, "Skip setup and just start the dev server")
	flag.Parse()

	cr := NewCommandRunner()

	// Check prerequisites
	if err := checkPrerequisites(cr); err != nil {
		log.Fatalf("Prerequisite check failed: %v", err)
	}

	// If skip-setup is true and we're in a project directory, just start the dev server
	if *skipSetup {
		if err := startDevServer(cr); err != nil {
			log.Fatalf("Failed to start development server: %v", err)
		}
		return
	}

	// Get project name
	projectName := promptProjectName()

	// Create the project
	if err := createProject(cr, projectName); err != nil {
		log.Fatalf("Failed to create project: %v", err)
	}

	// Check if setup script has already run
	if checkSetupScript(projectName) {
		fmt.Println("Setup script has already run. Starting development server...")
	}

	// Start the development server
	if err := startDevServer(cr); err != nil {
		log.Fatalf("Failed to start development server: %v", err)
	}
}