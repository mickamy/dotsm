package dotenv

import (
	"bufio"
	"cmp"
	"fmt"
	"io"
	"slices"
	"strings"
)

// Parse reads a dotenv-formatted stream and returns key-value pairs.
// It ignores blank lines and comments (lines starting with #).
// Quotes (single or double) around values are stripped.
func Parse(r io.Reader) (map[string]string, error) {
	result := make(map[string]string)
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("line %d: invalid format: %q", lineNum, line)
		}

		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("line %d: empty key", lineNum)
		}

		value = strings.TrimSpace(value)
		value = unquote(value)

		result[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	return result, nil
}

// Marshal writes key-value pairs as dotenv-formatted output.
// Keys are sorted alphabetically for deterministic output.
func Marshal(w io.Writer, kvs map[string]string) error {
	keys := make([]string, 0, len(kvs))
	for k := range kvs {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, cmp.Compare)

	for _, k := range keys {
		v := kvs[k]
		line := k + "=" + quote(v) + "\n"
		if _, err := io.WriteString(w, line); err != nil {
			return fmt.Errorf("writing key %q: %w", k, err)
		}
	}

	return nil
}

func unquote(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			inner := s[1 : len(s)-1]
			inner = strings.ReplaceAll(inner, `\"`, `"`)
			inner = strings.ReplaceAll(inner, `\n`, "\n")
			inner = strings.ReplaceAll(inner, `\r`, "\r")
			return inner
		}
		if s[0] == '\'' && s[len(s)-1] == '\'' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func quote(s string) string {
	if strings.ContainsAny(s, " \t\n\r\"'#") {
		r := strings.NewReplacer(
			`"`, `\"`,
			"\n", `\n`,
			"\r", `\r`,
		)
		return `"` + r.Replace(s) + `"`
	}
	return s
}
