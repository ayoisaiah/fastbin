package fetch

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/ayoisaiah/fastbin/internal/apperr"
	"github.com/ayoisaiah/fastbin/internal/config"
	"github.com/ayoisaiah/fastbin/internal/file"
	"github.com/go-resty/resty/v2"
)

var defaultClient = &http.Client{
	// TODO: Set timeout
}

var errNotOK = &apperr.Error{
	Message: "response is not OK", // TODO: update error message
}

func DownloadFile(url string) (*file.File, error) {
	baseName := path.Base(url)

	tempFile := filepath.Join(config.TempDir(), baseName)

	out, err := os.Create(tempFile)
	if err != nil {
		return nil, err
	}

	defer out.Close()

	client := resty.NewWithClient(defaultClient)

	resp, err := client.R().SetDoNotParseResponse(true).Get(url)
	if err != nil {
		return nil, err
	}

	fileSizeBytes, err := strconv.Atoi(resp.Header().Get("Content-Length"))
	if err != nil {
		// This will show a spinner
		fileSizeBytes = -1
	}

	if resp.IsError() {
		// TODO: Return appropriate error
		return nil, errNotOK
	}

	pr := newProgress(resp.RawBody(), int64(fileSizeBytes))

	go func() {
		_, err = io.Copy(out, pr)
		if err != nil {
			log.Fatal(err) // TODO: update error
		}
	}()

	err = pr.run()
	if err != nil {
		return nil, err
	}

	fmt.Println("✅ Download completed")

	f := &file.File{
		Name:     baseName,
		Location: tempFile,
	}

	err = f.SetType()
	if err != nil {
		return nil, err
	}

	return f, nil
}
