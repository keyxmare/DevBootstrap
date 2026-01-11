// Package detector provides system detection adapters.
package detector

import (
	"bufio"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
)

// SystemDetector implements the SystemDetector port.
type SystemDetector struct{}

// NewSystemDetector creates a new SystemDetector instance.
func NewSystemDetector() *SystemDetector {
	return &SystemDetector{}
}

// Detect detects and returns the current platform.
func (d *SystemDetector) Detect() *entity.Platform {
	osType := d.detectOS()
	arch := d.detectArch()
	osName, osVersion := d.detectOSDetails(osType)
	homeDir := d.detectHomeDir()
	username := d.detectUsername()
	isRoot := os.Geteuid() == 0
	hasSudo := d.checkSudoAvailable()

	return entity.NewPlatform(
		osType,
		arch,
		osName,
		osVersion,
		homeDir,
		username,
		isRoot,
		hasSudo,
	)
}

// detectOS determines the operating system type.
func (d *SystemDetector) detectOS() valueobject.OSType {
	switch runtime.GOOS {
	case "darwin":
		return valueobject.OSMacOS
	case "linux":
		return d.detectLinuxDistro()
	default:
		return valueobject.OSUnsupported
	}
}

// detectLinuxDistro determines the Linux distribution.
func (d *SystemDetector) detectLinuxDistro() valueobject.OSType {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return valueobject.OSLinuxOther
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var idLike string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			id := strings.TrimPrefix(line, "ID=")
			id = strings.Trim(id, "\"")
			id = strings.ToLower(id)

			switch id {
			case "ubuntu":
				return valueobject.OSUbuntu
			case "debian":
				return valueobject.OSDebian
			}
		}
		if strings.HasPrefix(line, "ID_LIKE=") {
			idLike = strings.TrimPrefix(line, "ID_LIKE=")
			idLike = strings.Trim(idLike, "\"")
			idLike = strings.ToLower(idLike)
		}
	}

	// Check ID_LIKE for derivative distributions
	if strings.Contains(idLike, "ubuntu") {
		return valueobject.OSUbuntu
	}
	if strings.Contains(idLike, "debian") {
		return valueobject.OSDebian
	}

	return valueobject.OSLinuxOther
}

// detectArch determines the CPU architecture.
func (d *SystemDetector) detectArch() valueobject.Architecture {
	switch runtime.GOARCH {
	case "arm64":
		return valueobject.ArchARM64
	case "amd64":
		return valueobject.ArchAMD64
	default:
		return valueobject.ArchUnknown
	}
}

// detectHomeDir returns the user's home directory.
func (d *SystemDetector) detectHomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}

	if u, err := user.Current(); err == nil {
		return u.HomeDir
	}

	return "/tmp"
}

// detectUsername returns the current user's username.
func (d *SystemDetector) detectUsername() string {
	if username := os.Getenv("USER"); username != "" {
		return username
	}

	if u, err := user.Current(); err == nil {
		return u.Username
	}

	return ""
}

// detectOSDetails returns the OS name and version.
func (d *SystemDetector) detectOSDetails(osType valueobject.OSType) (name, version string) {
	switch osType {
	case valueobject.OSMacOS:
		name = "macOS"
		if out, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
			version = strings.TrimSpace(string(out))
		}
	case valueobject.OSUbuntu, valueobject.OSDebian, valueobject.OSLinuxOther:
		name, version = d.readOSRelease()
	default:
		name = "Unknown"
		version = ""
	}
	return
}

// readOSRelease reads the OS name and version from /etc/os-release.
func (d *SystemDetector) readOSRelease() (name, version string) {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "Linux", ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			name = strings.TrimPrefix(line, "PRETTY_NAME=")
			name = strings.Trim(name, "\"")
		}
		if strings.HasPrefix(line, "VERSION_ID=") {
			version = strings.TrimPrefix(line, "VERSION_ID=")
			version = strings.Trim(version, "\"")
		}
	}

	if name == "" {
		name = "Linux"
	}

	return
}

// checkSudoAvailable checks if sudo can be used without a password.
func (d *SystemDetector) checkSudoAvailable() bool {
	if os.Geteuid() == 0 {
		return true
	}

	if _, err := exec.LookPath("sudo"); err != nil {
		return false
	}

	cmd := exec.Command("sudo", "-n", "true")
	return cmd.Run() == nil
}
