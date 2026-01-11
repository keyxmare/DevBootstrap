package valueobject

// AppTag represents a category tag for applications.
type AppTag string

const (
	// TagApp represents a general application.
	TagApp AppTag = "app"
	// TagConfig represents a configuration package.
	TagConfig AppTag = "config"
	// TagEditor represents an editor application.
	TagEditor AppTag = "editeur"
	// TagShell represents a shell application.
	TagShell AppTag = "shell"
	// TagContainer represents a container tool.
	TagContainer AppTag = "container"
	// TagFont represents a font package.
	TagFont AppTag = "police"
)

// String returns the string representation of AppTag.
func (t AppTag) String() string {
	return string(t)
}
