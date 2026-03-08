//nolint:gosec // test data and temp file paths
package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mickamy/dotsm/internal/cmd"
	"github.com/mickamy/dotsm/internal/sm"
)

func TestPull_ToFile(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	client := sm.New(&stubAPI{
		secret: `{"DB_HOST":"localhost","DB_PORT":"5432"}`,
	})

	out := filepath.Join(t.TempDir(), ".env")
	err := cmd.Pull(ctx, client, cmd.PullOptions{
		SecretID: "test/app",
		Output:   out,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	got := string(data)
	want := "DB_HOST=localhost\nDB_PORT=5432\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}

	info, err := os.Stat(out)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("permissions: got %o, want 0600", perm)
	}
}

func TestPull_ToFile_OverwritePermissions(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	client := sm.New(&stubAPI{
		secret: `{"FOO":"bar"}`,
	})

	out := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(out, []byte("OLD=data\n"), 0644); err != nil {
		t.Fatalf("create existing file: %v", err)
	}

	if err := cmd.Pull(ctx, client, cmd.PullOptions{
		SecretID: "test/app",
		Output:   out,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(out)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("permissions after overwrite: got %o, want 0600", perm)
	}
}

func TestPull_ToStdout(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	client := sm.New(&stubAPI{
		secret: `{"KEY":"val"}`,
	})

	// stdout output — just verify no error
	err := cmd.Pull(ctx, client, cmd.PullOptions{
		SecretID: "test/app",
		Output:   "-",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPull_InvalidOutputPath(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	client := sm.New(&stubAPI{
		secret: `{"KEY":"val"}`,
	})

	err := cmd.Pull(ctx, client, cmd.PullOptions{
		SecretID: "test/app",
		Output:   "/nonexistent/dir/.env",
	})
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}
