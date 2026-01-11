// Package repository provides application repository adapters.
package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
)

// MemoryRepository implements AppRepository using in-memory storage.
type MemoryRepository struct {
	mu   sync.RWMutex
	apps map[string]*entity.Application
}

// NewMemoryRepository creates a new MemoryRepository instance.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		apps: make(map[string]*entity.Application),
	}
}

// GetAll returns all registered applications.
func (r *MemoryRepository) GetAll(ctx context.Context) ([]*entity.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	apps := make([]*entity.Application, 0, len(r.apps))
	for _, app := range r.apps {
		apps = append(apps, app)
	}
	return apps, nil
}

// GetByID returns an application by ID.
func (r *MemoryRepository) GetByID(ctx context.Context, id valueobject.AppID) (*entity.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	app, ok := r.apps[id.String()]
	if !ok {
		return nil, fmt.Errorf("application not found: %s", id)
	}
	return app, nil
}

// Save saves or updates an application.
func (r *MemoryRepository) Save(ctx context.Context, app *entity.Application) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.apps[app.ID().String()] = app
	return nil
}

// UpdateStatus updates the status of an application.
func (r *MemoryRepository) UpdateStatus(ctx context.Context, id valueobject.AppID, status valueobject.AppStatus, version valueobject.Version) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	app, ok := r.apps[id.String()]
	if !ok {
		return fmt.Errorf("application not found: %s", id)
	}

	app.UpdateStatus(status, version)
	return nil
}

// Delete removes an application from the repository.
func (r *MemoryRepository) Delete(ctx context.Context, id valueobject.AppID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.apps, id.String())
	return nil
}
