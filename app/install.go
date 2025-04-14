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

func Move(source, destination string) error {
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

	dst, err := os.Create(destination)
	if err != nil {
		src.Close()
		return err
	}

	_, err = io.Copy(dst, src)
	src.Close()
	dst.Close()
	if err != nil {
		return err
	}
	fi, err := os.Stat(source)
	if err != nil {
		os.Remove(destination)
		return err
	}

	err = os.Chmod(destination, fi.Mode())
	if err != nil {
		os.Remove(destination)
		return err
	}

	os.Remove(source)

	return nil
}

func installAction(_ context.Context, cmd *cli.Command) error {
	args := cmd.Args()

	source := sources.New(args.Get(0))

	f, err := source.Download(args.Get(0))
	if err != nil {
		return err
	}

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

	for _, v := range f.Binaries {
		err := Move(
			v.Location,
			filepath.Join(xdg.BinHome, filepath.Base(v.Name)),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
