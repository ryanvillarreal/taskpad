package taskscmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ryanvillarreal/taskpad/internal/config"
	"github.com/ryanvillarreal/taskpad/internal/tasks"
	"github.com/spf13/cobra"
)

var (
	hActive = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	hPaused = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	hClosed = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("8"))

	colBase    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	colFocused = colBase.BorderForeground(lipgloss.Color("39"))

	selStyle    = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("236")).Padding(0, 1)
	normStyle   = lipgloss.NewStyle().Padding(0, 1)
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	overdueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	soonStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)

type editorDoneMsg struct{ err error }

type boardModel struct {
	cols   [3][]*tasks.Task
	col    int
	row    int
	svc    *tasks.Service
	cfg    config.Config
	width  int
	status string
}

func newBoard(svc *tasks.Service, cfg config.Config) boardModel {
	m := boardModel{svc: svc, cfg: cfg, width: 120}
	return m.reloaded()
}

func (m boardModel) reloaded() boardModel {
	all, _ := m.svc.List()
	m.cols = [3][]*tasks.Task{{}, {}, {}}
	for _, t := range all {
		switch t.Status {
		case tasks.StatusActive:
			m.cols[0] = append(m.cols[0], t)
		case tasks.StatusPaused:
			m.cols[1] = append(m.cols[1], t)
		case tasks.StatusClosed:
			m.cols[2] = append(m.cols[2], t)
		}
	}
	return m.clamped()
}

func (m boardModel) clamped() boardModel {
	if m.col < 0 {
		m.col = 0
	}
	if m.col > 2 {
		m.col = 2
	}
	n := len(m.cols[m.col])
	if n == 0 {
		m.row = 0
	} else if m.row >= n {
		m.row = n - 1
	}
	if m.row < 0 {
		m.row = 0
	}
	return m
}

func (m boardModel) current() *tasks.Task {
	col := m.cols[m.col]
	if len(col) == 0 || m.row >= len(col) {
		return nil
	}
	return col[m.row]
}

func (m boardModel) Init() tea.Cmd { return nil }

func (m boardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

	case tea.KeyMsg:
		m.status = ""
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.row > 0 {
				m.row--
			}
		case "down", "j":
			if m.row < len(m.cols[m.col])-1 {
				m.row++
			}
		case "left", "h":
			m.col--
			m = m.clamped()
		case "right", "l":
			m.col++
			m = m.clamped()
		case "c":
			if t := m.current(); t != nil {
				if _, err := m.svc.SetStatus(t.ID, tasks.StatusClosed); err != nil {
					m.status = "error: " + err.Error()
				} else {
					m.status = "closed " + t.ID
					m = m.reloaded()
				}
			}
		case "p":
			if t := m.current(); t != nil {
				if _, err := m.svc.SetStatus(t.ID, tasks.StatusPaused); err != nil {
					m.status = "error: " + err.Error()
				} else {
					m.status = "paused " + t.ID
					m = m.reloaded()
				}
			}
		case "a":
			if t := m.current(); t != nil {
				if _, err := m.svc.SetStatus(t.ID, tasks.StatusActive); err != nil {
					m.status = "error: " + err.Error()
				} else {
					m.status = "activated " + t.ID
					m = m.reloaded()
				}
			}
		case "enter":
			if t := m.current(); t != nil {
				path := filepath.Join(m.cfg.TasksDir, t.ID+".md")
				editorBin := os.Getenv("EDITOR")
				if editorBin == "" {
					editorBin = "vi"
				}
				return m, tea.ExecProcess(exec.Command(editorBin, path), func(err error) tea.Msg {
					return editorDoneMsg{err}
				})
			}
		}

	case editorDoneMsg:
		if msg.err != nil {
			m.status = "editor error: " + msg.err.Error()
		}
		m = m.reloaded()
	}

	return m, nil
}

func (m boardModel) View() string {
	colW := (m.width - 12) / 3
	if colW < 24 {
		colW = 24
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderCol(0, "active", hActive, colW),
		m.renderCol(1, "paused", hPaused, colW),
		m.renderCol(2, "closed", hClosed, colW),
	)

	help := helpStyle.Render("↑↓/jk  ←→/hl columns  enter edit  c close  p pause  a activate  q quit")
	if m.status != "" {
		help = statusStyle.Render(m.status) + "  " + help
	}

	return board + "\n" + help
}

func (m boardModel) renderCol(idx int, label string, hs lipgloss.Style, width int) string {
	ts := m.cols[idx]
	header := hs.Render(fmt.Sprintf("%s (%d)", label, len(ts)))
	sep := dimStyle.Render(strings.Repeat("─", width))

	rows := []string{header, sep}
	if len(ts) == 0 {
		rows = append(rows, dimStyle.Render("(empty)"))
	} else {
		for i, t := range ts {
			rows = append(rows, m.renderTask(t, idx == m.col && i == m.row, width))
		}
	}

	style := colBase.Width(width)
	if idx == m.col {
		style = colFocused.Width(width)
	}
	return style.Render(strings.Join(rows, "\n"))
}

func (m boardModel) renderTask(t *tasks.Task, selected bool, width int) string {
	maxTitle := width - 9
	title := t.Title
	if len(title) > maxTitle {
		title = title[:maxTitle-1] + "…"
	}

	line := dimStyle.Render(t.ID) + "  " + title
	if !t.DueAt.IsZero() {
		line += "\n         " + fmtDue(t.DueAt)
	}

	if selected {
		return selStyle.Render(line)
	}
	return normStyle.Render(line)
}

func fmtDue(due time.Time) string {
	now := time.Now()
	s := "due " + due.Local().Format("Mon Jan 2 3:04pm")
	if due.Before(now) {
		return overdueStyle.Render(s)
	}
	if due.Before(now.Add(24 * time.Hour)) {
		return soonStyle.Render(s)
	}
	return dimStyle.Render(s)
}

var boardCmd = &cobra.Command{
	Use:   "board",
	Short: "interactive kanban board",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		svc := tasks.NewService(tasks.NewStore(cfg.TasksDir))
		p := tea.NewProgram(newBoard(svc, cfg), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			slog.Error("board failed", "err", err)
			os.Exit(1)
		}
	},
}

func init() {
	TaskCmd.AddCommand(boardCmd)
}
