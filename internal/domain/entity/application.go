package entity

import "github.com/keyxmare/DevBootstrap/internal/domain/valueobject"

// Application represents an installable application in the domain.
type Application struct {
	id          valueobject.AppID
	name        string
	description string
	tags        []valueobject.AppTag
	status      valueobject.AppStatus
	version     valueobject.Version
}

// NewApplication creates a new Application instance.
func NewApplication(id string, name, description string, tags []valueobject.AppTag) *Application {
	return &Application{
		id:          valueobject.NewAppID(id),
		name:        name,
		description: description,
		tags:        tags,
		status:      valueobject.StatusNotInstalled,
	}
}

// ID returns the application ID.
func (a *Application) ID() valueobject.AppID {
	return a.id
}

// Name returns the application name.
func (a *Application) Name() string {
	return a.name
}

// Description returns the application description.
func (a *Application) Description() string {
	return a.description
}

// Tags returns the application tags.
func (a *Application) Tags() []valueobject.AppTag {
	return a.tags
}

// Status returns the current installation status.
func (a *Application) Status() valueobject.AppStatus {
	return a.status
}

// Version returns the installed version.
func (a *Application) Version() valueobject.Version {
	return a.version
}

// IsInstalled returns true if the application is installed.
func (a *Application) IsInstalled() bool {
	return a.status.IsInstalled()
}

// MarkInstalled marks the application as installed with the given version.
func (a *Application) MarkInstalled(version valueobject.Version) {
	a.status = valueobject.StatusInstalled
	a.version = version
}

// MarkNotInstalled marks the application as not installed.
func (a *Application) MarkNotInstalled() {
	a.status = valueobject.StatusNotInstalled
	a.version = valueobject.Version{}
}

// UpdateStatus updates the application status and version.
func (a *Application) UpdateStatus(status valueobject.AppStatus, version valueobject.Version) {
	a.status = status
	a.version = version
}
