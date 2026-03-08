package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mickamy/dotsm/internal/cmd"
	"github.com/mickamy/dotsm/internal/sm"
)

func TestPush(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	api := &stubAPI{}
	client := sm.New(api)

	input := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(input, []byte("FOO=bar\nBAZ=qux\n"), 0600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	err := cmd.Push(ctx, client, cmd.PushOptions{
		SecretID: "test/app",
		Input:    input,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !api.putCalled {
		t.Error("expected Put to be called")
	}
}

func TestPush_DryRun(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	api := &stubAPI{}
	client := sm.New(api)

	input := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(input, []byte("FOO=bar\n"), 0600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	err := cmd.Push(ctx, client, cmd.PushOptions{
		SecretID: "test/app",
		Input:    input,
		DryRun:   true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if api.putCalled {
		t.Error("expected Put NOT to be called in dry-run mode")
	}
}

func TestPush_FileNotFound(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	client := sm.New(&stubAPI{})

	err := cmd.Push(ctx, client, cmd.PushOptions{
		SecretID: "test/app",
		Input:    "/nonexistent/.env",
	})
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestPush_InvalidDotenv(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	api := &stubAPI{}
	client := sm.New(api)

	input := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(input, []byte("INVALID_LINE_NO_EQUALS\n"), 0600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	err := cmd.Push(ctx, client, cmd.PushOptions{
		SecretID: "test/app",
		Input:    input,
	})
	if err == nil {
		t.Fatal("expected error for invalid dotenv")
	}

	if api.putCalled {
		t.Error("expected Put NOT to be called on parse error")
	}
}
