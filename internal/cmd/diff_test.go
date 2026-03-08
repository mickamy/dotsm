package cmd_test

import (
	"bytes"
	"testing"

	"github.com/mickamy/dotsm/internal/cmd"
)

func TestComputeDiff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		remote      map[string]string
		local       map[string]string
		wantAdded   []string
		wantRemoved []string
		wantChanged []string
		wantHasDiff bool
	}{
		{
			name:        "no diff",
			remote:      map[string]string{"A": "1", "B": "2"},
			local:       map[string]string{"A": "1", "B": "2"},
			wantHasDiff: false,
		},
		{
			name:        "added keys",
			remote:      map[string]string{"A": "1"},
			local:       map[string]string{"A": "1", "B": "2", "C": "3"},
			wantAdded:   []string{"B", "C"},
			wantHasDiff: true,
		},
		{
			name:        "removed keys",
			remote:      map[string]string{"A": "1", "B": "2", "C": "3"},
			local:       map[string]string{"A": "1"},
			wantRemoved: []string{"B", "C"},
			wantHasDiff: true,
		},
		{
			name:        "changed values",
			remote:      map[string]string{"A": "old", "B": "same"},
			local:       map[string]string{"A": "new", "B": "same"},
			wantChanged: []string{"A"},
			wantHasDiff: true,
		},
		{
			name:        "mixed",
			remote:      map[string]string{"KEEP": "v", "REMOVE": "x", "CHANGE": "old"},
			local:       map[string]string{"KEEP": "v", "ADD": "y", "CHANGE": "new"},
			wantAdded:   []string{"ADD"},
			wantRemoved: []string{"REMOVE"},
			wantChanged: []string{"CHANGE"},
			wantHasDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := cmd.ComputeDiff(tt.remote, tt.local)

			if got := result.HasDiff(); got != tt.wantHasDiff {
				t.Errorf("HasDiff: got %v, want %v", got, tt.wantHasDiff)
			}

			assertSliceEqual(t, "Added", result.Added, tt.wantAdded)
			assertSliceEqual(t, "Removed", result.Removed, tt.wantRemoved)

			gotChanged := make([]string, 0, len(result.Changed))
			for k := range result.Changed {
				gotChanged = append(gotChanged, k)
			}
			assertSliceEqual(t, "Changed", gotChanged, tt.wantChanged)
		})
	}
}

func TestPrintDiff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result *cmd.DiffResult
		want   string
	}{
		{
			name:   "no diff",
			result: &cmd.DiffResult{Changed: map[string][2]string{}},
			want:   "No differences.\n",
		},
		{
			name: "all types",
			result: &cmd.DiffResult{
				Added:   []string{"NEW_KEY"},
				Removed: []string{"OLD_KEY"},
				Changed: map[string][2]string{"MOD_KEY": {"old", "new"}},
			},
			want: "+ NEW_KEY\n- OLD_KEY\n~ MOD_KEY: \"old\" → \"new\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			cmd.PrintDiff(&buf, tt.result)
			if got := buf.String(); got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func assertSliceEqual(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) == 0 {
		got = nil
	}
	if len(want) == 0 {
		want = nil
	}
	if len(got) != len(want) {
		t.Errorf("%s: got %v, want %v", label, got, want)
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("%s[%d]: got %q, want %q", label, i, got[i], want[i])
		}
	}
}
