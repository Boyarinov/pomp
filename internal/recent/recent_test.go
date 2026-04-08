package recent

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeJSONL(t *testing.T, path, content string, mtime time.Time) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

func TestScan(t *testing.T) {
	root := t.TempDir()
	sessions := filepath.Join(root, "sessions")

	realA := filepath.Join(root, "projA")
	realB := filepath.Join(root, "projB")
	for _, d := range []string{realA, realB} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	stale := filepath.Join(root, "deleted")

	now := time.Now()
	// projA older
	writeJSONL(t, filepath.Join(sessions, "a1", "s.jsonl"),
		`{"type":"session","cwd":`+quote(realA)+`,"title":"Alpha"}`+"\n", now.Add(-2*time.Hour))
	// projA newer duplicate (different subdir)
	writeJSONL(t, filepath.Join(sessions, "a2", "s.jsonl"),
		`{"type":"session","cwd":`+quote(realA)+`,"title":"Alpha2"}`+"\n", now.Add(-10*time.Minute))
	// projB
	writeJSONL(t, filepath.Join(sessions, "b1", "s.jsonl"),
		`{"type":"session","cwd":`+quote(realB)+`,"title":"-"}`+"\n", now.Add(-1*time.Hour))
	// stale
	writeJSONL(t, filepath.Join(sessions, "c1", "s.jsonl"),
		`{"type":"session","cwd":`+quote(stale)+`,"title":"Ghost"}`+"\n", now)

	got, err := Scan(sessions)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 entries, got %d: %+v", len(got), got)
	}
	if got[0].Path != realA || got[0].Title != "Alpha2" {
		t.Errorf("entry[0]=%+v", got[0])
	}
	if got[1].Path != realB || got[1].Title != filepath.Base(realB) {
		t.Errorf("entry[1]=%+v", got[1])
	}
}

func TestScanMissingDir(t *testing.T) {
	got, err := Scan(filepath.Join(t.TempDir(), "nope"))
	if err != nil || got != nil {
		t.Fatalf("want (nil, nil), got (%v, %v)", got, err)
	}
}

// quote produces a JSON string literal for a filesystem path (handles backslashes).
func quote(s string) string {
	out := []byte{'"'}
	for _, r := range s {
		switch r {
		case '\\':
			out = append(out, '\\', '\\')
		case '"':
			out = append(out, '\\', '"')
		default:
			out = append(out, string(r)...)
		}
	}
	out = append(out, '"')
	return string(out)
}
