package system

import (
	"bufio"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
)

// Detect detects and returns information about the current system.
func Detect() *SystemInfo {
	info := &SystemInfo{
		OS:       detectOS(),
		Arch:     detectArch(),
		HomeDir:  detectHomeDir(),
		Username: detectUsername(),
		IsRoot:   os.Geteuid() == 0,
		HasSudo:  checkSudoAvailable(),
	}

	info.OSName, info.OSVersion = detectOSDetails(info.OS)

	return info
}

// detectUsername returns the current user's username.
func detectUsername() string {
	// Try USER environment variable first
	if username := os.Getenv("USER"); username != "" {
		return username
	}

	// Fall back to user.Current()
	if u, err := user.Current(); err == nil {
		return u.Username
	}

	return ""
}

// detectOS determines the operating system type.
func detectOS() OSType {
	switch runtime.GOOS {
	case "darwin":
		return OSMacOS
	case "linux":
		return detectLinuxDistro()
	default:
		return OSUnsupported
	}
}

// detectLinuxDistro determines the Linux distribution.
func detectLinuxDistro() OSType {
	// Read /etc/os-release to determine the distribution
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return OSLinuxOther
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			id := strings.TrimPrefix(line, "ID=")
			id = strings.Trim(id, "\"")
			id = strings.ToLower(id)

			switch id {
			case "ubuntu":
				return OSUbuntu
			case "debian":
				return OSDebian
			default:
				// Check ID_LIKE for derivative distributions
				continue
			}
		}
		if strings.HasPrefix(line, "ID_LIKE=") {
			idLike := strings.TrimPrefix(line, "ID_LIKE=")
			idLike = strings.Trim(idLike, "\"")
			idLike = strings.ToLower(idLike)

			if strings.Contains(idLike, "ubuntu") {
				return OSUbuntu
			}
			if strings.Contains(idLike, "debian") {
				return OSDebian
			}
		}
	}

	return OSLinuxOther
}

// detectArch determines the CPU architecture.
func detectArch() Architecture {
	switch runtime.GOARCH {
	case "arm64":
		return ArchARM64
	case "amd64":
		return ArchAMD64
	default:
		return ArchUnknown
	}
}

// detectHomeDir returns the user's home directory.
func detectHomeDir() string {
	// Try HOME environment variable first
	if home := os.Getenv("HOME"); home != "" {
		return home
	}

	// Fall back to user.Current()
	if u, err := user.Current(); err == nil {
		return u.HomeDir
	}

	// Last resort
	return "/tmp"
}

// detectOSDetails returns the OS name and version.
func detectOSDetails(osType OSType) (name, version string) {
	switch osType {
	case OSMacOS:
		name = "macOS"
		// Get macOS version using sw_vers
		if out, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
			version = strings.TrimSpace(string(out))
		}
	case OSUbuntu, OSDebian, OSLinuxOther:
		name, version = readOSRelease()
	default:
		name = "Unknown"
		version = ""
	}
	return
}

// readOSRelease reads the OS name and version from /etc/os-release.
func readOSRelease() (name, version string) {
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
func checkSudoAvailable() bool {
	// If already root, sudo is not needed
	if os.Geteuid() == 0 {
		return true
	}

	// Check if sudo exists
	if _, err := exec.LookPath("sudo"); err != nil {
		return false
	}

	// Try sudo -n true (non-interactive)
	cmd := exec.Command("sudo", "-n", "true")
	return cmd.Run() == nil
}
