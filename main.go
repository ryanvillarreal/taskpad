package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/rvillarreal/taskpad/internal/app"
	"github.com/rvillarreal/taskpad/internal/calendar"
	"github.com/rvillarreal/taskpad/internal/client"
	"github.com/rvillarreal/taskpad/internal/config"
	"github.com/rvillarreal/taskpad/internal/model"
	"github.com/rvillarreal/taskpad/internal/nlp"
	"github.com/rvillarreal/taskpad/internal/notes"
	"github.com/spf13/cobra"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

var (
	cfg        config.Config
	apiClient  *client.Client
	calService calendar.Service
)

func main() {
	cfg = config.Load()
	apiClient = client.New(cfg.APIURL, cfg.APIKey)
	initCalendar()

	root := &cobra.Command{
		Use:   "taskpad",
		Short: "A terminal-first taskpad for todos, notes, and server mode",
	}

	root.AddCommand(newServerCommand())
	root.AddCommand(newTodoCommands()...)
	root.AddCommand(newTodayCommand())
	root.AddCommand(newNoteCommand())
	root.AddCommand(newConfigCommand())
	root.AddCommand(newCompletionCommand(root))

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func newServerCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Run the taskpad API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			sub, err := fs.Sub(migrationsFS, "migrations")
			if err != nil {
				return err
			}
			return app.RunServer(cfg, sub)
		},
	}
}

func newCompletionCommand(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:       "completion [bash|zsh|fish]",
		Short:     "Generate a shell completion script",
		ValidArgs: []string{"bash", "zsh", "fish"},
		Args:      cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return root.GenBashCompletion(os.Stdout)
			case "zsh":
				return root.GenZshCompletion(os.Stdout)
			case "fish":
				return root.GenFishCompletion(os.Stdout, true)
			default:
				return fmt.Errorf("unsupported shell: %s (use bash, zsh, or fish)", args[0])
			}
		},
	}
}

func newConfigCommand() *cobra.Command {
	configCmd := &cobra.Command{Use: "config", Short: "Inspect config"}
	configCmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Print the config file path",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(config.ConfigPath())
		},
	})
	return configCmd
}

func newTodoCommands() []*cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add [title...]",
		Short: "Add a new todo",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runAdd,
	}
	addCmd.Flags().StringP("urgency", "u", "normal", "Urgency: now, high, normal, low, backburner")
	addCmd.Flags().String("status", "active", "Status: active, paused, done")
	addCmd.Flags().StringSliceP("tag", "t", nil, "Tags (repeatable)")
	addCmd.Flags().StringP("due", "d", "", "Due date (RFC3339, e.g. 2026-04-10T17:00:00Z)")
	addCmd.Flags().Bool("no-sync", false, "Don't sync to calendar")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List todos",
		RunE:  runList,
	}
	listCmd.Flags().BoolP("done", "d", false, "Show only completed todos")
	listCmd.Flags().BoolP("pending", "p", false, "Show only non-completed todos")
	listCmd.Flags().String("status", "", "Filter by status: active, paused, done")
	listCmd.Flags().String("urgency", "", "Filter by urgency: now, high, normal, low, backburner")
	listCmd.Flags().StringP("tag", "t", "", "Filter by tag")
	listCmd.Flags().IntP("limit", "l", 20, "Number of results")

	getCmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get a todo by ID",
		Args:  cobra.ExactArgs(1),
		RunE:  runGet,
	}

	doneCmd := &cobra.Command{
		Use:   "done [id]",
		Short: "Mark a todo as complete",
		Args:  cobra.ExactArgs(1),
		RunE:  runDone,
	}

	undoneCmd := &cobra.Command{
		Use:   "undone [id]",
		Short: "Mark a todo as active again",
		Args:  cobra.ExactArgs(1),
		RunE:  runUndone,
	}

	deleteCmd := &cobra.Command{
		Use:   "rm [id]",
		Short: "Delete a todo",
		Args:  cobra.ExactArgs(1),
		RunE:  runDelete,
	}
	deleteCmd.Flags().Bool("no-sync", false, "Don't remove from calendar")

	return []*cobra.Command{addCmd, listCmd, getCmd, doneCmd, undoneCmd, deleteCmd}
}

func newTodayCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "today",
		Short: "Show active work that matters today",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := apiClient.ListTodos(map[string]string{
				"status": "active",
				"limit":  "200",
				"sort":   "due_date",
				"order":  "asc",
			})
			if err != nil {
				return err
			}

			now := time.Now()
			items := make([]model.Todo, 0)
			for _, todo := range result.Data {
				if includeInToday(todo, now) {
					items = append(items, todo)
				}
			}

			if len(items) == 0 {
				fmt.Println("Nothing urgent for today.")
				return nil
			}

			printTodoTable(items)
			return nil
		},
	}
}

func newNoteCommand() *cobra.Command {
	noteCmd := &cobra.Command{
		Use:   "note",
		Short: "Manage notes",
	}

	noteAddCmd := &cobra.Command{
		Use:   "add [title...]",
		Short: "Add a new note",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runNoteAdd,
	}
	noteAddCmd.Flags().StringP("content", "c", "", "Note content")
	noteAddCmd.Flags().StringSliceP("tag", "t", nil, "Tags (repeatable)")
	noteAddCmd.Flags().Bool("no-sync", false, "Don't write the note to the local Markdown directory")
	noteAddCmd.Flags().String("dir", "", "Markdown notes directory override")

	noteListCmd := &cobra.Command{
		Use:   "list",
		Short: "List notes from the API",
		RunE:  runNoteList,
	}
	noteListCmd.Flags().StringP("tag", "t", "", "Filter by tag")
	noteListCmd.Flags().StringP("search", "s", "", "Search title and content")
	noteListCmd.Flags().IntP("limit", "l", 20, "Number of results")

	noteGetCmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get a note by ID from the API",
		Args:  cobra.ExactArgs(1),
		RunE:  runNoteGet,
	}

	noteDeleteCmd := &cobra.Command{
		Use:   "rm [id]",
		Short: "Delete a note",
		Args:  cobra.ExactArgs(1),
		RunE:  runNoteDelete,
	}
	noteDeleteCmd.Flags().Bool("no-sync", false, "Don't remove the synced local Markdown note")
	noteDeleteCmd.Flags().String("dir", "", "Markdown notes directory override")

	noteSearchCmd := &cobra.Command{
		Use:   "search [query...]",
		Short: "Search local Markdown notes",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runNoteSearch,
	}
	noteSearchCmd.Flags().IntP("limit", "l", 10, "Maximum results")
	noteSearchCmd.Flags().String("dir", "", "Markdown notes directory override")

	noteOpenCmd := &cobra.Command{
		Use:   "open [query...]",
		Short: "Open the best matching local note in $EDITOR",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runNoteOpen,
	}
	noteOpenCmd.Flags().String("dir", "", "Markdown notes directory override")

	noteViewCmd := &cobra.Command{
		Use:   "view [query...]",
		Short: "Print a local note or one heading section to the terminal",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runNoteView,
	}
	noteViewCmd.Flags().String("dir", "", "Markdown notes directory override")
	noteViewCmd.Flags().String("heading", "", "Heading to extract from the note")

	noteDayCmd := &cobra.Command{
		Use:   "day [date]",
		Short: "Open a daily note (today by default); syncs with the API",
		RunE:  runNoteDay,
	}
	noteDayCmd.Flags().Bool("view", false, "Print to terminal without opening editor")
	noteDayCmd.Flags().Bool("pull", false, "Pull from API to local disk only")
	noteDayCmd.Flags().Bool("push", false, "Push local file to API only")
	noteDayCmd.Flags().Bool("no-sync", false, "Work locally without API sync")
	noteDayCmd.Flags().String("dir", "", "Markdown notes directory override")

	noteCmd.AddCommand(noteAddCmd, noteListCmd, noteGetCmd, noteDeleteCmd, noteSearchCmd, noteOpenCmd, noteViewCmd, noteDayCmd)
	return noteCmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	title := strings.Join(args, " ")
	urgency, _ := cmd.Flags().GetString("urgency")
	status, _ := cmd.Flags().GetString("status")
	tags, _ := cmd.Flags().GetStringSlice("tag")
	due, _ := cmd.Flags().GetString("due")
	noSync, _ := cmd.Flags().GetBool("no-sync")

	var parsedDue *time.Time
	if due == "" {
		if result := nlp.ExtractDate(title, time.Now()); result != nil {
			parsedDue = &result.Time
			title = nlp.StripDate(title, result)
			due = result.Time.Format(time.RFC3339)
			fmt.Printf("Detected date: %s\n", result.Time.Format("Mon Jan 2, 2006 3:04 PM"))
		}
	} else if t, err := time.Parse(time.RFC3339, due); err == nil {
		parsedDue = &t
	}

	todo, err := apiClient.CreateTodo(model.CreateTodoRequest{
		Title:   title,
		Status:  model.TodoStatus(status),
		Urgency: model.TodoUrgency(urgency),
		Tags:    tags,
		DueDate: optionalString(due),
	})
	if err != nil {
		return err
	}

	fmt.Printf("Created: %s\n  ID: %s\n", todo.Title, todo.ID)

	if !noSync && calService != nil && parsedDue != nil {
		eventID, err := calService.CreateEvent(context.Background(), calendar.Event{
			Title:   todo.Title,
			DueDate: *parsedDue,
		})
		if err != nil {
			fmt.Printf("  Warning: calendar sync failed: %v\n", err)
		} else {
			_, updateErr := apiClient.UpdateTodo(todo.ID, model.UpdateTodoRequest{
				CalendarEventID: &eventID,
			})
			if updateErr != nil {
				fmt.Printf("  Warning: failed to persist calendar event ID: %v\n", updateErr)
			}
			fmt.Printf("  Synced to calendar\n")
		}
	}

	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	params := map[string]string{}
	if done, _ := cmd.Flags().GetBool("done"); done {
		params["completed"] = "true"
	}
	if pending, _ := cmd.Flags().GetBool("pending"); pending {
		params["completed"] = "false"
	}
	if status, _ := cmd.Flags().GetString("status"); status != "" {
		params["status"] = status
	}
	if urgency, _ := cmd.Flags().GetString("urgency"); urgency != "" {
		params["urgency"] = urgency
	}
	if tag, _ := cmd.Flags().GetString("tag"); tag != "" {
		params["tag"] = tag
	}
	if l, _ := cmd.Flags().GetInt("limit"); l > 0 {
		params["limit"] = fmt.Sprintf("%d", l)
	}

	result, err := apiClient.ListTodos(params)
	if err != nil {
		return err
	}
	if len(result.Data) == 0 {
		fmt.Println("No todos found.")
		return nil
	}
	printTodoTable(result.Data)
	fmt.Printf("\nShowing %d of %d todos\n", len(result.Data), result.Total)
	return nil
}

func runGet(cmd *cobra.Command, args []string) error {
	todo, err := apiClient.GetTodo(args[0])
	if err != nil {
		return err
	}
	printTodo(todo)
	return nil
}

func runDone(cmd *cobra.Command, args []string) error {
	todo, err := apiClient.CompleteTodo(args[0], true)
	if err != nil {
		return err
	}
	fmt.Printf("Completed: %s\n", todo.Title)
	return nil
}

func runUndone(cmd *cobra.Command, args []string) error {
	todo, err := apiClient.CompleteTodo(args[0], false)
	if err != nil {
		return err
	}
	fmt.Printf("Reopened: %s\n", todo.Title)
	return nil
}

func runDelete(cmd *cobra.Command, args []string) error {
	noSync, _ := cmd.Flags().GetBool("no-sync")
	if !noSync && calService != nil {
		todo, err := apiClient.GetTodo(args[0])
		if err == nil && todo.CalendarEventID != "" {
			if err := calService.DeleteEvent(context.Background(), todo.CalendarEventID); err != nil {
				fmt.Printf("Warning: failed to remove calendar event: %v\n", err)
			}
		}
	}
	if err := apiClient.DeleteTodo(args[0]); err != nil {
		return err
	}
	fmt.Println("Deleted.")
	return nil
}

func runNoteAdd(cmd *cobra.Command, args []string) error {
	title := strings.Join(args, " ")
	content, _ := cmd.Flags().GetString("content")
	tags, _ := cmd.Flags().GetStringSlice("tag")
	noSync, _ := cmd.Flags().GetBool("no-sync")
	if tags == nil {
		tags = []string{}
	}

	note, err := apiClient.CreateNote(model.CreateNoteRequest{
		Title:   title,
		Content: content,
		Tags:    tags,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Created note: %s\n  ID: %s\n", note.Title, note.ID)

	if !noSync {
		writer, err := resolveNoteWriter(cmd)
		if err != nil {
			fmt.Printf("  Warning: local note sync unavailable: %v\n", err)
		} else if writer != nil {
			path, err := writer.Write(note)
			if err != nil {
				fmt.Printf("  Warning: failed to write Markdown note: %v\n", err)
			} else {
				fmt.Printf("  Saved to: %s\n", path)
			}
		}
	}
	return nil
}

func runNoteList(cmd *cobra.Command, args []string) error {
	params := map[string]string{}
	if t, _ := cmd.Flags().GetString("tag"); t != "" {
		params["tag"] = t
	}
	if s, _ := cmd.Flags().GetString("search"); s != "" {
		params["search"] = s
	}
	if l, _ := cmd.Flags().GetInt("limit"); l > 0 {
		params["limit"] = fmt.Sprintf("%d", l)
	}

	result, err := apiClient.ListNotes(params)
	if err != nil {
		return err
	}
	if len(result.Data) == 0 {
		fmt.Println("No notes found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "TITLE\tTAGS\tID\n")
	for _, n := range result.Data {
		tags := "-"
		if len(n.Tags) > 0 {
			tags = strings.Join(n.Tags, ", ")
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", n.Title, tags, shortID(n.ID))
	}
	w.Flush()
	fmt.Printf("\nShowing %d of %d notes\n", len(result.Data), result.Total)
	return nil
}

func runNoteGet(cmd *cobra.Command, args []string) error {
	note, err := apiClient.GetNote(args[0])
	if err != nil {
		return err
	}
	printNote(note)
	return nil
}

func runNoteDelete(cmd *cobra.Command, args []string) error {
	noSync, _ := cmd.Flags().GetBool("no-sync")
	if err := apiClient.DeleteNote(args[0]); err != nil {
		return err
	}

	if !noSync {
		writer, err := resolveNoteWriter(cmd)
		if err != nil {
			fmt.Printf("Warning: local note sync unavailable: %v\n", err)
		} else if writer != nil {
			path, err := writer.DeleteByNoteID(args[0])
			if err != nil {
				fmt.Printf("Warning: failed to remove Markdown note: %v\n", err)
			} else if path != "" {
				fmt.Printf("Removed local note: %s\n", path)
			}
		}
	}
	fmt.Println("Deleted.")
	return nil
}

func runNoteSearch(cmd *cobra.Command, args []string) error {
	dir, err := resolveNotesDir(cmd)
	if err != nil {
		return err
	}
	query := strings.Join(args, " ")
	limit, _ := cmd.Flags().GetInt("limit")
	results, err := notes.Search(dir, query, limit)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		fmt.Println("No local note matches found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "TITLE\tSNIPPET\tPATH\n")
	for _, result := range results {
		fmt.Fprintf(w, "%s\t%s\t%s\n", result.Title, result.Snippet, result.Path)
	}
	w.Flush()
	return nil
}

func runNoteOpen(cmd *cobra.Command, args []string) error {
	dir, err := resolveNotesDir(cmd)
	if err != nil {
		return err
	}
	results, err := notes.Search(dir, strings.Join(args, " "), 1)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		return fmt.Errorf("no local note matches found")
	}
	if strings.TrimSpace(cfg.Editor) == "" {
		return fmt.Errorf("$EDITOR is not set")
	}
	return notes.OpenInEditor(cfg.Editor, results[0].Path)
}

func runNoteView(cmd *cobra.Command, args []string) error {
	dir, err := resolveNotesDir(cmd)
	if err != nil {
		return err
	}
	results, err := notes.Search(dir, strings.Join(args, " "), 1)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		return fmt.Errorf("no local note matches found")
	}
	heading, _ := cmd.Flags().GetString("heading")
	text, err := notes.View(results[0].Path, heading)
	if err != nil {
		return err
	}
	fmt.Print(text)
	return nil
}

func runNoteDay(cmd *cobra.Command, args []string) error {
	dir, err := resolveNotesDir(cmd)
	if err != nil {
		return err
	}

	date, err := resolveDate(args)
	if err != nil {
		return err
	}

	noSync, _ := cmd.Flags().GetBool("no-sync")
	viewOnly, _ := cmd.Flags().GetBool("view")
	pullOnly, _ := cmd.Flags().GetBool("pull")
	pushOnly, _ := cmd.Flags().GetBool("push")

	// -- push only: read local file, push body to API --
	if pushOnly {
		path := notes.DailyNoteFilePath(dir, date)
		body, err := notes.ParseDailyNoteBody(path)
		if err != nil {
			return fmt.Errorf("read local note: %w", err)
		}
		if _, err := apiClient.UpsertDailyNote(date, body); err != nil {
			return fmt.Errorf("push to API: %w", err)
		}
		// Refresh edited timestamp on disk.
		if _, err := notes.WriteLocalDailyNote(dir, date, body); err != nil {
			fmt.Printf("Warning: failed to update local file: %v\n", err)
		}
		fmt.Printf("Pushed daily note for %s\n", date)
		return nil
	}

	// -- pull from API (unless no-sync) --
	if !noSync {
		n, err := apiClient.GetDailyNote(date)
		if err != nil && !client.IsNotFound(err) {
			return fmt.Errorf("pull from API: %w", err)
		}
		content := ""
		if n != nil {
			content = n.Content
		}
		if _, err := notes.WriteLocalDailyNote(dir, date, content); err != nil {
			fmt.Printf("Warning: failed to write local note: %v\n", err)
		}
		if pullOnly {
			fmt.Printf("Pulled daily note for %s\n", date)
			return nil
		}
	} else {
		// Ensure the file exists locally even without sync.
		path := notes.DailyNoteFilePath(dir, date)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if _, err := notes.WriteLocalDailyNote(dir, date, ""); err != nil {
				return fmt.Errorf("create local note: %w", err)
			}
		}
	}

	path := notes.DailyNoteFilePath(dir, date)

	// -- view only: print to terminal --
	if viewOnly {
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read local note: %w", err)
		}
		fmt.Print(string(raw))
		return nil
	}

	// -- open in editor --
	if strings.TrimSpace(cfg.Editor) == "" {
		return fmt.Errorf("$EDITOR is not set")
	}
	if err := notes.OpenInEditor(cfg.Editor, path); err != nil {
		return err
	}

	// -- push after editor closes (unless no-sync) --
	if !noSync {
		body, err := notes.ParseDailyNoteBody(path)
		if err != nil {
			return fmt.Errorf("read note after edit: %w", err)
		}
		if _, err := apiClient.UpsertDailyNote(date, body); err != nil {
			fmt.Printf("Warning: failed to push to API: %v\n", err)
		} else {
			// Refresh edited timestamp.
			_, _ = notes.WriteLocalDailyNote(dir, date, body)
			fmt.Printf("Synced to API.\n")
		}
	}
	return nil
}

// resolveDate parses a date from CLI args. No args returns today.
// Accepts YYYY-MM-DD or natural language (yesterday, last week, etc.).
func resolveDate(args []string) (string, error) {
	now := time.Now()
	if len(args) == 0 {
		return now.Local().Format("2006-01-02"), nil
	}
	input := strings.Join(args, " ")

	// ISO date shortcut.
	if _, err := time.Parse("2006-01-02", input); err == nil {
		return input, nil
	}

	// Natural language via NLP.
	if result := nlp.ExtractDate(input, now); result != nil {
		return result.Time.Local().Format("2006-01-02"), nil
	}

	return "", fmt.Errorf("unrecognised date %q — use YYYY-MM-DD, 'yesterday', or 'last week'", input)
}

func printTodoTable(todos []model.Todo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "STATUS\tURGENCY\tDUE\tTAGS\tTITLE\tID\n")
	for _, t := range todos {
		tags := "-"
		if len(t.Tags) > 0 {
			tags = strings.Join(t.Tags, ",")
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", t.Status, t.Urgency, formatDueDate(t.DueDate), tags, t.Title, shortID(t.ID))
	}
	w.Flush()
}

func printTodo(t *model.Todo) {
	fmt.Printf("Title:       %s\n", t.Title)
	fmt.Printf("ID:          %s\n", t.ID)
	fmt.Printf("Status:      %s\n", t.Status)
	fmt.Printf("Urgency:     %s\n", t.Urgency)
	if len(t.Tags) > 0 {
		fmt.Printf("Tags:        %s\n", strings.Join(t.Tags, ", "))
	}
	if t.Description != "" {
		fmt.Printf("Description: %s\n", t.Description)
	}
	if t.DueDate != nil {
		fmt.Printf("Due:         %s\n", t.DueDate.Format(time.RFC3339))
	}
	fmt.Printf("Created:     %s\n", t.CreatedAt.Format(time.RFC3339))
}

func printNote(n *model.Note) {
	fmt.Printf("Title:   %s\n", n.Title)
	fmt.Printf("ID:      %s\n", n.ID)
	if len(n.Tags) > 0 {
		fmt.Printf("Tags:    %s\n", strings.Join(n.Tags, ", "))
	}
	fmt.Printf("Created: %s\n", n.CreatedAt.Format(time.RFC3339))
	if n.Content != "" {
		fmt.Printf("\n%s\n", n.Content)
	}
}

func formatDueDate(due *time.Time) string {
	if due == nil {
		return "-"
	}
	return due.Local().Format("2006-01-02 15:04")
}

func shortID(id string) string {
	if len(id) >= 8 {
		return id[:8]
	}
	return id
}

func resolveNoteWriter(cmd *cobra.Command) (notes.Writer, error) {
	dir, err := resolveNotesDir(cmd)
	if err != nil {
		return nil, err
	}
	return notes.NewMarkdownWriter(dir), nil
}

func resolveNotesDir(cmd *cobra.Command) (string, error) {
	if cmd != nil {
		if value, err := cmd.Flags().GetString("dir"); err == nil && value != "" {
			return value, nil
		}
	}
	if strings.TrimSpace(cfg.NotesDir) == "" {
		return "", fmt.Errorf("notes directory is not configured")
	}
	return cfg.NotesDir, nil
}

func includeInToday(todo model.Todo, now time.Time) bool {
	if todo.Status != model.TodoStatusActive {
		return false
	}
	if todo.Urgency == model.TodoUrgencyBackburner {
		return false
	}
	if todo.DueDate != nil {
		dueDay := todo.DueDate.Local().Format("2006-01-02")
		nowDay := now.Local().Format("2006-01-02")
		if dueDay <= nowDay {
			return true
		}
	}
	return todo.Urgency == model.TodoUrgencyNow || todo.Urgency == model.TodoUrgencyHigh
}

func optionalString(v string) *string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return &v
}

func initCalendar() {
	if strings.TrimSpace(cfg.CalDAV.URL) == "" {
		return
	}
	service, err := calendar.NewCalDAV(calendar.CalDAVConfig{
		ServerURL:    cfg.CalDAV.URL,
		Username:     cfg.CalDAV.Username,
		Password:     cfg.CalDAV.Password,
		CalendarPath: cfg.CalDAV.CalendarPath,
	})
	if err == nil {
		calService = service
	}
}
