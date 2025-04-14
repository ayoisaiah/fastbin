package fastbin

import (
	"context"
	"io"
	"os"

	"github.com/ayoisaiah/fastbin/app"
	"github.com/urfave/cli/v3"
)

func defaultAction(_ context.Context, _ *cli.Command) error {
	return nil
}

func Execute(reader io.Reader, writer io.Writer) error {
	bm, err := app.Get(reader, writer)
	if err != nil {
		return err
	}

	bm.Action = defaultAction

	return bm.Run(context.Background(), os.Args)
}
