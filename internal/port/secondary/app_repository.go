package secondary

import (
	"context"

	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
)

// AppRepository defines the interface for application storage/registry.
type AppRepository interface {
	// GetAll returns all registered applications.
	GetAll(ctx context.Context) ([]*entity.Application, error)

	// GetByID returns an application by ID.
	GetByID(ctx context.Context, id valueobject.AppID) (*entity.Application, error)

	// Save saves or updates an application.
	Save(ctx context.Context, app *entity.Application) error

	// UpdateStatus updates the status of an application.
	UpdateStatus(ctx context.Context, id valueobject.AppID, status valueobject.AppStatus, version valueobject.Version) error

	// Delete removes an application from the repository.
	Delete(ctx context.Context, id valueobject.AppID) error
}
