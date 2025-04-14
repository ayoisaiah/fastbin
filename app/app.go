package app

import (
	"context"
	"io"
	"net/mail"

	"github.com/urfave/cli/v3"
)

// Get returns an Fastbin instance that reads from `reader` and writes to `writer`.
func Get(reader io.Reader, writer io.Writer) (*cli.Command, error) {
	app := CreateCLIApp(reader, writer)

	return app, nil
}

func CreateCLIApp(r io.Reader, w io.Writer) *cli.Command {
	// Override the default version printer
	oldVersionPrinter := cli.VersionPrinter
	cli.VersionPrinter = func(ctx *cli.Command) {
		oldVersionPrinter(ctx)
	}

	app := &cli.Command{
		Name: "fastbin",
		Authors: []any{
			&mail.Address{
				Name:    "Ayooluwa Isaiah",
				Address: "ayo@freshman.tech",
			},
		},
		Usage: "Fastbin is a binary manager",
		Commands: []*cli.Command{
			installCmd,
		},
		Version:                   "v0.0.1",
		EnableShellCompletion:     true,
		Flags:                     []cli.Flag{},
		UseShortOptionHandling:    true,
		DisableSliceFlagSeparator: true,
		OnUsageError: func(_ context.Context, _ *cli.Command, err error, _ bool) error {
			return err
		},
		Writer: w,
		Reader: r,
	}

	// Override the default help template
	// cli.AppHelpTemplate = helpText(app)

	return app
}
