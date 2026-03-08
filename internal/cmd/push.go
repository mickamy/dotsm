package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/mickamy/dotsm/internal/dotenv"
	"github.com/mickamy/dotsm/internal/sm"
)

// PushOptions holds configuration for the push command.
type PushOptions struct {
	SecretID string
	Input    string // file path; required
	DryRun   bool
}

// Push reads a dotenv file and stores it as a JSON secret in Secrets Manager.
func Push(ctx context.Context, client *sm.Client, opts PushOptions) error {
	f, err := os.Open(opts.Input)
	if err != nil {
		return fmt.Errorf("open %q: %w", opts.Input, err)
	}
	defer f.Close()

	kvs, err := dotenv.Parse(f)
	if err != nil {
		return fmt.Errorf("parse %q: %w", opts.Input, err)
	}

	if opts.DryRun {
		fmt.Fprintf(os.Stderr, "Would push %d keys to %s (dry-run)\n", len(kvs), opts.SecretID)
		return nil
	}

	if err := client.Put(ctx, opts.SecretID, kvs); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Pushed %d keys to %s\n", len(kvs), opts.SecretID)
	return nil
}
