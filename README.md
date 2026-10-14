<div align="center">

# tasked

**Manage Google Tasks from your terminal with priorities.**

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Google Tasks](https://img.shields.io/badge/Google%20Tasks-API-4285F4?logo=google&logoColor=white)](https://developers.google.com/tasks)
[![Bubble Tea](https://img.shields.io/badge/TUI-Bubble%20Tea-FF75B7)](https://github.com/charmbracelet/bubbletea)
[![Last commit](https://img.shields.io/github/last-commit/n3tw0rth/tasked)](https://github.com/n3tw0rth/tasked/commits)

`cli` · `google-tasks` · `tui` · `productivity` · `todo`

</div>

Google Tasks has no priority field, so `tasked` adds one: each task carries a
`[p1]`–`[p5]` tag in its notes, and everything (listing, sorting, chips in the
UI) is built around it. Tasks stay perfectly usable from the official Google
Tasks apps; the tag just travels along in the notes.

- **Inline TUI**: quick pickers and forms (Bubble Tea), no full-screen takeover
- **Priorities**: p1 (highest) to p5 (lowest), sorted first in every view
- **Profiles**: multiple Google accounts, switch with one command
- **Scriptable**: every picker can be skipped with arguments and flags, and
  `ls`/`search` print `--json`

## Install

No Google Cloud setup needed: install, then `tasked login`.

### From source (Go 1.26+)

```sh
git clone https://github.com/n3tw0rth/tasked
cd tasked
just install           # or: just build
```

or `go install github.com/n3tw0rth/tasked@latest`.

### Nix / NixOS

```sh
nix run github:n3tw0rth/tasked            # try it
nix profile install github:n3tw0rth/tasked
```

A dev shell with Go, gopls, and just is available via `nix develop`.

## Quick start

```sh
tasked login                # sign in with Google (opens browser)
tasked add                  # create a task: title, notes, due date, priority
tasked ls                   # tasks sorted by priority, then due date
tasked done                 # mark tasks completed (multi-select)
tasked lists switch         # optional: work in a list other than "My Tasks"
```

Until you pick a list with `lists switch`, commands use your account's
default list.

## Commands

| Command | Description |
| --- | --- |
| `add [title]` | Create a task: inline form, or directly with `--due`, `-p`, `--notes` |
| `ls` | List tasks sorted by priority (`--all`, `-p p1..p5`, `--json`) |
| `edit [task]` | Change title, notes, due date (`--no-due` to clear), or priority |
| `done [task...]` | Mark tasks completed |
| `reopen [task...]` | Mark completed tasks as not done |
| `move [p1-p5 \| list] [task...]` | Set a priority or move tasks to another list |
| `rm [task...]` | Delete tasks (`--yes` skips the prompt) |
| `search <query>` | Search tasks by title/notes (`--json`) |
| `lists` | Manage task lists: `ls`, `create`, `switch [list]`, `rm [list]` |
| `profile` | Manage accounts: `list`, `add`, `use`, `remove` |
| `login` / `logout` | Sign in to Google / remove the saved token |
| `auth status` | Show which OAuth client is used to sign in |

`ls` and `search` show subtasks nested under their parent task (each level
sorted by priority). Task commands open an inline picker when no tasks are
given. A task argument
can be a task ID, an exact title, or a unique part of a title
(case-insensitive); ambiguous matches are errors, never guesses.

Every command accepts `--profile <name>` to target a specific account, and
task commands accept `--list <name or ID>` to work in another list, without
changing what is active.

### Due dates

`tmr`, `fri` (next Friday), `3d`, `2w`, `+5`, `eow` (this Sunday), `today`, or
`2026-07-15`. The form previews the date as you type. Google Tasks stores
dates only, not times of day, so input with a time is rejected.

### Scripting

```sh
tasked add Pay rent --due fri -p 1
tasked done "Pay rent"
tasked move Work milk eggs          # move two tasks to the "Work" list
tasked ls --json | jq -r '.[] | select(.priority == 1) | .title'
tasked rm --yes "old task"
```

Interactive commands fail with an error instead of prompting when stdin is
not a terminal.

### Priorities

| Tag | Meaning |
| --- | --- |
| `p1` | Highest |
| `p2` | High |
| `p3` | Medium (default for new tasks and untagged ones, e.g. added in the Google app) |
| `p4` | Low |
| `p5` | Lowest |

```sh
tasked move p1        # pick tasks, bump them to highest
tasked move Work      # pick tasks, move them to the "Work" list
tasked ls --priority p1
```

## Signing in

`tasked login` uses the standard installed-app OAuth flow (browser sign-in
with a loopback redirect and PKCE) with tasked's own public OAuth client, so
there is nothing to set up. Sign-in tokens are stored per profile under
`~/.config/tasked/tokens/`, readable only by you; nothing is sent anywhere
except Google.

### Using your own OAuth client

If you'd rather use your own Google Cloud project (for your own API quota, or
because your Google Workspace admin blocks third-party apps):

1. In the [Google Cloud console](https://console.cloud.google.com/), create a
   project and enable the **Google Tasks API**.
2. Configure the OAuth consent screen (Google Auth Platform) with user type
   **External**. Under **Audience**, add your account as a **test user** or
   **publish the app**: in *Testing*, Google expires sign-ins after 7 days.
3. Create an OAuth client ID of type **Desktop app**.
4. Set `TASKED_CLIENT_ID` and `TASKED_CLIENT_SECRET` (e.g. from a secrets
   manager like sops-nix), then run `tasked login` again for each profile.

`tasked auth status` shows which client is in use.

> When signing in, make sure the **“See, edit, and delete your tasks”**
> checkbox is ticked on the Google consent screen, or API calls will fail
> with an insufficient-scopes error. Re-run `tasked login` to re-consent.

## Development

```sh
just build       # build ./tasked
just test
just check       # gofmt + vet + tests with -race, same as CI
just clean
```

## License

[MIT](LICENSE)
