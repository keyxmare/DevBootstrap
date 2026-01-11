package primary

import (
	"context"

	"github.com/keyxmare/DevBootstrap/internal/domain/result"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
)

// InstallerService defines the primary port for installation operations.
type InstallerService interface {
	// Install installs an application by ID.
	Install(ctx context.Context, appID valueobject.AppID, opts InstallOptions) (*result.InstallResult, error)

	// Uninstall removes an application by ID.
	Uninstall(ctx context.Context, appID valueobject.AppID, opts UninstallOptions) (*result.UninstallResult, error)

	// InstallMultiple installs multiple applications.
	InstallMultiple(ctx context.Context, appIDs []valueobject.AppID, opts InstallOptions) (map[string]*result.InstallResult, error)

	// UninstallMultiple uninstalls multiple applications.
	UninstallMultiple(ctx context.Context, appIDs []valueobject.AppID, opts UninstallOptions) (map[string]*result.UninstallResult, error)

	// Verify verifies that an application is installed correctly.
	Verify(ctx context.Context, appID valueobject.AppID) bool
}
