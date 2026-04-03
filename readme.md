# taskpad

Taskpad is a terminal-first productivity tool built in Go. It started as a todo app because the first project everyone seems to build is a todo app, and that makes it a useful way to test the limits of AI on something familiar but still full of real workflow tradeoffs.

The project now includes:
- a single `taskpad` binary
- `taskpad server` for the local API
- todo management with status, urgency, tags, and due dates
- note capture backed by the API and synced to Markdown on disk
- local note search, note viewing, and editor-based note opening
- optional CalDAV sync for date-based tasks

## why this exists

This is a testbed for pushing on what AI can and cannot do well in a real developer workflow:
- evolving architecture over time
- terminal-first UX
- local files plus API-backed state
- calendar integration
- search and retrieval of notes from the command line

Using a todo app on purpose keeps the problem understandable while making it easy to spot where the tool gets awkward, where abstractions break down, and where AI-generated code needs stronger product judgment.

## install

**Prerequisites:** Go 1.21 or later.

```bash
git clone https://github.com/rvillarreal/taskpad
cd taskpad
go build -o taskpad .
```

Move it somewhere on your `$PATH` if you want to use it without a prefix:

```bash
mv taskpad /usr/local/bin/taskpad
```

The server needs a running SQLite database. The default path is `./taskpad.db` in whichever directory you launch the server from. Override this in the config file or with the `DB_PATH` environment variable.

## quick start

Start the API server (keep this running in a separate terminal or as a background process):

```bash
taskpad server
```

Add and inspect tasks:

```bash
taskpad add "call dentist on friday" --urgency high --tag health
taskpad list
taskpad today
```

Work with notes:

```bash
taskpad note add "bettercap docs" -c "remember the cmdline options"
taskpad note search "bettercap"
taskpad note view "bettercap" --heading "cmdline options"
```

## config

Config is loaded from a JSON file, then overridden by environment variables. The config file location:

- **macOS:** `~/Library/Application Support/taskpad/config.json`
- **Linux:** `~/.config/taskpad/config.json`

Print the resolved path at any time:

```bash
taskpad config path
```

### config file

```json
{
  "api_url": "http://localhost:8080",
  "notes_dir": "/Users/you/notes",
  "editor": "nvim",
  "migrations_dir": "./migrations",
  "server": {
    "port": "8080",
    "db_path": "./taskpad.db"
  },
  "caldav": {
    "url": "https://caldav.example.com",
    "username": "you",
    "password": "secret",
    "calendar_path": "/calendars/you/default/"
  }
}
```

All fields are optional. Unset fields fall back to defaults or environment variables.

### environment variables

| Variable | Config field |
|---|---|
| `TASKPAD_URL` | `api_url` |
| `TASKPAD_NOTES_DIR` | `notes_dir` |
| `TASKPAD_EDITOR` | `editor` (falls back to `$EDITOR`) |
| `MIGRATIONS_DIR` | `migrations_dir` |
| `PORT` | `server.port` |
| `DB_PATH` | `server.db_path` |
| `TASKPAD_CALDAV_URL` | `caldav.url` |
| `TASKPAD_CALDAV_USER` | `caldav.username` |
| `TASKPAD_CALDAV_PASS` | `caldav.password` |
| `TASKPAD_CALDAV_CALENDAR` | `caldav.calendar_path` |

## command reference

### server

```
taskpad server
```

Starts the local REST API on the configured port (default `8080`).

### todos

```
taskpad add [title...]
  --urgency, -u   now | high | normal | low | backburner  (default: normal)
  --status        active | paused | done                  (default: active)
  --tag, -t       tag name, repeatable
  --due, -d       due date in natural language or RFC3339
  --no-sync       skip CalDAV sync

taskpad list
  --status        filter by status
  --urgency       filter by urgency
  --tag, -t       filter by tag
  --done, -d      show only completed todos
  --pending, -p   show only non-completed todos
  --limit, -l     max results (default: 20)

taskpad get [id]
taskpad done [id]
taskpad undone [id]
taskpad rm [id]
  --no-sync       skip CalDAV removal

taskpad today
```

`today` shows active todos that are due today or earlier, plus anything marked `now` or `high` urgency. Backburnered items are always excluded.

Due dates in `add` accept natural language: `on friday`, `tomorrow`, `next monday`. The parsed date is confirmed in the output before saving.

### notes

```
taskpad note add [title...]
  --content, -c   note body text
  --tag, -t       tag name, repeatable
  --no-sync       skip writing to the local Markdown directory
  --dir           override the notes directory for this command

taskpad note list
  --tag, -t       filter by tag
  --search, -s    search title and content
  --limit, -l     max results (default: 20)

taskpad note get [id]
taskpad note rm [id]
  --no-sync       skip removing the local Markdown file
  --dir           override the notes directory for this command

taskpad note search [query...]
  --limit, -l     max results (default: 10)
  --dir           override the notes directory for this command

taskpad note open [query...]
  --dir           override the notes directory for this command

taskpad note view [query...]
  --heading       extract a specific heading section
  --dir           override the notes directory for this command
```

`note search`, `note open`, and `note view` operate on local Markdown files in `notes_dir`. `note list` and `note get` query the API.

`note open` requires `$EDITOR` or `editor` in the config.

### misc

```
taskpad config path   print the config file path
taskpad completion [bash|zsh|fish]   print a shell completion script
```

## shell completions

Generate and install a completion script for your shell.

**zsh:**

```bash
taskpad completion zsh > "${fpath[1]}/_taskpad"
```

Or source it directly in `~/.zshrc`:

```bash
source <(taskpad completion zsh)
```

**bash:**

```bash
taskpad completion bash > /etc/bash_completion.d/taskpad
# or, for a single user:
taskpad completion bash > ~/.local/share/bash-completion/completions/taskpad
```

**fish:**

```bash
taskpad completion fish > ~/.config/fish/completions/taskpad.fish
```

## current direction

The focus right now is staying in the terminal as much as possible:
- better task workflows like `today`
- better note retrieval and viewing
- local Markdown interoperability with tools like Obsidian
- practical automation without overcomplicating the model
