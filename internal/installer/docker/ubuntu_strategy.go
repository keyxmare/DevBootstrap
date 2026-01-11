package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
	"github.com/keyxmare/DevBootstrap/internal/domain/result"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
	"github.com/keyxmare/DevBootstrap/internal/installer/strategy"
	"github.com/keyxmare/DevBootstrap/internal/port/primary"
	"github.com/keyxmare/DevBootstrap/internal/port/secondary"
)

// UbuntuStrategy implements Docker installation on Ubuntu/Debian.
type UbuntuStrategy struct {
	strategy.BaseStrategy
}

// NewUbuntuStrategy creates a new Ubuntu/Debian Docker installer strategy.
func NewUbuntuStrategy(deps strategy.Dependencies, platform *entity.Platform) *UbuntuStrategy {
	return &UbuntuStrategy{
		BaseStrategy: strategy.NewBaseStrategy(deps, platform),
	}
}

// CheckStatus checks if Docker is already installed.
func (s *UbuntuStrategy) CheckStatus(ctx context.Context) (valueobject.AppStatus, string, error) {
	if s.CommandExists("docker") {
		version := s.GetCommandVersion("docker")
		return valueobject.StatusInstalled, version, nil
	}
	return valueobject.StatusNotInstalled, "", nil
}

// Install installs Docker Engine on Ubuntu/Debian.
func (s *UbuntuStrategy) Install(ctx context.Context, opts primary.InstallOptions) (*result.InstallResult, error) {
	installResult := result.NewSuccess("")

	// Step 1: Remove old packages
	s.Section("Suppression des anciens paquets Docker")
	oldPackages := []string{
		"docker.io", "docker-doc", "docker-compose", "docker-compose-v2",
		"podman-docker", "containerd", "runc",
	}
	s.Run(ctx,
		append([]string{"apt-get", "remove", "-y"}, oldPackages...),
		secondary.WithSudo(),
		secondary.WithDescription("Suppression des anciens paquets"),
	)

	// Step 2: Install prerequisites
	s.Section("Installation des prerequis")
	prereqs := []string{"ca-certificates", "curl", "gnupg"}
	prereqResult := s.Run(ctx,
		append([]string{"apt-get", "install", "-y"}, prereqs...),
		secondary.WithSudo(),
		secondary.WithDescription("Installation des prerequis"),
	)
	if !prereqResult.Success {
		return result.NewFailure("Echec de l'installation des prerequis", prereqResult.Stderr), nil
	}
	s.Success("Prerequis installes")

	// Step 3: Add Docker's official GPG key
	s.Section("Configuration du depot Docker")
	s.Info("Ajout de la cle GPG Docker...")

	// Create keyrings directory
	s.Run(ctx, []string{"mkdir", "-p", "/etc/apt/keyrings"}, secondary.WithSudo())

	// Download and add GPG key
	gpgKeyURL := "https://download.docker.com/linux/ubuntu/gpg"
	gpgKeyPath := "/etc/apt/keyrings/docker.asc"

	// Download GPG key
	tmpGPG := filepath.Join(os.TempDir(), "docker.gpg")
	if err := s.Deps.HTTPClient.Download(ctx, gpgKeyURL, tmpGPG); err != nil {
		return result.NewFailure("Echec du telechargement de la cle GPG"), nil
	}

	// Move to keyrings
	s.Run(ctx, []string{"mv", tmpGPG, gpgKeyPath}, secondary.WithSudo())
	s.Run(ctx, []string{"chmod", "a+r", gpgKeyPath}, secondary.WithSudo())
	s.Success("Cle GPG ajoutee")

	// Step 4: Set up repository
	s.Info("Configuration du depot APT...")

	// Get distribution codename
	codename := s.getDistroCodename(ctx)
	arch := s.getArch(ctx)

	repoLine := fmt.Sprintf("deb [arch=%s signed-by=%s] https://download.docker.com/linux/ubuntu %s stable",
		arch, gpgKeyPath, codename)

	// Write repository file
	repoFile := "/etc/apt/sources.list.d/docker.list"
	s.Run(ctx,
		[]string{"bash", "-c", fmt.Sprintf("echo '%s' > %s", repoLine, repoFile)},
		secondary.WithSudo(),
	)
	s.Success("Depot configure")

	// Step 5: Update package index
	s.Section("Mise a jour des paquets")
	updateResult := s.Run(ctx,
		[]string{"apt-get", "update"},
		secondary.WithSudo(),
		secondary.WithDescription("Mise a jour de l'index des paquets"),
	)
	if !updateResult.Success {
		installResult.AddWarning("Echec de la mise a jour de l'index")
	}

	// Step 6: Install Docker packages
	s.Section("Installation de Docker")
	dockerPackages := []string{
		"docker-ce",
		"docker-ce-cli",
		"containerd.io",
		"docker-buildx-plugin",
	}

	// Check if compose should be installed
	installCompose := true
	if opts.DockerOptions != nil {
		installCompose = opts.DockerOptions.InstallCompose
	}
	if installCompose {
		dockerPackages = append(dockerPackages, "docker-compose-plugin")
	}

	dockerInstallResult := s.Run(ctx,
		append([]string{"apt-get", "install", "-y"}, dockerPackages...),
		secondary.WithSudo(),
		secondary.WithDescription("Installation de Docker Engine"),
		secondary.WithTimeout(10*time.Minute),
	)

	if !dockerInstallResult.Success {
		return result.NewFailure("Echec de l'installation de Docker", dockerInstallResult.Stderr), nil
	}
	s.Success("Docker installe")

	// Step 7: Add user to docker group
	addToGroup := true
	if opts.DockerOptions != nil {
		addToGroup = opts.DockerOptions.AddUserToDockerGroup
	}
	if addToGroup {
		s.Section("Configuration du groupe docker")
		username := s.Platform.Username()

		if username != "" && username != "root" {
			s.Run(ctx, []string{"usermod", "-aG", "docker", username}, secondary.WithSudo())
			s.Success(fmt.Sprintf("Utilisateur '%s' ajoute au groupe docker", username))
			installResult.AddWarning("Deconnectez-vous et reconnectez-vous pour utiliser Docker sans sudo")
		}
	}

	// Step 8: Enable and start Docker service
	startOnBoot := true
	if opts.DockerOptions != nil {
		startOnBoot = opts.DockerOptions.StartOnBoot
	}
	if startOnBoot {
		s.Section("Configuration du service")
		s.Run(ctx, []string{"systemctl", "enable", "docker"}, secondary.WithSudo())
		s.Run(ctx, []string{"systemctl", "start", "docker"}, secondary.WithSudo())
		s.Success("Service Docker demarre et active au boot")
	}

	// Verify installation
	if s.Verify(ctx) {
		version := s.GetCommandVersion("docker")
		return result.NewSuccess("Docker installe avec succes").WithVersion(version), nil
	}

	installResult.AddWarning("L'installation semble terminee mais la verification a echoue")
	return installResult, nil
}

// getDistroCodename returns the distribution codename.
func (s *UbuntuStrategy) getDistroCodename(ctx context.Context) string {
	res := s.Run(ctx, []string{"lsb_release", "-cs"})
	if res.Success {
		return strings.TrimSpace(res.Stdout)
	}

	// Fallback: read from /etc/os-release
	res = s.Run(ctx, []string{"bash", "-c", ". /etc/os-release && echo $VERSION_CODENAME"})
	if res.Success {
		return strings.TrimSpace(res.Stdout)
	}

	return "jammy" // Default to Ubuntu 22.04
}

// getArch returns the architecture string for APT.
func (s *UbuntuStrategy) getArch(ctx context.Context) string {
	res := s.Run(ctx, []string{"dpkg", "--print-architecture"})
	if res.Success {
		return strings.TrimSpace(res.Stdout)
	}
	return "amd64"
}

// Verify verifies the Docker installation.
func (s *UbuntuStrategy) Verify(ctx context.Context) bool {
	res := s.Run(ctx, []string{"docker", "--version"})
	return res.Success
}

// Uninstall removes Docker from Ubuntu/Debian.
func (s *UbuntuStrategy) Uninstall(ctx context.Context, opts primary.UninstallOptions) (*result.UninstallResult, error) {
	s.Section("Desinstallation de Docker")

	// Stop Docker service
	s.Info("Arret du service Docker...")
	s.Run(ctx, []string{"systemctl", "stop", "docker"}, secondary.WithSudo())
	s.Run(ctx, []string{"systemctl", "disable", "docker"}, secondary.WithSudo())

	// Remove Docker packages
	s.Info("Suppression des paquets Docker...")
	packages := []string{
		"docker-ce",
		"docker-ce-cli",
		"containerd.io",
		"docker-buildx-plugin",
		"docker-compose-plugin",
	}
	s.Run(ctx,
		append([]string{"apt-get", "purge", "-y"}, packages...),
		secondary.WithSudo(),
	)
	s.Success("Paquets Docker supprimes")

	// Remove Docker data
	if opts.RemoveData {
		s.Info("Suppression des donnees Docker...")
		dataDirectories := []string{
			"/var/lib/docker",
			"/var/lib/containerd",
			"/etc/docker",
			filepath.Join(s.Platform.HomeDir(), ".docker"),
		}

		for _, dir := range dataDirectories {
			s.Run(ctx, []string{"rm", "-rf", dir}, secondary.WithSudo())
		}
		s.Success("Donnees Docker supprimees")
	}

	// Remove repository
	s.Run(ctx, []string{"rm", "-f", "/etc/apt/sources.list.d/docker.list"}, secondary.WithSudo())
	s.Run(ctx, []string{"rm", "-f", "/etc/apt/keyrings/docker.asc"}, secondary.WithSudo())

	return result.NewUninstallSuccess("Docker desinstalle avec succes"), nil
}
