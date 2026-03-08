package cmd

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/mickamy/dotsm/internal/dotenv"
	"github.com/mickamy/dotsm/internal/sm"
)

// DiffOptions holds configuration for the diff command.
type DiffOptions struct {
	SecretID string
	Input    string // file path; required
}

// DiffResult represents the result of comparing local and remote secrets.
type DiffResult struct {
	Added   []string          // keys in local but not in remote
	Removed []string          // keys in remote but not in local
	Changed map[string][2]string // key → [remote, local]
}

// HasDiff returns true if there are any differences.
func (d DiffResult) HasDiff() bool {
	return len(d.Added) > 0 || len(d.Removed) > 0 || len(d.Changed) > 0
}

// Diff compares a local dotenv file with a remote Secrets Manager secret.
func Diff(ctx context.Context, client *sm.Client, opts DiffOptions) (*DiffResult, error) {
	f, err := os.Open(opts.Input)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", opts.Input, err)
	}
	defer f.Close()

	local, err := dotenv.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("parse %q: %w", opts.Input, err)
	}

	remote, err := client.Get(ctx, opts.SecretID)
	if err != nil {
		return nil, err
	}

	return ComputeDiff(remote, local), nil
}

// ComputeDiff compares two maps and returns the differences.
func ComputeDiff(remote, local map[string]string) *DiffResult {
	result := &DiffResult{
		Changed: make(map[string][2]string),
	}

	for k, lv := range local {
		rv, ok := remote[k]
		if !ok {
			result.Added = append(result.Added, k)
		} else if rv != lv {
			result.Changed[k] = [2]string{rv, lv}
		}
	}

	for k := range remote {
		if _, ok := local[k]; !ok {
			result.Removed = append(result.Removed, k)
		}
	}

	slices.SortFunc(result.Added, cmp.Compare)
	slices.SortFunc(result.Removed, cmp.Compare)

	return result
}

// PrintDiff writes a human-readable diff to the given writer.
func PrintDiff(w io.Writer, r *DiffResult) {
	if !r.HasDiff() {
		fmt.Fprintln(w, "No differences.")
		return
	}

	for _, k := range r.Added {
		fmt.Fprintf(w, "+ %s\n", k)
	}
	for _, k := range r.Removed {
		fmt.Fprintf(w, "- %s\n", k)
	}

	changedKeys := make([]string, 0, len(r.Changed))
	for k := range r.Changed {
		changedKeys = append(changedKeys, k)
	}
	slices.SortFunc(changedKeys, cmp.Compare)

	for _, k := range changedKeys {
		pair := r.Changed[k]
		fmt.Fprintf(w, "~ %s: %q → %q\n", k, pair[0], pair[1])
	}
}
