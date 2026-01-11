// Package docker provides Docker installation strategies.
package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
	"github.com/keyxmare/DevBootstrap/internal/domain/result"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
	"github.com/keyxmare/DevBootstrap/internal/installer/strategy"
	"github.com/keyxmare/DevBootstrap/internal/port/primary"
	"github.com/keyxmare/DevBootstrap/internal/port/secondary"
)

// Docker Desktop download URLs
const (
	dockerDMGURLARM64 = "https://desktop.docker.com/mac/main/arm64/Docker.dmg"
	dockerDMGURLAMD64 = "https://desktop.docker.com/mac/main/amd64/Docker.dmg"
)

// MacOSStrategy implements Docker installation on macOS.
type MacOSStrategy struct {
	strategy.BaseStrategy
}

// NewMacOSStrategy creates a new macOS Docker installer strategy.
func NewMacOSStrategy(deps strategy.Dependencies, platform *entity.Platform) *MacOSStrategy {
	return &MacOSStrategy{
		BaseStrategy: strategy.NewBaseStrategy(deps, platform),
	}
}

// getDMGURL returns the appropriate DMG URL for the current architecture.
func (s *MacOSStrategy) getDMGURL() string {
	if s.Platform.Arch() == valueobject.ArchARM64 {
		return dockerDMGURLARM64
	}
	return dockerDMGURLAMD64
}

// CheckStatus checks if Docker is already installed.
func (s *MacOSStrategy) CheckStatus(ctx context.Context) (valueobject.AppStatus, string, error) {
	// Check if docker command exists and daemon is accessible
	if s.CommandExists("docker") {
		res := s.Run(ctx, []string{"docker", "info"}, secondary.WithTimeout(10*time.Second))
		if res.Success {
			version := s.GetCommandVersion("docker")
			return valueobject.StatusInstalled, version, nil
		}
	}

	// Check if Docker Desktop app exists
	dockerAppPaths := []string{
		"/Applications/Docker.app",
		filepath.Join(s.Platform.HomeDir(), "Applications/Docker.app"),
	}

	for _, path := range dockerAppPaths {
		if s.Deps.FileSystem.Exists(path) {
			return valueobject.StatusInstalled, "(Docker Desktop installe)", nil
		}
	}

	return valueobject.StatusNotInstalled, "", nil
}

// Install installs Docker Desktop by downloading and installing the official DMG.
func (s *MacOSStrategy) Install(ctx context.Context, opts primary.InstallOptions) (*result.InstallResult, error) {
	dmgURL := s.getDMGURL()
	archName := "Intel"
	if s.Platform.Arch() == valueobject.ArchARM64 {
		archName = "Apple Silicon"
	}

	s.Info("Telechargement de Docker Desktop pour " + archName + "...")
	s.Info("URL: " + dmgURL)

	// Create temporary directory for download
	tmpDir := filepath.Join(os.TempDir(), "docker-install")
	if err := s.Deps.FileSystem.MkdirAll(tmpDir, 0755); err != nil {
		return result.NewFailure("Impossible de creer le repertoire temporaire", err.Error()), nil
	}
	defer s.Deps.FileSystem.RemoveAll(tmpDir)

	dmgPath := filepath.Join(tmpDir, "Docker.dmg")

	// Download the DMG
	if err := s.Deps.HTTPClient.Download(ctx, dmgURL, dmgPath); err != nil {
		return result.NewFailure(
			"Echec du telechargement de Docker Desktop",
			"Vous pouvez telecharger manuellement depuis: https://www.docker.com/products/docker-desktop/",
		), nil
	}
	s.Success("Telechargement termine")

	// Mount the DMG
	s.Info("Montage du DMG...")
	mountPoint := "/Volumes/Docker"

	// Unmount if already mounted
	if s.Deps.FileSystem.Exists(mountPoint) {
		s.Run(ctx, []string{"hdiutil", "detach", mountPoint, "-quiet"})
	}

	res := s.Run(ctx, []string{"hdiutil", "attach", dmgPath, "-nobrowse", "-quiet"},
		secondary.WithDescription("Montage du DMG"),
		secondary.WithTimeout(60*time.Second),
	)

	if !res.Success {
		return result.NewFailure("Echec du montage du DMG", res.Stderr), nil
	}

	// Ensure we unmount at the end
	defer func() {
		s.Info("Demontage du DMG...")
		s.Run(ctx, []string{"hdiutil", "detach", mountPoint, "-quiet"},
			secondary.WithTimeout(30*time.Second))
	}()

	// Copy Docker.app to /Applications
	s.Info("Installation de Docker Desktop dans /Applications...")
	dockerAppSrc := "/Volumes/Docker/Docker.app"

	if !s.Deps.FileSystem.Exists(dockerAppSrc) {
		return result.NewFailure("Docker.app non trouve dans le DMG"), nil
	}

	// Remove existing installation if present
	dockerAppDest := "/Applications/Docker.app"
	if s.Deps.FileSystem.Exists(dockerAppDest) {
		s.Info("Suppression de l'ancienne version...")
		if err := s.Deps.FileSystem.RemoveAll(dockerAppDest); err != nil {
			return result.NewFailure("Impossible de supprimer l'ancienne version", err.Error()), nil
		}
	}

	// Copy the app
	copyResult := s.Run(ctx, []string{"cp", "-R", dockerAppSrc, dockerAppDest},
		secondary.WithDescription("Copie de Docker.app"),
		secondary.WithTimeout(120*time.Second),
	)

	if !copyResult.Success {
		return result.NewFailure("Echec de la copie de Docker.app", copyResult.Stderr), nil
	}

	s.Success("Docker Desktop installe dans /Applications")

	// Start Docker Desktop
	s.startDocker(ctx)

	// Show additional info
	s.showFinalInfo()

	return result.NewSuccess("Docker Desktop installe avec succes"), nil
}

// startDocker starts Docker Desktop.
func (s *MacOSStrategy) startDocker(ctx context.Context) {
	// Check if Docker is already running
	res := s.Run(ctx, []string{"docker", "info"}, secondary.WithTimeout(10*time.Second))
	if res.Success {
		s.Success("Docker est deja en cours d'execution")
		return
	}

	// Start Docker Desktop
	s.Info("Demarrage de Docker Desktop...")
	startResult := s.Run(ctx, []string{"open", "-a", "Docker"},
		secondary.WithDescription("Demarrage de Docker Desktop"),
	)

	if !startResult.Success {
		s.Warning("Impossible de demarrer Docker Desktop automatiquement")
		s.Info("Veuillez demarrer Docker Desktop manuellement depuis Applications")
		return
	}

	// Wait for Docker to be ready
	s.Info("Attente du demarrage de Docker...")
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(2 * time.Second)
		res := s.Run(ctx, []string{"docker", "info"}, secondary.WithTimeout(5*time.Second))
		if res.Success {
			s.Success("Docker est pret!")
			return
		}
		if s.Deps.Reporter != nil {
			s.Deps.Reporter.Progress(fmt.Sprintf("Attente... (%d/%d)", i+1, maxAttempts))
		}
	}

	s.Warning("Docker prend du temps a demarrer")
	s.Info("Veuillez attendre que l'icone Docker dans la barre de menu indique 'Running'")
}

// showFinalInfo shows additional information after installation.
func (s *MacOSStrategy) showFinalInfo() {
	s.Section("Informations supplementaires")
	s.Info("Docker Desktop inclut:")
	s.Info("  - Docker Engine")
	s.Info("  - Docker CLI")
	s.Info("  - Docker Compose")
	s.Info("  - Docker BuildKit")
	s.Info("  - Kubernetes (optionnel)")
	s.Info("")
	s.Info("Pour activer Kubernetes:")
	s.Info("  Docker Desktop > Preferences > Kubernetes > Enable Kubernetes")
}

// Verify verifies the Docker installation.
func (s *MacOSStrategy) Verify(ctx context.Context) bool {
	res := s.Run(ctx, []string{"docker", "--version"}, secondary.WithTimeout(10*time.Second))
	return res.Success
}

// Uninstall removes Docker Desktop.
func (s *MacOSStrategy) Uninstall(ctx context.Context, opts primary.UninstallOptions) (*result.UninstallResult, error) {
	s.Section("Desinstallation de Docker Desktop")

	// Quit Docker Desktop
	s.Info("Fermeture de Docker Desktop...")
	s.Run(ctx, []string{"osascript", "-e", `quit app "Docker"`})
	time.Sleep(2 * time.Second)

	// Remove Docker.app
	dockerApp := "/Applications/Docker.app"
	if s.Deps.FileSystem.Exists(dockerApp) {
		s.Info("Suppression de Docker.app...")
		if err := s.Deps.FileSystem.RemoveAll(dockerApp); err != nil {
			return result.NewUninstallFailure("Impossible de supprimer Docker.app", err.Error()), nil
		}
		s.Success("Docker.app supprime")
	}

	// Remove Docker data directories
	if opts.RemoveData {
		dataDirectories := []string{
			filepath.Join(s.Platform.HomeDir(), ".docker"),
			filepath.Join(s.Platform.HomeDir(), "Library/Group Containers/group.com.docker"),
			filepath.Join(s.Platform.HomeDir(), "Library/Containers/com.docker.docker"),
			filepath.Join(s.Platform.HomeDir(), "Library/Application Support/Docker Desktop"),
			filepath.Join(s.Platform.HomeDir(), "Library/Preferences/com.docker.docker.plist"),
			filepath.Join(s.Platform.HomeDir(), "Library/Saved Application State/com.electron.docker-frontend.savedState"),
			filepath.Join(s.Platform.HomeDir(), "Library/Logs/Docker Desktop"),
		}

		s.Info("Suppression des donnees Docker...")
		for _, dir := range dataDirectories {
			if s.Deps.FileSystem.Exists(dir) {
				s.Deps.FileSystem.RemoveAll(dir)
			}
		}
		s.Success("Donnees Docker supprimees")
	}

	return result.NewUninstallSuccess("Docker Desktop desinstalle avec succes"), nil
}
