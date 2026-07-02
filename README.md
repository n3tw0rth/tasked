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
- **Scriptable**: plain flags and output where it matters

## Install

Requires Go 1.26+ and a [Google OAuth client](#oauth-credentials).

```sh
git clone https://github.com/n3tw0rth/tasked
cd tasked
cp .env.example .env   # add your OAuth client id/secret
just install           # or: just build
```

## Quick start

```sh
tasked login                # sign in with Google (opens browser)
tasked lists switch         # pick the task list to work on
tasked add                  # create a task: title, due date, priority
tasked ls                   # tasks sorted by priority, then due date
tasked done                 # mark tasks completed (multi-select)
```

## Commands

| Command | Description |
| --- | --- |
| `add` | Create a task via an inline form (title, due, priority) |
| `ls` | List tasks sorted by priority (`--all`, `--priority p1..p5`, `--list <id>`) |
| `done` | Mark tasks completed (multi-select picker) |
| `move [p1-p5]` | Change task priorities: pick tasks, then a priority (or pass it as the arg) |
| `rm` | Delete tasks (multi-select picker) |
| `search <query>` | Search tasks by title/notes in the active list |
| `lists` | Manage task lists: `ls`, `create`, `switch`, `rm` |
| `profile` | Manage accounts: `list`, `add`, `use`, `remove` |
| `login` / `logout` | Sign in to Google / remove the saved token |

Every command accepts `--profile <name>` to target a specific account without
switching the active profile.

### Priorities

| Tag | Meaning |
| --- | --- |
| `p1` | Highest |
| `p2` | High |
| `p3` | Medium |
| `p4` | Low |
| `p5` | Lowest (default for untagged tasks) |

```sh
tasked move p1        # pick tasks, bump them to highest
tasked ls --priority p1
```

## OAuth credentials

`tasked` uses the standard installed-app OAuth flow (loopback redirect). You
need a **Desktop app** OAuth client from the
[Google Cloud console](https://console.cloud.google.com/apis/credentials),
with the **Google Tasks API** enabled for the project.

Credentials are resolved in this order:

1. `TASKED_CLIENT_ID` / `TASKED_CLIENT_SECRET` environment variables
2. Values embedded at build time via `-ldflags` (`just build` reads them
   from `.env`)

Tokens are stored per profile under `~/.config/tasked/tokens/` with `0600`
permissions. `.env` is gitignored: never commit real credentials.

> When signing in, make sure the **“See, edit, and delete your tasks”**
> checkbox is ticked on the Google consent screen, or API calls will fail
> with an insufficient-scopes error. Re-run `tasked login` to re-consent.

## Development

```sh
just build       # release build with credentials from .env
just build-dev   # dev build, credentials from env vars at runtime
just test
just clean
```
