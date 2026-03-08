package dotenv_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mickamy/dotsm/internal/dotenv"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "simple key-value",
			input: "FOO=bar\nBAZ=qux",
			want:  map[string]string{"FOO": "bar", "BAZ": "qux"},
		},
		{
			name:  "double quoted value",
			input: `FOO="hello world"`,
			want:  map[string]string{"FOO": "hello world"},
		},
		{
			name:  "single quoted value",
			input: "FOO='hello world'",
			want:  map[string]string{"FOO": "hello world"},
		},
		{
			name:  "blank lines and comments",
			input: "# comment\n\nFOO=bar\n\n# another\nBAZ=qux\n",
			want:  map[string]string{"FOO": "bar", "BAZ": "qux"},
		},
		{
			name:  "value with equals sign",
			input: "URL=https://example.com?a=1&b=2",
			want:  map[string]string{"URL": "https://example.com?a=1&b=2"},
		},
		{
			name:  "empty value",
			input: "FOO=",
			want:  map[string]string{"FOO": ""},
		},
		{
			name:  "spaces around key and value",
			input: "  FOO  =  bar  ",
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:    "missing equals",
			input:   "INVALID_LINE",
			wantErr: true,
		},
		{
			name:    "empty key",
			input:   "=value",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := dotenv.Parse(strings.NewReader(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("got %d keys, want %d", len(got), len(tt.want))
			}
			for k, wantV := range tt.want {
				gotV, ok := got[k]
				if !ok {
					t.Errorf("missing key %q", k)
					continue
				}
				if gotV != wantV {
					t.Errorf("key %q: got %q, want %q", k, gotV, wantV)
				}
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kvs  map[string]string
		want string
	}{
		{
			name: "sorted output",
			kvs:  map[string]string{"ZZZ": "last", "AAA": "first", "MMM": "middle"},
			want: "AAA=first\nMMM=middle\nZZZ=last\n",
		},
		{
			name: "value with spaces is quoted",
			kvs:  map[string]string{"FOO": "hello world"},
			want: "FOO=\"hello world\"\n",
		},
		{
			name: "value with double quote is escaped",
			kvs:  map[string]string{"FOO": `say "hi"`},
			want: "FOO=\"say \\\"hi\\\"\"\n",
		},
		{
			name: "empty value",
			kvs:  map[string]string{"FOO": ""},
			want: "FOO=\n",
		},
		{
			name: "empty map",
			kvs:  map[string]string{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := dotenv.Marshal(&buf, tt.kvs)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestParse_UnescapeDoubleQuote(t *testing.T) {
	t.Parallel()

	got, err := dotenv.Parse(strings.NewReader(`FOO="say \"hi\""` + "\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["FOO"] != `say "hi"` {
		t.Errorf("got %q, want %q", got["FOO"], `say "hi"`)
	}
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	original := map[string]string{
		"DB_HOST":      "localhost",
		"DB_PASSWORD":  "s3cret",
		"APP_NAME":     "my app",
		"QUOTED":       `say "hello"`,
		"MULTILINE":    "line1\nline2\nline3",
		"WITH_CR":      "a\r\nb",
		"BACKSLASH":    `C:\new\folder`,
		"LITERAL_BS_N": `foo\nbar`,
	}

	var buf bytes.Buffer
	if err := dotenv.Marshal(&buf, original); err != nil {
		t.Fatalf("marshal: %v", err)
	}

	parsed, err := dotenv.Parse(&buf)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	for k, wantV := range original {
		gotV, ok := parsed[k]
		if !ok {
			t.Errorf("missing key %q after round-trip", k)
			continue
		}
		if gotV != wantV {
			t.Errorf("key %q: got %q, want %q", k, gotV, wantV)
		}
	}
}
