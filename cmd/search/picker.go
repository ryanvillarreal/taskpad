package searchcmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ryanvillarreal/taskpad/internal/search"
)

var (
	pKindNote = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	pKindLink = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	pKindTask = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	pSel      = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("236")).Padding(0, 1)
	pNorm     = lipgloss.NewStyle().Padding(0, 1)
	pDim      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	pHelp     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	pMatch    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3"))
)

type pickerModel struct {
	results  []search.Result
	cursor   int
	selected *search.Result
	query    string
}

func newPicker(results []search.Result, query string) pickerModel {
	return pickerModel{
		results: results,
		query:   strings.ToLower(query),
	}
}

func (m pickerModel) Init() tea.Cmd { return nil }

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.results)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.results) > 0 {
				r := m.results[m.cursor]
				m.selected = &r
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m pickerModel) View() string {
	if len(m.results) == 0 {
		return pDim.Render("no results") + "\n" + pHelp.Render("q quit")
	}

	var rows []string
	for i, r := range m.results {
		row := m.renderRow(r, i == m.cursor)
		rows = append(rows, row)
	}

	help := pHelp.Render("↑↓/jk navigate  enter open  q quit")
	return strings.Join(rows, "\n") + "\n" + help
}

func (m pickerModel) renderRow(r search.Result, selected bool) string {
	kind := m.renderKind(r.Kind)

	var main string
	switch r.Kind {
	case search.KindNote:
		main = fmt.Sprintf("%-12s  %s", r.ID, highlightMatch(r.Snippet, m.query))
	case search.KindLink:
		title := truncate(r.Title, 36)
		main = fmt.Sprintf("%-36s  %s", title, pDim.Render(truncate(r.Snippet, 40)))
	case search.KindTask:
		main = fmt.Sprintf("%-36s  %s", truncate(r.Title, 36), pDim.Render(r.Snippet))
	}

	line := kind + "  " + main
	if selected {
		return pSel.Render(line)
	}
	return pNorm.Render(line)
}

func (m pickerModel) renderKind(k search.Kind) string {
	switch k {
	case search.KindNote:
		return pKindNote.Render("note")
	case search.KindLink:
		return pKindLink.Render("link")
	case search.KindTask:
		return pKindTask.Render("task")
	}
	return string(k)
}

func highlightMatch(text, q string) string {
	if q == "" {
		return text
	}
	lower := strings.ToLower(text)
	idx := strings.Index(lower, q)
	if idx < 0 {
		return text
	}
	return text[:idx] + pMatch.Render(text[idx:idx+len(q)]) + text[idx+len(q):]
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
