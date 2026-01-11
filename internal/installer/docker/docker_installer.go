package docker

import (
	"fmt"

	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
	"github.com/keyxmare/DevBootstrap/internal/installer/strategy"
)

// NewStrategy creates the appropriate Docker installer strategy for the platform.
func NewStrategy(deps strategy.Dependencies, platform *entity.Platform) (strategy.InstallerStrategy, error) {
	if platform.IsMacOS() {
		return NewMacOSStrategy(deps, platform), nil
	}
	if platform.IsDebian() {
		return NewUbuntuStrategy(deps, platform), nil
	}
	return nil, fmt.Errorf("unsupported platform for Docker: %s", platform.OS())
}
