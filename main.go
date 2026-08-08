package main

import (
	"fmt"
	"os"
	"strconv"
)

var Version = "dev"

func main() {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "version", "-v", "--version":
			fmt.Printf("didfix %s\n", Version)
			return
		}
	}

	if len(os.Args) >= 2 && os.Args[1] == "init" {
		shell := ""
		install := false
		for _, a := range os.Args[2:] {
			if a == "bash" || a == "zsh" {
				shell = a
			}
			if a == "--install" || a == "-i" {
				install = true
			}
		}
		if err := printOrInstallInit(shell, install); err != nil {
			fmt.Fprintln(os.Stderr, "init error:", err)
			os.Exit(1)
		}
		return
	}

	db, err := openDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, "didfix: db error:", err)
		os.Exit(1)
	}
	defer db.Close()

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "log":
		cmd, cwd, sessionID := parseLogFlags(os.Args[2:])
		if cmd != "" {
			_ = logCommand(db, cmd, cwd, sessionID)
		}

	case "tag":
		note := ""
		n := 1
		for i := 2; i < len(os.Args); i++ {
			if (os.Args[i] == "--note" || os.Args[i] == "-n" || os.Args[i] == "-m") && i+1 < len(os.Args) {
				note = os.Args[i+1]
				i++
			} else if (os.Args[i] == "--last" || os.Args[i] == "-l") && i+1 < len(os.Args) {
				n, _ = strconv.Atoi(os.Args[i+1])
				i++
			} else if note == "" && !startsWithDash(os.Args[i]) {
				note = os.Args[i]
			}
		}

		if note == "" {
			fmt.Fprintln(os.Stderr, "error: --note is required (e.g. didfix tag --note \"fixed memory leak\" --last 3)")
			os.Exit(1)
		}

		tagID, count, err := tagLastN(db, note, n)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error tagging commands:", err)
			os.Exit(1)
		}
		fmt.Printf("tagged %d command(s) under tag #%d: %q\n", count, tagID, note)

	case "search":
		if len(os.Args) < 3 {
			fmt.Println("usage: didfix search \"query\"")
			return
		}
		query := os.Args[2]
		if err := search(db, query); err != nil {
			fmt.Fprintln(os.Stderr, "search error:", err)
			os.Exit(1)
		}

	case "show":
		if len(os.Args) < 3 {
			fmt.Println("usage: didfix show <tag_id>")
			return
		}
		tagID, err := strconv.ParseInt(os.Args[2], 10, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid tag ID:", os.Args[2])
			os.Exit(1)
		}
		if err := showTag(db, tagID); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

	case "list", "recent":
		limit := 10
		if len(os.Args) >= 3 {
			if l, err := strconv.Atoi(os.Args[2]); err == nil && l > 0 {
				limit = l
			}
		}
		if err := listRecent(db, limit); err != nil {
			fmt.Fprintln(os.Stderr, "list error:", err)
			os.Exit(1)
		}

	case "untag":
		if len(os.Args) < 3 {
			fmt.Println("usage: didfix untag <tag_id>")
			return
		}
		tagID, err := strconv.ParseInt(os.Args[2], 10, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid tag ID:", os.Args[2])
			os.Exit(1)
		}
		if err := untag(db, tagID); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func parseLogFlags(args []string) (cmd, cwd, sessionID string) {
	for i := 0; i < len(args); i++ {
		if args[i] == "--cmd" && i+1 < len(args) {
			cmd = args[i+1]
			i++
		} else if args[i] == "--cwd" && i+1 < len(args) {
			cwd = args[i+1]
			i++
		} else if args[i] == "--session" && i+1 < len(args) {
			sessionID = args[i+1]
			i++
		}
	}
	return
}

func startsWithDash(s string) bool {
	return len(s) > 0 && s[0] == '-'
}

func printUsage() {
	usage := `didfix - Terminal command fix logger & history search

Usage:
  didfix log --cmd "<cmd>" --cwd "<path>"    Record a command (used by shell hook)
  didfix tag --note "<text>" [--last N]      Tag last N commands with a fix note
  didfix search "<query>"                    Search tagged fixes using full-text search
  didfix show <tag_id>                       Show detailed information for a tag
  didfix list [N]                            List N recent tagged fixes (default 10)
  didfix untag <tag_id>                      Remove a tag and unbind commands
  didfix init [bash|zsh] [--install]         Output or install shell hook configuration
  didfix version                             Print version information
`
	fmt.Print(usage)
}
