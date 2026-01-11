package docker

import (
	"os"
	"path/filepath"
	"time"

	"github.com/keyxmare/DevBootstrap/internal/installers"
	"github.com/keyxmare/DevBootstrap/internal/runner"
	"github.com/keyxmare/DevBootstrap/internal/system"
)

// Docker Desktop download URLs
const (
	dockerDMGURLARM64 = "https://desktop.docker.com/mac/main/arm64/Docker.dmg"
	dockerDMGURLAMD64 = "https://desktop.docker.com/mac/main/amd64/Docker.dmg"
)

// macOSInstaller handles Docker installation on macOS.
type macOSInstaller struct {
	*installers.BaseInstaller
}

// newMacOSInstaller creates a new macOS Docker installer.
func newMacOSInstaller(base *installers.BaseInstaller) *macOSInstaller {
	return &macOSInstaller{BaseInstaller: base}
}

// getDMGURL returns the appropriate DMG URL for the current architecture.
func (m *macOSInstaller) getDMGURL() string {
	if m.SystemInfo.Arch == system.ArchARM64 {
		return dockerDMGURLARM64
	}
	return dockerDMGURLAMD64
}

// CheckExisting checks if Docker is already installed.
func (m *macOSInstaller) CheckExisting() (installers.AppStatus, string) {
	// Check if docker command exists and daemon is accessible
	if m.Runner.CommandExists("docker") {
		result := m.Runner.Run([]string{"docker", "info"}, runner.WithTimeout(10*time.Second))
		if result.Success {
			version := m.Runner.GetCommandVersion("docker", "--version")
			return installers.StatusInstalled, version
		}
	}

	// Check if Docker Desktop app exists
	dockerAppPaths := []string{
		"/Applications/Docker.app",
		filepath.Join(m.SystemInfo.HomeDir, "Applications/Docker.app"),
	}

	for _, path := range dockerAppPaths {
		if _, err := os.Stat(path); err == nil {
			return installers.StatusInstalled, "(Docker Desktop installe)"
		}
	}

	return installers.StatusNotInstalled, ""
}

// Install installs Docker Desktop by downloading and installing the official DMG.
func (m *macOSInstaller) Install(opts *installers.InstallOptions) *installers.InstallResult {
	dmgURL := m.getDMGURL()
	archName := "Intel"
	if m.SystemInfo.Arch == system.ArchARM64 {
		archName = "Apple Silicon"
	}

	m.CLI.PrintInfo("Telechargement de Docker Desktop pour " + archName + "...")
	m.CLI.PrintInfo("URL: " + dmgURL)

	// Create temporary directory for download
	tmpDir, err := os.MkdirTemp("", "docker-install-*")
	if err != nil {
		return installers.NewFailureResult("Impossible de creer le repertoire temporaire", err.Error())
	}
	defer os.RemoveAll(tmpDir)

	dmgPath := filepath.Join(tmpDir, "Docker.dmg")

	// Download the DMG
	if !m.Runner.DownloadFile(dmgURL, dmgPath, runner.WithDescription("Telechargement de Docker Desktop")) {
		return installers.NewFailureResult(
			"Echec du telechargement de Docker Desktop",
			"Vous pouvez telecharger manuellement depuis: https://www.docker.com/products/docker-desktop/",
		)
	}
	m.CLI.PrintSuccess("Telechargement termine")

	// Mount the DMG
	m.CLI.PrintInfo("Montage du DMG...")
	mountPoint := "/Volumes/Docker"

	// Unmount if already mounted
	if _, err := os.Stat(mountPoint); err == nil {
		m.Runner.Run([]string{"hdiutil", "detach", mountPoint, "-quiet"})
	}

	result := m.Runner.Run(
		[]string{"hdiutil", "attach", dmgPath, "-nobrowse", "-quiet"},
		runner.WithDescription("Montage du DMG"),
		runner.WithTimeout(60*time.Second),
	)

	if !result.Success {
		return installers.NewFailureResult("Echec du montage du DMG", result.Stderr)
	}

	// Ensure we unmount at the end
	defer func() {
		m.CLI.PrintInfo("Demontage du DMG...")
		m.Runner.Run([]string{"hdiutil", "detach", mountPoint, "-quiet"}, runner.WithTimeout(30*time.Second))
	}()

	// Copy Docker.app to /Applications
	m.CLI.PrintInfo("Installation de Docker Desktop dans /Applications...")
	dockerAppSrc := "/Volumes/Docker/Docker.app"

	if _, err := os.Stat(dockerAppSrc); os.IsNotExist(err) {
		return installers.NewFailureResult("Docker.app non trouve dans le DMG")
	}

	// Remove existing installation if present
	dockerAppDest := "/Applications/Docker.app"
	if _, err := os.Stat(dockerAppDest); err == nil {
		m.CLI.PrintInfo("Suppression de l'ancienne version...")
		if err := m.Runner.RemoveAll(dockerAppDest); err != nil {
			return installers.NewFailureResult("Impossible de supprimer l'ancienne version", err.Error())
		}
	}

	// Copy the app
	copyResult := m.Runner.Run(
		[]string{"cp", "-R", dockerAppSrc, dockerAppDest},
		runner.WithDescription("Copie de Docker.app"),
		runner.WithTimeout(120*time.Second),
	)

	if !copyResult.Success {
		return installers.NewFailureResult("Echec de la copie de Docker.app", copyResult.Stderr)
	}

	m.CLI.PrintSuccess("Docker Desktop installe dans /Applications")

	// Start Docker Desktop
	m.startDocker()

	// Show additional info
	m.showFinalInfo()

	return installers.NewSuccessResult("Docker Desktop installe avec succes")
}

// startDocker starts Docker Desktop.
func (m *macOSInstaller) startDocker() {
	// Check if Docker is already running
	result := m.Runner.Run([]string{"docker", "info"}, runner.WithTimeout(10*time.Second))
	if result.Success {
		m.CLI.PrintSuccess("Docker est deja en cours d'execution")
		return
	}

	// Start Docker Desktop
	m.CLI.PrintInfo("Demarrage de Docker Desktop...")
	startResult := m.Runner.Run(
		[]string{"open", "-a", "Docker"},
		runner.WithDescription("Demarrage de Docker Desktop"),
	)

	if !startResult.Success {
		m.CLI.PrintWarning("Impossible de demarrer Docker Desktop automatiquement")
		m.CLI.PrintInfo("Veuillez demarrer Docker Desktop manuellement depuis Applications")
		return
	}

	// Wait for Docker to be ready
	m.CLI.PrintInfo("Attente du demarrage de Docker...")
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(2 * time.Second)
		result := m.Runner.Run([]string{"docker", "info"}, runner.WithTimeout(5*time.Second))
		if result.Success {
			m.CLI.PrintSuccess("Docker est pret!")
			return
		}
		m.CLI.PrintProgress("Attente... (" + string(rune('0'+i+1)) + "/" + string(rune('0'+maxAttempts)) + ")")
	}

	m.CLI.PrintWarning("Docker prend du temps a demarrer")
	m.CLI.PrintInfo("Veuillez attendre que l'icone Docker dans la barre de menu indique 'Running'")
}

// showFinalInfo shows additional information after installation.
func (m *macOSInstaller) showFinalInfo() {
	m.CLI.PrintSection("Informations supplementaires")
	m.CLI.PrintInfo("Docker Desktop inclut:")
	m.CLI.Println("  - Docker Engine")
	m.CLI.Println("  - Docker CLI")
	m.CLI.Println("  - Docker Compose")
	m.CLI.Println("  - Docker BuildKit")
	m.CLI.Println("  - Kubernetes (optionnel)")
	m.CLI.Println("")
	m.CLI.PrintInfo("Pour activer Kubernetes:")
	m.CLI.Println("  Docker Desktop > Preferences > Kubernetes > Enable Kubernetes")
}

// Verify verifies the Docker installation.
func (m *macOSInstaller) Verify() bool {
	result := m.Runner.Run([]string{"docker", "--version"}, runner.WithTimeout(10*time.Second))
	return result.Success
}

// Uninstall removes Docker Desktop.
func (m *macOSInstaller) Uninstall(opts *installers.UninstallOptions) *installers.UninstallResult {
	m.CLI.PrintSection("Desinstallation de Docker Desktop")

	// Quit Docker Desktop
	m.CLI.PrintInfo("Fermeture de Docker Desktop...")
	m.Runner.Run([]string{"osascript", "-e", `quit app "Docker"`})
	time.Sleep(2 * time.Second)

	// Remove Docker.app
	dockerApp := "/Applications/Docker.app"
	if _, err := os.Stat(dockerApp); err == nil {
		m.CLI.PrintInfo("Suppression de Docker.app...")
		if err := m.Runner.RemoveAll(dockerApp); err != nil {
			return installers.NewUninstallFailureResult("Impossible de supprimer Docker.app", err.Error())
		}
		m.CLI.PrintSuccess("Docker.app supprime")
	}

	// Remove Docker data directories
	if opts.RemoveData {
		dataDirectories := []string{
			filepath.Join(m.SystemInfo.HomeDir, ".docker"),
			filepath.Join(m.SystemInfo.HomeDir, "Library/Group Containers/group.com.docker"),
			filepath.Join(m.SystemInfo.HomeDir, "Library/Containers/com.docker.docker"),
			filepath.Join(m.SystemInfo.HomeDir, "Library/Application Support/Docker Desktop"),
			filepath.Join(m.SystemInfo.HomeDir, "Library/Preferences/com.docker.docker.plist"),
			filepath.Join(m.SystemInfo.HomeDir, "Library/Saved Application State/com.electron.docker-frontend.savedState"),
			filepath.Join(m.SystemInfo.HomeDir, "Library/Logs/Docker Desktop"),
		}

		m.CLI.PrintInfo("Suppression des donnees Docker...")
		for _, dir := range dataDirectories {
			if _, err := os.Stat(dir); err == nil {
				m.Runner.RemoveAll(dir)
			}
		}
		m.CLI.PrintSuccess("Donnees Docker supprimees")
	}

	return installers.NewUninstallSuccessResult("Docker Desktop desinstalle avec succes")
}
