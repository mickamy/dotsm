package cmd

import (
	"fmt"
	"io"
	"os"

	"context"

	"github.com/mickamy/dotsm/internal/dotenv"
	"github.com/mickamy/dotsm/internal/sm"
)

// PullOptions holds configuration for the pull command.
type PullOptions struct {
	SecretID string
	Output   string // file path; empty or "-" means stdout
}

// Pull fetches a secret from Secrets Manager and writes it as a dotenv file.
func Pull(ctx context.Context, client *sm.Client, opts PullOptions) error {
	kvs, err := client.Get(ctx, opts.SecretID)
	if err != nil {
		return fmt.Errorf("pull: %w", err)
	}

	var w io.Writer
	if opts.Output == "" || opts.Output == "-" {
		w = os.Stdout
	} else {
		f, err := os.Create(opts.Output)
		if err != nil {
			return fmt.Errorf("create %q: %w", opts.Output, err)
		}
		defer func() { _ = f.Close() }()
		w = f
	}

	if err := dotenv.Marshal(w, kvs); err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if opts.Output != "" && opts.Output != "-" {
		fmt.Fprintf(os.Stderr, "Wrote %d keys to %s\n", len(kvs), opts.Output)
	}

	return nil
}
