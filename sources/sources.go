package sources

import (
	"net/http"

	"github.com/ayoisaiah/fastbin/internal/file"
	"github.com/ayoisaiah/fastbin/sources/direct"
)

var httpClient = http.Client{}

type Source interface {
	Download(url string) (*file.File, error)
}

func New(url string) Source {
	// TODO: Selecting the correct source
	return &direct.Direct{}
}
