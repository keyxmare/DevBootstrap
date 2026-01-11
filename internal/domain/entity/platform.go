// Package entity contains domain entities.
package entity

import "github.com/keyxmare/DevBootstrap/internal/domain/valueobject"

// Platform represents the target system platform.
type Platform struct {
	os        valueobject.OSType
	arch      valueobject.Architecture
	osName    string
	osVersion string
	homeDir   string
	username  string
	isRoot    bool
	hasSudo   bool
}

// NewPlatform creates a new Platform instance.
func NewPlatform(
	os valueobject.OSType,
	arch valueobject.Architecture,
	osName, osVersion, homeDir, username string,
	isRoot, hasSudo bool,
) *Platform {
	return &Platform{
		os:        os,
		arch:      arch,
		osName:    osName,
		osVersion: osVersion,
		homeDir:   homeDir,
		username:  username,
		isRoot:    isRoot,
		hasSudo:   hasSudo,
	}
}

// OS returns the operating system type.
func (p *Platform) OS() valueobject.OSType {
	return p.os
}

// Arch returns the CPU architecture.
func (p *Platform) Arch() valueobject.Architecture {
	return p.arch
}

// OSName returns the OS name (e.g., "macOS", "Ubuntu 22.04").
func (p *Platform) OSName() string {
	return p.osName
}

// OSVersion returns the OS version string.
func (p *Platform) OSVersion() string {
	return p.osVersion
}

// HomeDir returns the user's home directory.
func (p *Platform) HomeDir() string {
	return p.homeDir
}

// Username returns the current username.
func (p *Platform) Username() string {
	return p.username
}

// IsRoot returns true if running as root.
func (p *Platform) IsRoot() bool {
	return p.isRoot
}

// HasSudo returns true if sudo is available without password.
func (p *Platform) HasSudo() bool {
	return p.hasSudo
}

// IsMacOS returns true if the platform is macOS.
func (p *Platform) IsMacOS() bool {
	return p.os.IsMacOS()
}

// IsLinux returns true if the platform is Linux.
func (p *Platform) IsLinux() bool {
	return p.os.IsLinux()
}

// IsDebian returns true if the platform is Debian-based.
func (p *Platform) IsDebian() bool {
	return p.os.IsDebian()
}
