package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/rvillarreal/taskpad/internal/calendar"
	"github.com/rvillarreal/taskpad/internal/client"
	"github.com/rvillarreal/taskpad/internal/model"
	"github.com/rvillarreal/taskpad/internal/nlp"
	"github.com/spf13/cobra"
)

var (
	apiClient *client.Client
	calService calendar.Service
)

func main() {
	serverURL := os.Getenv("TASKPAD_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}
	apiClient = client.New(serverURL)

	// Initialize CalDAV if configured.
	initCalendar()

	root := &cobra.Command{
		Use:   "taskpad",
		Short: "A CLI for managing todos and notes",
	}

	// --- Todo commands ---
	addCmd := &cobra.Command{
		Use:   "add [title...]",
		Short: "Add a new todo",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runAdd,
	}
	addCmd.Flags().StringP("priority", "p", "medium", "Priority: low, medium, high")
	addCmd.Flags().StringP("due", "d", "", "Due date (RFC3339, e.g. 2026-04-10T17:00:00Z)")
	addCmd.Flags().Bool("no-sync", false, "Don't sync to calendar")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List todos",
		RunE:  runList,
	}
	listCmd.Flags().BoolP("done", "d", false, "Show only completed todos")
	listCmd.Flags().BoolP("pending", "p", false, "Show only pending todos")
	listCmd.Flags().String("priority", "", "Filter by priority: low, medium, high")
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
		Short: "Mark a todo as incomplete",
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

	// --- Note commands ---
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

	noteListCmd := &cobra.Command{
		Use:   "list",
		Short: "List notes",
		RunE:  runNoteList,
	}
	noteListCmd.Flags().StringP("tag", "t", "", "Filter by tag")
	noteListCmd.Flags().StringP("search", "s", "", "Search title and content")
	noteListCmd.Flags().IntP("limit", "l", 20, "Number of results")

	noteGetCmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get a note by ID",
		Args:  cobra.ExactArgs(1),
		RunE:  runNoteGet,
	}

	noteDeleteCmd := &cobra.Command{
		Use:   "rm [id]",
		Short: "Delete a note",
		Args:  cobra.ExactArgs(1),
		RunE:  runNoteDelete,
	}

	noteCmd.AddCommand(noteAddCmd, noteListCmd, noteGetCmd, noteDeleteCmd)
	root.AddCommand(addCmd, listCmd, getCmd, doneCmd, undoneCmd, deleteCmd, noteCmd)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// --- Todo handlers ---

func runAdd(cmd *cobra.Command, args []string) error {
	title := strings.Join(args, " ")
	priority, _ := cmd.Flags().GetString("priority")
	due, _ := cmd.Flags().GetString("due")
	noSync, _ := cmd.Flags().GetBool("no-sync")

	// Try to extract a date from the title via NLP if no explicit --due flag.
	var parsedDue *time.Time
	if due == "" {
		if result := nlp.ExtractDate(title, time.Now()); result != nil {
			parsedDue = &result.Time
			title = nlp.StripDate(title, result)
			dueStr := result.Time.Format(time.RFC3339)
			due = dueStr
			fmt.Printf("Detected date: %s\n", result.Time.Format("Mon Jan 2, 2006 3:04 PM"))
		}
	} else {
		t, err := time.Parse(time.RFC3339, due)
		if err == nil {
			parsedDue = &t
		}
	}

	req := model.CreateTodoRequest{
		Title:    title,
		Priority: model.Priority(priority),
	}
	if due != "" {
		req.DueDate = &due
	}

	todo, err := apiClient.CreateTodo(req)
	if err != nil {
		return err
	}
	fmt.Printf("Created: %s\n  ID: %s\n", todo.Title, todo.ID)

	// Sync to calendar if enabled and there's a due date.
	if !noSync && calService != nil && parsedDue != nil {
		eventID, err := calService.CreateEvent(context.Background(), calendar.Event{
			Title:   todo.Title,
			DueDate: *parsedDue,
		})
		if err != nil {
			fmt.Printf("  Warning: calendar sync failed: %v\n", err)
		} else {
			// Update todo with the calendar event ID.
			todo.CalendarEventID = eventID
			apiClient.UpdateTodo(todo.ID, model.UpdateTodoRequest{})
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
	if p, _ := cmd.Flags().GetString("priority"); p != "" {
		params["priority"] = p
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

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "STATUS\tPRIORITY\tTITLE\tID\n")
	for _, t := range result.Data {
		status := "[ ]"
		if t.Completed {
			status = "[x]"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", status, t.Priority, t.Title, shortID(t.ID))
	}
	w.Flush()
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

	// If calendar sync is enabled, fetch the todo first to get the event ID.
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

// --- Note handlers ---

func runNoteAdd(cmd *cobra.Command, args []string) error {
	title := strings.Join(args, " ")
	content, _ := cmd.Flags().GetString("content")
	tags, _ := cmd.Flags().GetStringSlice("tag")

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
	if err := apiClient.DeleteNote(args[0]); err != nil {
		return err
	}
	fmt.Println("Deleted.")
	return nil
}

// --- Display helpers ---

func printTodo(t *model.Todo) {
	status := "pending"
	if t.Completed {
		status = "completed"
	}
	fmt.Printf("Title:       %s\n", t.Title)
	fmt.Printf("ID:          %s\n", t.ID)
	fmt.Printf("Status:      %s\n", status)
	fmt.Printf("Priority:    %s\n", t.Priority)
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

func shortID(id string) string {
	if len(id) >= 8 {
		return id[:8]
	}
	return id
}

// --- Calendar setup ---

func initCalendar() {
	serverURL := os.Getenv("TASKPAD_CALDAV_URL")
	if serverURL == "" {
		return // CalDAV not configured, sync disabled.
	}

	cfg := calendar.CalDAVConfig{
		ServerURL:    serverURL,
		Username:     os.Getenv("TASKPAD_CALDAV_USER"),
		Password:     os.Getenv("TASKPAD_CALDAV_PASS"),
		CalendarPath: os.Getenv("TASKPAD_CALDAV_CALENDAR"),
	}

	var err error
	calService, err = calendar.NewCalDAV(cfg)
	if err != nil {
		log.Printf("Warning: CalDAV init failed: %v (calendar sync disabled)", err)
		calService = nil
	}
}
