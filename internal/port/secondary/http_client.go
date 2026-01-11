package secondary

import (
	"context"
	"io"
)

// HTTPClient defines the interface for HTTP operations.
type HTTPClient interface {
	// Download downloads a file from URL to destination.
	Download(ctx context.Context, url, destination string) error

	// DownloadToWriter downloads content to a writer.
	DownloadToWriter(ctx context.Context, url string, w io.Writer) error

	// Get performs an HTTP GET and returns the response body.
	Get(ctx context.Context, url string) ([]byte, error)
}
