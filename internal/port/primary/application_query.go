package primary

import (
	"context"

	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
)

// ApplicationQuery defines the primary port for querying applications.
type ApplicationQuery interface {
	// ListAll returns all available applications.
	ListAll(ctx context.Context) ([]*entity.Application, error)

	// GetByID returns an application by its ID.
	GetByID(ctx context.Context, id valueobject.AppID) (*entity.Application, error)

	// CheckStatus checks the installation status of an application.
	CheckStatus(ctx context.Context, id valueobject.AppID) (valueobject.AppStatus, string, error)

	// ListInstalled returns only installed applications.
	ListInstalled(ctx context.Context) ([]*entity.Application, error)

	// ListNotInstalled returns only not-installed applications.
	ListNotInstalled(ctx context.Context) ([]*entity.Application, error)

	// RefreshStatus refreshes the status of all applications.
	RefreshStatus(ctx context.Context) error
}
