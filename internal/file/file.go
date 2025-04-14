package file

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/ayoisaiah/fastbin/internal/apperr"
	"github.com/charmbracelet/huh"
	"github.com/ulikunitz/xz"
)

var ExecNotFound = &apperr.Error{
	Message: "Executable not found",
}

type Type int

const (
	BinaryFile Type = iota
	ArchiveFile
)

var TempDir string

func init() {
	tempDir := os.TempDir()

	err := os.MkdirAll(filepath.Join(tempDir, "fastbin"), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	TempDir = filepath.Join(os.TempDir(), "fastbin")
}

type File struct {
	Name     string
	Type     Type
	Location string
	Binaries []Binary
	Reader   *tar.Reader
}

type Binary struct {
	Name     string
	Location string
	Hash     string
}

type Executables []*tar.Header

func (f *File) FindExecutable() (Executables, error) {
	var executables Executables

	file, err := os.Open(f.Location)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	fileExt := filepath.Ext(f.Location)

	var r io.Reader

	switch fileExt {
	case ".gz":
		r, err = gzip.NewReader(file)
	case ".bzip2":
		r = bzip2.NewReader(file)
	case ".xz":
		r, err = xz.NewReader(file)
	}

	if err != nil {
		return nil, err
	}

	tr := tar.NewReader(r)

	f.Reader = tr

	files := make(map[string]*tar.Header)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			if len(executables) == 0 {
				return prompt(files)
			}

			break
		}

		if err != nil {
			return nil, err
		}

		if hdr.Typeflag == tar.TypeDir {
			continue
		}

		if filepath.Ext(hdr.Name) != "" {
			continue
		}

		files[hdr.Name] = hdr

		if isExecutable(hdr.Mode) {
			executables = append(executables, hdr)
		}
	}

	return executables, nil
}

func (f *File) Extract(execs Executables) error {
	var e []*tar.Header

	if len(execs) > 1 {
		var options []huh.Option[*tar.Header]

		for _, v := range execs {
			options = append(options, huh.NewOption(v.Name, v))
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[*tar.Header]().
					Title("Which binaries to install?").
					Options(options...).
					Value(&e),
			),
		)

		err := form.Run()
		if err != nil {
			return fmt.Errorf("form interaction failed: %w", err)
		}

	} else {
		e = execs
	}

	for _, hdr := range e {
		tempFile := filepath.Join(TempDir, hdr.Name)

		err := os.MkdirAll(filepath.Dir(tempFile), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directories: %w", err)
		}

		outFile, err := os.Create(tempFile)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		if _, err := io.Copy(outFile, f.Reader); err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}

		err = os.Chmod(outFile.Name(), os.FileMode(hdr.Mode))
		if err != nil {
			return fmt.Errorf("os.Chomd failed: %w", err)
		}

		hash, err := getHash(outFile)
		if err != nil {
			return fmt.Errorf("hash failed: %w", err)
		}

		f.Binaries = append(f.Binaries, Binary{
			Name:     hdr.Name,
			Location: tempFile,
			Hash:     hash,
		})

		_ = outFile.Close()
	}

	return nil
}

func prompt(files map[string]*tar.Header) (Executables, error) {
	var selected string

	var options []huh.Option[string]

	for _, v := range files {
		options = append(options, huh.NewOption(v.Name, v.Name))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select the binary file").
				Options(options...).
				Value(&selected),
		),
	)

	err := form.Run()
	if err != nil {
		return nil, fmt.Errorf("form interaction failed: %w", err)
	}

	v, ok := files[selected]
	if !ok {
		return nil, ExecNotFound
	}

	return []*tar.Header{v}, nil
}

func isExecutable(mode int64) bool {
	// Check if any executable bits are set
	return mode&0100 != 0 || mode&0010 != 0 || mode&0001 != 0
}

// getHash retrieves the appropriate hash value for the specified file.
func getHash(file *os.File) (string, error) {
	newHash := sha256.New()

	if _, err := io.Copy(newHash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(newHash.Sum(nil)), nil
}
