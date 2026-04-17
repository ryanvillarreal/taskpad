# taskpad

initially built using AI I have decided to do it myself properly.

Why? I need my own solution for customized note taking, tracking, etc. I want it
to be self-hostable. I want it to be secure and and private. I don't trust third-party sources.

### Wishlist

1. add a note from cmdline
2. add a todo from cmdline
3. android app extension for mobile support
4. single source of truth (the webserver)
5. allow for easy backups of notes
6. streamlined parsing of web sources

### examples?

Install
`go install github.com/ryanvillarreal/taskpad`

Auth

```
./taskpad auth login -- idk yet
./taskpad auth api {key} -- adds an api key or generates an api key for .env
```

Configure

```
./taskpad config list
./taskpad config edit
./taskpad config default
```

Notes

```
./taskpad note add -c "quick short note" - add text quickly from cmdline
./taskpad note add -- open editor in note mode - longer format
./taskpad note rm {id} - remove it by id
./taskpad note search -c "grep-ish"
./taskpad note view {id} - view the note by id
./taskpad note find -c "grep-ish"
```

Todos

```
./taskpad todo add -c 'quick todo'
./taskpad todo add -- opens todo in editor mode for longer form todos
./taskpad todo today -- list of action items either due today or past due
./taskpad todo list -- view all todos
./taskpad todo add -c 'thing to do on {date,time,day,month,year etc}'
```
