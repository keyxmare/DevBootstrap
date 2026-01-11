// Package http provides HTTP client adapters.
package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/keyxmare/DevBootstrap/internal/port/secondary"
)

// Downloader implements HTTPClient for downloading files.
type Downloader struct {
	client   *http.Client
	reporter secondary.ProgressReporter
}

// NewDownloader creates a new Downloader instance.
func NewDownloader(reporter secondary.ProgressReporter) *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 30 * time.Minute,
		},
		reporter: reporter,
	}
}

// Download downloads a file from URL to destination.
func (d *Downloader) Download(ctx context.Context, url, destination string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	out, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// DownloadToWriter downloads content to a writer.
func (d *Downloader) DownloadToWriter(ctx context.Context, url string, w io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}

	return nil
}

// Get performs an HTTP GET and returns the response body.
func (d *Downloader) Get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get failed with status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
