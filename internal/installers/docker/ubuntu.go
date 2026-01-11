package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/keyxmare/DevBootstrap/internal/installers"
	"github.com/keyxmare/DevBootstrap/internal/runner"
)

// ubuntuInstaller handles Docker installation on Ubuntu/Debian.
type ubuntuInstaller struct {
	*installers.BaseInstaller
}

// newUbuntuInstaller creates a new Ubuntu/Debian Docker installer.
func newUbuntuInstaller(base *installers.BaseInstaller) *ubuntuInstaller {
	return &ubuntuInstaller{BaseInstaller: base}
}

// CheckExisting checks if Docker is already installed.
func (u *ubuntuInstaller) CheckExisting() (installers.AppStatus, string) {
	if u.Runner.CommandExists("docker") {
		version := u.Runner.GetCommandVersion("docker", "--version")
		return installers.StatusInstalled, version
	}
	return installers.StatusNotInstalled, ""
}

// Install installs Docker Engine on Ubuntu/Debian.
func (u *ubuntuInstaller) Install(opts *installers.InstallOptions) *installers.InstallResult {
	result := &installers.InstallResult{Success: true}

	// Step 1: Remove old packages
	u.CLI.PrintSection("Suppression des anciens paquets Docker")
	oldPackages := []string{
		"docker.io", "docker-doc", "docker-compose", "docker-compose-v2",
		"podman-docker", "containerd", "runc",
	}
	u.Runner.Run(
		append([]string{"apt-get", "remove", "-y"}, oldPackages...),
		runner.WithSudo(),
		runner.WithDescription("Suppression des anciens paquets"),
	)

	// Step 2: Install prerequisites
	u.CLI.PrintSection("Installation des prerequis")
	prereqs := []string{"ca-certificates", "curl", "gnupg"}
	prereqResult := u.Runner.Run(
		append([]string{"apt-get", "install", "-y"}, prereqs...),
		runner.WithSudo(),
		runner.WithDescription("Installation des prerequis"),
	)
	if !prereqResult.Success {
		return installers.NewFailureResult("Echec de l'installation des prerequis", prereqResult.Stderr)
	}
	u.CLI.PrintSuccess("Prerequis installes")

	// Step 3: Add Docker's official GPG key
	u.CLI.PrintSection("Configuration du depot Docker")
	u.CLI.PrintInfo("Ajout de la cle GPG Docker...")

	// Create keyrings directory
	u.Runner.Run([]string{"mkdir", "-p", "/etc/apt/keyrings"}, runner.WithSudo())

	// Download and add GPG key
	gpgKeyURL := "https://download.docker.com/linux/ubuntu/gpg"
	gpgKeyPath := "/etc/apt/keyrings/docker.asc"

	// Download GPG key
	tmpGPG := filepath.Join(os.TempDir(), "docker.gpg")
	if !u.Runner.DownloadFile(gpgKeyURL, tmpGPG, runner.WithDescription("Telechargement de la cle GPG")) {
		return installers.NewFailureResult("Echec du telechargement de la cle GPG")
	}

	// Move to keyrings
	u.Runner.Run([]string{"mv", tmpGPG, gpgKeyPath}, runner.WithSudo())
	u.Runner.Run([]string{"chmod", "a+r", gpgKeyPath}, runner.WithSudo())
	u.CLI.PrintSuccess("Cle GPG ajoutee")

	// Step 4: Set up repository
	u.CLI.PrintInfo("Configuration du depot APT...")

	// Get distribution codename
	codename := u.getDistroCodename()
	arch := u.getArch()

	repoLine := fmt.Sprintf("deb [arch=%s signed-by=%s] https://download.docker.com/linux/ubuntu %s stable",
		arch, gpgKeyPath, codename)

	// Write repository file
	repoFile := "/etc/apt/sources.list.d/docker.list"
	u.Runner.Run(
		[]string{"bash", "-c", fmt.Sprintf("echo '%s' > %s", repoLine, repoFile)},
		runner.WithSudo(),
	)
	u.CLI.PrintSuccess("Depot configure")

	// Step 5: Update package index
	u.CLI.PrintSection("Mise a jour des paquets")
	updateResult := u.Runner.Run(
		[]string{"apt-get", "update"},
		runner.WithSudo(),
		runner.WithDescription("Mise a jour de l'index des paquets"),
	)
	if !updateResult.Success {
		result.AddWarning("Echec de la mise a jour de l'index")
	}

	// Step 6: Install Docker packages
	u.CLI.PrintSection("Installation de Docker")
	dockerPackages := []string{
		"docker-ce",
		"docker-ce-cli",
		"containerd.io",
		"docker-buildx-plugin",
	}

	if opts.InstallCompose {
		dockerPackages = append(dockerPackages, "docker-compose-plugin")
	}

	installResult := u.Runner.Run(
		append([]string{"apt-get", "install", "-y"}, dockerPackages...),
		runner.WithSudo(),
		runner.WithDescription("Installation de Docker Engine"),
		runner.WithTimeout(10*time.Minute),
	)

	if !installResult.Success {
		return installers.NewFailureResult("Echec de l'installation de Docker", installResult.Stderr)
	}
	u.CLI.PrintSuccess("Docker installe")

	// Step 7: Add user to docker group
	if opts.AddUserToDockerGroup {
		u.CLI.PrintSection("Configuration du groupe docker")
		username := os.Getenv("USER")
		if username == "" {
			username = os.Getenv("LOGNAME")
		}

		if username != "" && username != "root" {
			u.Runner.Run([]string{"usermod", "-aG", "docker", username}, runner.WithSudo())
			u.CLI.PrintSuccess(fmt.Sprintf("Utilisateur '%s' ajoute au groupe docker", username))
			result.AddWarning("Deconnectez-vous et reconnectez-vous pour utiliser Docker sans sudo")
		}
	}

	// Step 8: Enable and start Docker service
	if opts.StartOnBoot {
		u.CLI.PrintSection("Configuration du service")
		u.Runner.Run([]string{"systemctl", "enable", "docker"}, runner.WithSudo())
		u.Runner.Run([]string{"systemctl", "start", "docker"}, runner.WithSudo())
		u.CLI.PrintSuccess("Service Docker demarre et active au boot")
	}

	// Verify installation
	if u.Verify() {
		version := u.Runner.GetCommandVersion("docker", "--version")
		result.Message = "Docker installe avec succes"
		result.Version = version
	} else {
		result.AddWarning("L'installation semble terminee mais la verification a echoue")
	}

	return result
}

// getDistroCodename returns the distribution codename.
func (u *ubuntuInstaller) getDistroCodename() string {
	result := u.Runner.Run([]string{"lsb_release", "-cs"})
	if result.Success {
		return strings.TrimSpace(result.Stdout)
	}

	// Fallback: read from /etc/os-release
	result = u.Runner.Run([]string{"bash", "-c", ". /etc/os-release && echo $VERSION_CODENAME"})
	if result.Success {
		return strings.TrimSpace(result.Stdout)
	}

	return "jammy" // Default to Ubuntu 22.04
}

// getArch returns the architecture string for APT.
func (u *ubuntuInstaller) getArch() string {
	result := u.Runner.Run([]string{"dpkg", "--print-architecture"})
	if result.Success {
		return strings.TrimSpace(result.Stdout)
	}
	return "amd64"
}

// Verify verifies the Docker installation.
func (u *ubuntuInstaller) Verify() bool {
	result := u.Runner.Run([]string{"docker", "--version"})
	return result.Success
}

// Uninstall removes Docker from Ubuntu/Debian.
func (u *ubuntuInstaller) Uninstall(opts *installers.UninstallOptions) *installers.UninstallResult {
	u.CLI.PrintSection("Desinstallation de Docker")

	// Stop Docker service
	u.CLI.PrintInfo("Arret du service Docker...")
	u.Runner.Run([]string{"systemctl", "stop", "docker"}, runner.WithSudo())
	u.Runner.Run([]string{"systemctl", "disable", "docker"}, runner.WithSudo())

	// Remove Docker packages
	u.CLI.PrintInfo("Suppression des paquets Docker...")
	packages := []string{
		"docker-ce",
		"docker-ce-cli",
		"containerd.io",
		"docker-buildx-plugin",
		"docker-compose-plugin",
	}
	u.Runner.Run(
		append([]string{"apt-get", "purge", "-y"}, packages...),
		runner.WithSudo(),
	)
	u.CLI.PrintSuccess("Paquets Docker supprimes")

	// Remove Docker data
	if opts.RemoveData {
		u.CLI.PrintInfo("Suppression des donnees Docker...")
		dataDirectories := []string{
			"/var/lib/docker",
			"/var/lib/containerd",
			"/etc/docker",
			filepath.Join(u.SystemInfo.HomeDir, ".docker"),
		}

		for _, dir := range dataDirectories {
			u.Runner.Run([]string{"rm", "-rf", dir}, runner.WithSudo())
		}
		u.CLI.PrintSuccess("Donnees Docker supprimees")
	}

	// Remove repository
	u.Runner.Run([]string{"rm", "-f", "/etc/apt/sources.list.d/docker.list"}, runner.WithSudo())
	u.Runner.Run([]string{"rm", "-f", "/etc/apt/keyrings/docker.asc"}, runner.WithSudo())

	return installers.NewUninstallSuccessResult("Docker desinstalle avec succes")
}
