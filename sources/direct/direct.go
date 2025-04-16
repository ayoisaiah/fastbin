package direct

import (
	"github.com/ayoisaiah/fastbin/internal/fetch"
	"github.com/ayoisaiah/fastbin/internal/file"
)

type Direct struct{}

func (d *Direct) Download(url string) (*file.File, error) {
	// TODO: Possibly log the source
	// > Installing x from remote archive
	return fetch.DownloadFile(url)
}
