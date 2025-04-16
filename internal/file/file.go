package file

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ayoisaiah/fastbin/internal/apperr"
	"github.com/ayoisaiah/fastbin/internal/config"
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

// newCompressedReader returns an appropriate reader based on the file extension
func newCompressedReader(file *os.File, fileExt string) (io.Reader, error) {
	switch fileExt {
	case ".gz":
		return gzip.NewReader(file)
	case ".bzip2":
		return bzip2.NewReader(file), nil
	case ".xz":
		return xz.NewReader(file)
	default:
		return nil, fmt.Errorf("unsupported compression format: %s", fileExt)
	}
}

// isPotentialBinary checks if a file could be a binary based on its name and type
func isPotentialBinary(hdr *tar.Header) bool {
	return hdr.Typeflag != tar.TypeDir && filepath.Ext(hdr.Name) == ""
}

func (f *File) SetType() error {
	ext := filepath.Ext(f.Location)

	fi, err := os.Stat(f.Location)
	if err != nil {
		return err
	}

	f.Type = ArchiveFile
	if ext == "" {
		f.Type = BinaryFile
		f.Binaries = []Binary{
			{
				Name:     f.Name,
				Location: f.Location,
			},
		}

		err := os.Chmod(f.Location, fi.Mode()|0100)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *File) IsBinary() bool {
	return f.Type == BinaryFile
}

func (f *File) FindExecutable() (Executables, error) {
	file, err := os.Open(f.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer file.Close()

	fileExt := filepath.Ext(f.Location)

	r, err := newCompressedReader(file, fileExt)
	if err != nil {
		return nil, fmt.Errorf("failed to create compressed reader: %w", err)
	}

	f.Reader = tar.NewReader(r)

	executables, files, err := f.findBinaries(f.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to find binaries: %w", err)
	}

	if len(executables) == 0 {
		return prompt(files)
	}

	return executables, nil
}

func (f *File) findBinaries(
	tr *tar.Reader,
) (Executables, map[string]*tar.Header, error) {
	var executables Executables
	files := make(map[string]*tar.Header)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, nil, err
		}

		if !isPotentialBinary(hdr) {
			continue
		}

		files[hdr.Name] = hdr

		if isExecutable(hdr.Mode) {
			executables = append(executables, hdr)
		}
	}

	return executables, files, nil
}

// makeExecutable ensures the file has execute permissions while preserving other permissions
func makeExecutable(path string, mode os.FileMode) error {
	// Add execute permission for user (owner)
	executableMode := mode | 0100
	return os.Chmod(path, executableMode)
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
					Title("Multiple executables found in the archive. What would you like to install?").
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
		tempFile := filepath.Join(config.TempDir(), hdr.Name)

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

		// Convert tar header mode to os.FileMode and ensure it's executable
		mode := os.FileMode(hdr.Mode)
		if err := makeExecutable(outFile.Name(), mode); err != nil {
			return fmt.Errorf("failed to make file executable: %w", err)
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

		err = outFile.Close()
		if err != nil {
			return fmt.Errorf("closing file failed: %w", err)
		}
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
