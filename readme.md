# taskpad

Why? I need my own solution for customized note taking, tracking, etc. I want it
to be self-hostable. I want it to be secure and and private. I don't trust third-party sources.

### Wishlist

1. add a note from cmdline
2. add a todo from cmdline
3. android app extension for mobile support
4. single source of truth (the webserver)
5. allow for easy backups of notes
6. streamlined parsing of web sources

### Install

```
go install github.com/ryanvillarreal/taskpad@latest
```

### How it works

Notes are plain markdown files on disk — one per day, named `MM.DD.YYYY.md`. Each
file carries YAML frontmatter (id, timestamps, tags) so it stays compatible with
Obsidian and other markdown tools.

The binary is both client and server. Run `taskpad serve` in one terminal; use
`taskpad note`, `taskpad ls`, `taskpad rm`, and `taskpad config` from another.
If the server isn't running, client commands fall back to direct filesystem
access automatically.

### Commands

Global flags: `-v` / `--verbose` to enable debug logging.

**Serve**

```
./taskpad serve
```

Starts the HTTP API. Port, host, TLS cert paths, and notes directory are all
read from the config file.

**Notes**

```
./taskpad note                  # edit today's note in $EDITOR
./taskpad note 4.16.2026        # edit a specific note (errors if missing)
./taskpad note 4.16.26          # short-year form works too
./taskpad note 01.01.2030 --new # create/edit a note at that date
./taskpad ls                    # list all notes, newest first
./taskpad rm 4.16.2026          # delete a note
```

Editor resolution: `$VISUAL` → `$EDITOR` → `vi`.

**Config**

```
./taskpad config                # open the config file in $EDITOR
```

Config lives at `$XDG_CONFIG_HOME/taskpad/config.json` (typically
`~/.config/taskpad/config.json`). Notes default to
`$XDG_DATA_HOME/taskpad/notes/` (typically `~/.local/share/taskpad/notes/`).

### API

```
GET    /health
GET    /notes                  # { "count": N }
GET    /notes/{id}             # raw markdown (full file, frontmatter included)
POST   /notes/{id}             # body is raw markdown; creates or merges
DELETE /notes/{id}             # 204 on success, 404 if missing
```

Example:

```
curl http://localhost:8080/notes/04.16.2026
curl -X POST http://localhost:8080/notes/04.16.2026 \
    --data-binary @today.md
curl -X DELETE http://localhost:8080/notes/04.16.2026
```

### Planned

```
./taskpad auth login
./taskpad auth api {key}
./taskpad todo add -c 'quick todo'
./taskpad todo today
./taskpad todo list
./taskpad note search -c "grep-ish"
```
