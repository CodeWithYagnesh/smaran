# smaran

`smaran` is a terminal command history logger, fix/tag manager, and searchable command helper for shell workflows.
It records every command run by your shell hook, lets you attach fix notes, and makes it easy to search both tagged fixes and raw command history.

## Quick Start

```sh
git clone https://github.com/CodeWithYagnesh/smaran.git
cd smaran
make install
smaran init bash --install
```

Then restart your shell or source your shell profile to begin capturing commands.

## What is in this README?

- Installation methods: source build and GitHub install script
- Shell hook setup for Bash/Zsh
- Basic usage and examples
- Full command reference
- Notes and troubleshooting tips

## Install

### Install from source

Build and install from the repository:

```sh
make install
```

This installs `smaran` to `/usr/local/bin/smaran` by default.

Remove the installed binary with:

```sh
make uninstall
```

### Install from GitHub

Install `smaran` directly from GitHub using the repository install script:

```sh
curl -fsSL https://raw.githubusercontent.com/CodeWithYagnesh/smaran/main/install.sh | sh
```

To install to a custom directory, set `INSTALL_DIR` before piping the script:

```sh
INSTALL_DIR="$HOME/.local/bin" curl -fsSL https://raw.githubusercontent.com/CodeWithYagnesh/smaran/main/install.sh | sh
```

If installation requires `sudo`, the script will prompt for it automatically when needed.

### Verify installation

```sh
smaran version
```

If the command prints a version, `smaran` is installed correctly.

## Configure the shell hook

The shell hook captures each command and the working directory so `smaran` can link raw history entries with tags.

Generate configuration for Bash or Zsh:

```sh
smaran init bash
smaran init zsh
```

Install the hook automatically:

```sh
smaran init bash --install
```

Restart your shell or source your shell configuration file after installation.

## Basic Usage

### Record a command from a shell hook

```sh
smaran log --cmd "<cmd>" --cwd "<path>"
```

This command usually runs automatically from the shell hook.

### Tag recent commands with a fix note

```sh
smaran tag --note "fixed path issue" --last 2
```

This creates a new tag and attaches the note to the last `N` recorded commands.

### Add and tag a command directly

```sh
smaran tag --cmd "git checkout main" --note "restore branch" --cwd "/home/user"
```

This adds a raw command entry and tags it immediately.

### Update an existing tag note

```sh
smaran tag --update 3 --note "improved fix description"
```

Use the tag ID from `smaran tags` or `smaran list`.

## Search and List Commands

### Search tagged fixes

```sh
smaran search "query"
```

Searches fix notes and tagged command history using full-text search.

### Show commands attached to a tag

```sh
smaran show <tag_id>
```

### List recent tagged fixes

```sh
smaran list [N]
```

Defaults to `10` if `N` is omitted.

### List recent tags

```sh
smaran tags [N]
```

### List raw commands

```sh
smaran commands [N]
```

### Search raw command history with regex

```sh
smaran commands -s "<regex>" [--limit N]
```

### Filter raw commands by tag

```sh
smaran commands --tag <tag_id> [--limit N]
```

## Maintenance Commands

### Delete a raw history entry

```sh
smaran delete <command_id>
```

### Remove a tag and unbind its commands

```sh
smaran untag <tag_id>
```

### Disable ANSI colors for a single command

```sh
smaran --no-color <command>
```

Use this when you want plain text output for scripts or terminals that do not support ANSI escapes.

## Full Command Reference

```sh
smaran log --cmd "<cmd>" --cwd "<path>"
smaran tag --note "<text>" [--last N]
smaran tag --cmd "<command>" --note "<text>" [--cwd "<path>"]
smaran tag --update <tag_id> --note "<text>"
smaran search "<query>"
smaran show <tag_id>
smaran list [N]
smaran tags [N]
smaran commands [N]
smaran commands -s "<regex>" [--limit N]
smaran commands --tag <tag_id> [--limit N]
smaran delete <command_id>
smaran untag <tag_id>
smaran --no-color <command>
smaran init [bash|zsh] [--install]
smaran version
```

## Examples

```sh
# Tag the last 3 raw commands with a fix note
smaran tag --note "fixed version mismatch" --last 3

# Add a tagged command without running it
smaran tag --cmd "git pull origin main" --note "restore latest branch" --cwd "$HOME/projects"

# Search tagged notes for a keyword
smaran search "database reconnect"

# Search raw history with a regex
smaran commands -s "npm.*install" --limit 20

# List commands attached to a specific tag
smaran commands --tag 5 --limit 10

# Delete a bad history entry by ID
smaran delete 42

# Remove a tag and detach its commands
smaran untag 6

# Run a command without ANSI colors
smaran --no-color search "error"
```

## Why `smaran` helps

`smaran` is useful when you want to save the commands that fixed an issue, recover past shell work, and attach searchable notes to your terminal history.

The tag-based workflow makes it easy to find the exact commands associated with a fix, and raw command search helps rediscover forgotten steps.

## Notes

- Tag IDs appear in `smaran tags` and `smaran list` output.
- Command IDs appear in `smaran commands` output.
- Install the shell hook with `smaran init --install` to capture commands automatically.
- Use `smaran version` to verify the installed binary version.
