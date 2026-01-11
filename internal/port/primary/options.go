// Package primary contains driving/inbound port interfaces.
package primary

// InstallOptions represents options for installation operations.
type InstallOptions struct {
	DryRun        bool
	NoInteraction bool

	// Docker-specific
	DockerOptions *DockerOptions

	// VSCode-specific
	VSCodeOptions *VSCodeOptions

	// Neovim-specific
	NeovimOptions *NeovimOptions

	// Zsh-specific
	ZshOptions *ZshOptions

	// Font-specific
	FontOptions *FontOptions
}

// DockerOptions contains Docker-specific installation options.
type DockerOptions struct {
	InstallCompose       bool
	AddUserToDockerGroup bool
	StartOnBoot          bool
}

// VSCodeOptions contains VSCode-specific installation options.
type VSCodeOptions struct {
	Extensions []string
}

// NeovimOptions contains Neovim-specific installation options.
type NeovimOptions struct {
	ConfigPreset string // "minimal", "full", "custom"
	CustomConfig string
}

// ZshOptions contains Zsh-specific installation options.
type ZshOptions struct {
	InstallOhMyZsh bool
	Theme          string
	Plugins        []string
}

// FontOptions contains Font-specific installation options.
type FontOptions struct {
	FontFamily string
}

// UninstallOptions represents options for uninstallation operations.
type UninstallOptions struct {
	DryRun        bool
	NoInteraction bool

	// Common options
	RemoveConfig bool
	RemoveCache  bool
	RemoveData   bool

	// Docker-specific
	RemoveImages  bool
	RemoveVolumes bool

	// Zsh-specific
	RemoveOhMyZsh bool
	RemovePlugins bool
	RemoveZshrc   bool
}
