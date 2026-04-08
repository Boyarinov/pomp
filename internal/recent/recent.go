// Package recent scans omp session directories to discover recently used
// working directories.
package recent

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// Entry represents one recently used directory.
type Entry struct {
	Path     string
	Title    string
	LastUsed time.Time
}

type sessionHeader struct {
	Type  string `json:"type"`
	Cwd   string `json:"cwd"`
	Title string `json:"title"`
}

// Scan walks sessionsDir, reads the first line of the newest .jsonl in each
// subdir, validates that Path still exists, deduplicates by normalized Path
// (keeping newest LastUsed), and returns entries sorted newest first.
// A missing sessionsDir returns (nil, nil).
func Scan(sessionsDir string) ([]Entry, error) {
	dirs, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	byKey := make(map[string]Entry)
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		sub := filepath.Join(sessionsDir, d.Name())
		files, err := filepath.Glob(filepath.Join(sub, "*.jsonl"))
		if err != nil || len(files) == 0 {
			continue
		}
		// newest .jsonl by mtime
		type fi struct {
			path  string
			mtime time.Time
		}
		infos := make([]fi, 0, len(files))
		for _, f := range files {
			st, err := os.Stat(f)
			if err != nil {
				continue
			}
			infos = append(infos, fi{f, st.ModTime()})
		}
		if len(infos) == 0 {
			continue
		}
		sort.Slice(infos, func(i, j int) bool { return infos[i].mtime.After(infos[j].mtime) })
		newest := infos[0]

		hdr, ok := readHeader(newest.path)
		if !ok {
			continue
		}
		if hdr.Type != "session" || hdr.Cwd == "" {
			continue
		}
		path := filepath.Clean(hdr.Cwd)
		st, err := os.Stat(path)
		if err != nil || !st.IsDir() {
			continue
		}
		title := strings.TrimSpace(hdr.Title)
		if title == "" || title == "-" {
			title = filepath.Base(path)
		}
		key := normKey(path)
		if existing, ok := byKey[key]; ok && existing.LastUsed.After(newest.mtime) {
			continue
		}
		byKey[key] = Entry{Path: path, Title: title, LastUsed: newest.mtime}
	}

	out := make([]Entry, 0, len(byKey))
	for _, e := range byKey {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LastUsed.After(out[j].LastUsed) })
	return out, nil
}

func readHeader(path string) (sessionHeader, bool) {
	f, err := os.Open(path)
	if err != nil {
		return sessionHeader{}, false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64<<10), 1<<20)
	if !sc.Scan() {
		return sessionHeader{}, false
	}
	var h sessionHeader
	if err := json.Unmarshal(sc.Bytes(), &h); err != nil {
		return sessionHeader{}, false
	}
	return h, true
}

func normKey(path string) string {
	if runtime.GOOS == "windows" {
		return strings.ToLower(path)
	}
	return path
}
