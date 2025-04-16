package app

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/ayoisaiah/fastbin/sources"
	"github.com/urfave/cli/v3"
)

var installCmd = &cli.Command{
	Name:        "install",
	Usage:       "install <url>",
	Aliases:     []string{"i"},
	Description: "Install a binary from a URL",
	Action:      installAction,
}

func moveToBinDir(source, destination string) error {
	err := os.Rename(source, destination)
	if err != nil {
		return moveCrossDevice(source, destination)
	}

	return nil
}

func moveCrossDevice(source, destination string) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}

	defer src.Close()

	dst, err := os.Create(destination)
	if err != nil {
		return err
	}

	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	fi, err := os.Stat(source)
	if err != nil {
		return err
	}

	err = os.Chmod(destination, fi.Mode())
	if err != nil {
		return err
	}

	return os.Remove(source)
}

func installAction(_ context.Context, cmd *cli.Command) error {
	args := cmd.Args()

	url := args.Get(0)

	source := sources.New(url)

	f, err := source.Download(url)
	if err != nil {
		return err
	}

	if !f.IsBinary() {
		binaries, err := f.FindExecutable()
		if err != nil {
			return err
		}

		if len(binaries) <= 0 {
			return errors.New("no binaries found in archive")
		}

		err = f.Extract(binaries)
		if err != nil {
			return err
		}
	}

	os.Exit(1)

	for _, v := range f.Binaries {
		err := moveToBinDir(
			v.Location,
			filepath.Join(xdg.BinHome, filepath.Base(v.Name)),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
