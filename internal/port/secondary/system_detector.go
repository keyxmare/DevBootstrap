package secondary

import "github.com/keyxmare/DevBootstrap/internal/domain/entity"

// SystemDetector defines the interface for system detection.
type SystemDetector interface {
	// Detect detects and returns the current platform.
	Detect() *entity.Platform
}
