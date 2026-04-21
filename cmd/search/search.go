package searchcmd

import (
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ryanvillarreal/taskpad/internal/client"
	"github.com/ryanvillarreal/taskpad/internal/config"
	"github.com/ryanvillarreal/taskpad/internal/links"
	"github.com/ryanvillarreal/taskpad/internal/notes"
	"github.com/ryanvillarreal/taskpad/internal/search"
	"github.com/ryanvillarreal/taskpad/internal/tasks"
	"github.com/spf13/cobra"
)

var exact bool

var SearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "search notes, links, and tasks",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		notesSvc := notes.NewService(notes.NewStore(cfg.NotesDir))
		linksSvc := links.NewService(links.NewStore(cfg.LinksDir), cfg.FetchLimit)
		tasksSvc := tasks.NewService(tasks.NewStore(cfg.TasksDir))

		results, err := search.New(notesSvc, linksSvc, tasksSvc).Query(args[0], exact)
		if err != nil {
			slog.Error("search failed", "err", err)
			os.Exit(1)
		}

		if !isTerminal() {
			for _, r := range results {
				fmt.Printf("[%-4s]  %-14s  %s\n", r.Kind, r.ID, r.Snippet)
			}
			return
		}

		p := tea.NewProgram(newPicker(results, args[0]), tea.WithAltScreen())
		final, err := p.Run()
		if err != nil {
			slog.Error("picker failed", "err", err)
			os.Exit(1)
		}

		picked := final.(pickerModel).selected
		if picked == nil {
			return
		}

		switch picked.Kind {
		case search.KindNote:
			if err := client.EditNote(picked.ID, false); err != nil {
				slog.Error("open note failed", "err", err)
				os.Exit(1)
			}
		case search.KindLink:
			l, err := linksSvc.Get(picked.ID)
			if err != nil {
				slog.Error("fetch link failed", "err", err)
				os.Exit(1)
			}
			fmt.Println(l.URL)
		case search.KindTask:
			t, err := tasksSvc.Get(picked.ID)
			if err != nil {
				slog.Error("fetch task failed", "err", err)
				os.Exit(1)
			}
			printTask(t)
		}
	},
}

func printTask(t *tasks.Task) {
	fmt.Printf("id:      %s\n", t.ID)
	fmt.Printf("title:   %s\n", t.Title)
	fmt.Printf("status:  %s\n", t.Status)
	fmt.Printf("created: %s\n", t.CreatedAt.Local().Format("2006-01-02 15:04"))
	if !t.DueAt.IsZero() {
		fmt.Printf("due:     %s\n", t.DueAt.Local().Format("2006-01-02 15:04"))
	}
	if !t.ClosedAt.IsZero() {
		fmt.Printf("closed:  %s\n", t.ClosedAt.Local().Format("2006-01-02 15:04"))
	}
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func init() {
	SearchCmd.Flags().BoolVar(&exact, "exact", false, "case-sensitive matching")
}
