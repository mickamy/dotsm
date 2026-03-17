package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/mickamy/dotsm/internal/cmd"
	"github.com/mickamy/dotsm/internal/sm"
)

var version = "dev"

const usage = `dotsm - Sync AWS Secrets Manager with .env files

Usage:
  dotsm <command> [options]

Commands:
  pull     Fetch a secret and write as .env
  push     Read a .env file and store as a secret
  diff     Compare local .env with remote secret
  version  Print version

Run 'dotsm <command> -h' for command-specific help.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	subcmd := os.Args[1]

	switch subcmd {
	case "pull":
		runPull(os.Args[2:])
	case "push":
		runPush(os.Args[2:])
	case "diff":
		runDiff(os.Args[2:])
	case "version":
		fmt.Println("dotsm", version)
	case "-h", "--help", "help":
		fmt.Fprint(os.Stdout, usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", subcmd)
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}
}

func runPull(args []string) {
	fs := flag.NewFlagSet("pull", flag.ExitOnError)
	secretID := fs.String("secret", "", "Secret ID (required)")
	output := fs.String("output", "-", "Output file path (default: stdout)")
	region := fs.String("region", "", "AWS region (overrides default)")
	profile := fs.String("profile", "", "AWS profile")
	_ = fs.Parse(args)

	if *secretID == "" {
		fmt.Fprintln(os.Stderr, "error: -secret is required")
		fs.Usage()
		os.Exit(1)
	}

	client := newClient(*region, *profile)
	if err := cmd.Pull(context.Background(), client, cmd.PullOptions{
		SecretID: *secretID,
		Output:   *output,
	}); err != nil {
		fatal(err)
	}
}

func runPush(args []string) {
	fs := flag.NewFlagSet("push", flag.ExitOnError)
	secretID := fs.String("secret", "", "Secret ID (required)")
	input := fs.String("input", ".env", "Input .env file path")
	dryRun := fs.Bool("dry-run", false, "Show what would be pushed without writing")
	region := fs.String("region", "", "AWS region (overrides default)")
	profile := fs.String("profile", "", "AWS profile")
	_ = fs.Parse(args)

	if *secretID == "" {
		fmt.Fprintln(os.Stderr, "error: -secret is required")
		fs.Usage()
		os.Exit(1)
	}

	client := newClient(*region, *profile)
	if err := cmd.Push(context.Background(), client, cmd.PushOptions{
		SecretID: *secretID,
		Input:    *input,
		DryRun:   *dryRun,
	}); err != nil {
		fatal(err)
	}
}

func runDiff(args []string) {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)
	secretID := fs.String("secret", "", "Secret ID (required)")
	input := fs.String("input", ".env", "Input .env file path")
	showValues := fs.Bool("show-values", false, "Show actual secret values in diff output")
	region := fs.String("region", "", "AWS region (overrides default)")
	profile := fs.String("profile", "", "AWS profile")
	_ = fs.Parse(args)

	if *secretID == "" {
		fmt.Fprintln(os.Stderr, "error: -secret is required")
		fs.Usage()
		os.Exit(1)
	}

	client := newClient(*region, *profile)
	result, err := cmd.Diff(context.Background(), client, cmd.DiffOptions{
		SecretID: *secretID,
		Input:    *input,
	})
	if err != nil {
		fatal(err)
	}

	cmd.PrintDiff(os.Stdout, result, *showValues)
	if result.HasDiff() {
		os.Exit(1)
	}
}

func newClient(region, profile string) *sm.Client {
	var opts []func(*config.LoadOptions) error
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		fatal(fmt.Errorf("load AWS config: %w", err))
	}

	return sm.New(secretsmanager.NewFromConfig(cfg))
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
