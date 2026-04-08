// Package ui renders the pomp picker TUI.
package ui

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Boyarinov/pomp/internal/recent"
)

// Result is returned from Run.
type Result struct {
	Selected string
}

type item struct {
	title, path, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title + " " + i.path }

var frame = lipgloss.NewStyle().Margin(1, 2)

type model struct {
	list     list.Model
	selected string
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := frame.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter && m.list.FilterState() != list.Filtering {
			if it, ok := m.list.SelectedItem().(item); ok {
				m.selected = it.path
				return m, tea.Quit
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string { return frame.Render(m.list.View()) }

// Run shows the picker and blocks until the user selects or cancels.
func Run(entries []recent.Entry, currentDir string) (Result, error) {
	currentDir = filepath.Clean(currentDir)
	curKey := normKey(currentDir)

	items := []list.Item{item{
		title: "Current directory",
		path:  currentDir,
		desc:  currentDir,
	}}
	for _, e := range entries {
		if normKey(e.Path) == curKey {
			continue // dedup with current
		}
		items = append(items, item{
			title: e.Title,
			path:  e.Path,
			desc:  fmt.Sprintf("%s — %s", e.Path, humanize(time.Since(e.LastUsed))),
		})
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "pomp — pick a directory"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	m := model{list: l}
	final, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return Result{}, err
	}
	fm, _ := final.(model)
	return Result{Selected: fm.selected}, nil
}

func humanize(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw ago", int(d.Hours()/(24*7)))
	default:
		return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
	}
}

func normKey(p string) string {
	if runtime.GOOS == "windows" {
		return strings.ToLower(p)
	}
	return p
}
