package metrics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setVar swaps a package variable for one test and puts it back afterwards.
func setVar[T any](t *testing.T, target *T, value T) {
	t.Helper()

	original := *target
	*target = value
	t.Cleanup(func() { *target = original })
}

// writeFixture writes one file under root, creating its directories. The
// fixtures stand in for /proc and /sys.
func writeFixture(t *testing.T, root, rel, body string) {
	t.Helper()

	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create the fixture directory: %s", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write the fixture %s: %s", rel, err)
	}
}

// render runs one collector into a fresh writer and returns what it wrote.
func render(t *testing.T, collect func(*Writer)) string {
	t.Helper()

	var out strings.Builder
	w := NewWriter(&out)
	collect(w)
	if err := w.Err(); err != nil {
		t.Fatalf("writer: %s", err)
	}

	return out.String()
}

// assertText compares exposition text and prints both sides on a mismatch.
func assertText(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("exposition text differs\n--- got ---\n%s--- want ---\n%s", got, want)
	}
}
