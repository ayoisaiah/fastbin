package direct

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/ayoisaiah/fastbin/internal/file"
)

type Direct struct{}

func (d *Direct) Download(url string) (*file.File, error) {
	baseName := path.Base(url)

	tempFile := filepath.Join(file.TempDir, baseName)

	out, err := os.Create(tempFile)
	if err != nil {
		return nil, err
	}

	defer out.Close()

	// TODO: Use resty
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		// TODO: Add status
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body out to a file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return nil, err
	}

	return &file.File{
		Name:     baseName,
		Location: tempFile,
	}, nil
}
