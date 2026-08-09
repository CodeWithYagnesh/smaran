package main

import (
	"fmt"
	"os"
	"strconv"
)

var Version = "dev"
var noColor = false

func main() {
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		if args[i] == "--no-color" {
			noColor = true
			args = append(args[:i], args[i+1:]...)
			i--
		}
	}
	os.Args = append([]string{os.Args[0]}, args...)

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "version", "-v", "--version":
			fmt.Printf("smaran %s\n", Colorize(Cyan, Version))
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
			fmt.Fprintln(os.Stderr, errorText("init error: %v", err))
			os.Exit(1)
		}
		return
	}

	db, err := openDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, errorText("smaran: db error: %v", err))
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
		cmdToAdd := ""
		cwd := ""
		tagUpdateID := int64(0)
		var n int
		for i := 2; i < len(os.Args); i++ {
			if (os.Args[i] == "--note" || os.Args[i] == "-n" || os.Args[i] == "-m") && i+1 < len(os.Args) {
				note = os.Args[i+1]
				i++
			} else if (os.Args[i] == "--last" || os.Args[i] == "-l") && i+1 < len(os.Args) {
				n, _ = strconv.Atoi(os.Args[i+1])
				i++
			} else if os.Args[i] == "--cmd" && i+1 < len(os.Args) {
				cmdToAdd = os.Args[i+1]
				i++
			} else if os.Args[i] == "--cwd" && i+1 < len(os.Args) {
				cwd = os.Args[i+1]
				i++
			} else if (os.Args[i] == "--update" || os.Args[i] == "-u") && i+1 < len(os.Args) {
				if id, err := strconv.ParseInt(os.Args[i+1], 10, 64); err == nil {
					tagUpdateID = id
				} else {
					fmt.Fprintln(os.Stderr, errorText("invalid tag ID for --update: %v", os.Args[i+1]))
					os.Exit(1)
				}
				i++
			} else if note == "" && !startsWithDash(os.Args[i]) {
				note = os.Args[i]
			}
		}

		if note == "" {
			fmt.Fprintln(os.Stderr, errorText("error: --note is required (e.g. smaran tag --note \"fixed memory leak\" --last 3)"))
			os.Exit(1)
		}

		if tagUpdateID != 0 {
			if cmdToAdd != "" {
				fmt.Fprintln(os.Stderr, errorText("cannot use --cmd and --update together"))
				os.Exit(1)
			}
			if err := updateTagNote(db, tagUpdateID, note); err != nil {
				fmt.Fprintln(os.Stderr, errorText("error updating tag note: %v", err))
				os.Exit(1)
			}
			fmt.Println(successText("updated tag #%d note to %s", tagUpdateID, note))
			return
		}

		if cmdToAdd != "" {
			tagID, count, err := addTaggedCommand(db, cmdToAdd, cwd, note)
			if err != nil {
				fmt.Fprintln(os.Stderr, errorText("error tagging direct command: %v", err))
				os.Exit(1)
			}
			fmt.Println(successText("tagged %d command(s) under tag #%d: %s", count, tagID, note))
			return
		}

		tagID, count, err := tagLastN(db, note, n)
		if err != nil {
			fmt.Fprintln(os.Stderr, errorText("error tagging commands: %v", err))
			os.Exit(1)
		}
		fmt.Println(successText("tagged %d command(s) under tag #%d: %s", count, tagID, note))

	case "search":
		if len(os.Args) < 3 {
			fmt.Println(Colorize(Yellow, "usage: smaran search \"query\""))
			return
		}
		query := os.Args[2]
		if err := search(db, query); err != nil {
			fmt.Fprintln(os.Stderr, errorText("search error: %v", err))
			os.Exit(1)
		}

	case "show":
		if len(os.Args) < 3 {
			fmt.Println(Colorize(Yellow, "usage: smaran show <tag_id>"))
			return
		}
		tagID, err := strconv.ParseInt(os.Args[2], 10, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid tag ID:", os.Args[2])
			os.Exit(1)
		}
		if err := showTag(db, tagID); err != nil {
			fmt.Fprintln(os.Stderr, errorText("error: %v", err))
			os.Exit(1)
		}

	case "list":
		limit := 10
		if len(os.Args) >= 3 {
			if l, err := strconv.Atoi(os.Args[2]); err == nil && l > 0 {
				limit = l
			}
		}
		if err := listRecent(db, limit); err != nil {
			fmt.Fprintln(os.Stderr, errorText("list error: %v", err))
			os.Exit(1)
		}

	case "tags":
		limit := 10
		if len(os.Args) >= 3 {
			if l, err := strconv.Atoi(os.Args[2]); err == nil && l > 0 {
				limit = l
			}
		}
		if err := listTags(db, limit); err != nil {
			fmt.Fprintln(os.Stderr, errorText("tags error: %v", err))
			os.Exit(1)
		}

	case "delete":
		if len(os.Args) < 3 {
			fmt.Println(Colorize(Yellow, "usage: smaran delete <command_id>"))
			return
		}
		commandID, err := strconv.ParseInt(os.Args[2], 10, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid command ID:", os.Args[2])
			os.Exit(1)
		}
		if err := deleteCommand(db, commandID); err != nil {
			fmt.Fprintln(os.Stderr, errorText("error deleting command: %v", err))
			os.Exit(1)
		}

	case "commands", "cmds", "history":
		limit := 10
		pattern := ""
		tagID := int64(0)
		for i := 2; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "-s", "--search":
				if i+1 < len(os.Args) {
					pattern = os.Args[i+1]
					i++
				}
			case "--tag", "--tag-id":
				if i+1 < len(os.Args) {
					if id, err := strconv.ParseInt(os.Args[i+1], 10, 64); err == nil {
						tagID = id
					} else {
						fmt.Fprintln(os.Stderr, errorText("invalid tag ID: %v", os.Args[i+1]))
						os.Exit(1)
					}
					i++
				}
			case "-l", "--limit":
				if i+1 < len(os.Args) {
					if l, err := strconv.Atoi(os.Args[i+1]); err == nil && l > 0 {
						limit = l
					}
					i++
				}
			default:
				if pattern == "" {
					if l, err := strconv.Atoi(os.Args[i]); err == nil && l > 0 {
						limit = l
					} else {
						pattern = os.Args[i]
					}
				}
			}
		}
		if pattern != "" && tagID != 0 {
			fmt.Fprintln(os.Stderr, errorText("cannot use --search and --tag together"))
			os.Exit(1)
		}
		if tagID != 0 {
			if err := listCommandsByTag(db, tagID, limit); err != nil {
				fmt.Fprintln(os.Stderr, errorText("commands error: %v", err))
				os.Exit(1)
			}
			return
		}
		if pattern != "" {
			if err := searchHistory(db, pattern, limit); err != nil {
				fmt.Fprintln(os.Stderr, errorText("commands error: %v", err))
				os.Exit(1)
			}
			return
		}
		if err := listCommands(db, limit); err != nil {
			fmt.Fprintln(os.Stderr, errorText("commands error: %v", err))
			os.Exit(1)
		}

	case "untag":
		if len(os.Args) < 3 {
			fmt.Println(Colorize(Yellow, "usage: smaran untag <tag_id>"))
			return
		}
		tagID, err := strconv.ParseInt(os.Args[2], 10, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid tag ID:", os.Args[2])
			os.Exit(1)
		}
		if err := untag(db, tagID); err != nil {
			fmt.Fprintln(os.Stderr, errorText("error: %v", err))
			os.Exit(1)
		}

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Fprintln(os.Stderr, errorText("unknown command %q", subcommand))
		fmt.Fprintln(os.Stderr)
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

func errorText(format string, a ...interface{}) string {
	return Colorize(Red, fmt.Sprintf(format, a...))
}

func successText(format string, a ...interface{}) string {
	return Colorize(Green, fmt.Sprintf(format, a...))
}

func printUsage() {
	usage := fmt.Sprintf("%s - %s\n\nUsage:\n",
		Colorize(Green, "smaran"),
		Colorize(Gray, "Terminal command fix logger & history search"))
	usage += fmt.Sprintf("  %s %s %s\n",
		Colorize(Cyan, "smaran"),
		Colorize(Cyan, "log"),
		Colorize(Yellow, "--cmd \"<cmd>\" --cwd \"<path>\""))
	usage += fmt.Sprintf("    %s\n",
		Colorize(Gray, "Record a command (used by shell hook)"))
	usage += fmt.Sprintf("  %s %s %s      %s\n",
		Colorize(Cyan, "smaran"),
		Colorize(Cyan, "tag"),
		Colorize(Yellow, "--note \"<text>\" [--last N]"),
		Colorize(Gray, "Tag last N commands with a fix note"))
	usage += fmt.Sprintf("  %s %s %s  %s\n",
		Colorize(Cyan, "smaran"),
		Colorize(Cyan, "tag"),
		Colorize(Yellow, "--cmd \"<command>\" --note \"<text>\" [--cwd \"<path>\"]"),
		Colorize(Gray, "Add and tag a direct command without running it"))
	usage += fmt.Sprintf("  %s %s %s  %s\n",
		Colorize(Cyan, "smaran"),
		Colorize(Cyan, "tag"),
		Colorize(Yellow, "--update <tag_id> --note \"<text>\""),
		Colorize(Gray, "Update an existing tag note"))
	usage += fmt.Sprintf("  %s %s %s                    %s\n",
		Colorize(Cyan, "smaran"),
		Colorize(Cyan, "search"),
		Colorize(Yellow, "\"<query>\""),
		Colorize(Gray, "Search tagged fixes using full-text search"))
	usage += fmt.Sprintf("  %s %s %s                       %s\n",
		Colorize(Cyan, "smaran"),
		Colorize(Cyan, "show"),
		Colorize(Yellow, "<tag_id>"),
		Colorize(Gray, "Show detailed information for a tag"))
	usage += fmt.Sprintf("  %s %s [N]                            %s\n",
		Colorize(Cyan, "smaran"),
		Colorize(Cyan, "list"),
		Colorize(Gray, "List N recent tagged fixes (default 10)"))
	usage += fmt.Sprintf("  %s %s [N]                            %s\n",
		Colorize(Cyan, "smaran"),
		Colorize(Cyan, "tags"),
		Colorize(Gray, "List N recent tags (default 10)"))
	usage += fmt.Sprintf("  %s %s [N]                       %s\n",
		Colorize(Cyan, "smaran"),
		Colorize(Cyan, "commands"),
		Colorize(Gray, "List N recent raw commands (default 10)"))
	usage += fmt.Sprintf("  %s %s %s  %s\n",
		Colorize(Cyan, "smaran"),
		Colorize(Cyan, "commands"),
		Colorize(Yellow, "-s \"<regex>\" [--limit N]"),
		Colorize(Gray, "Search raw command history using regex"))
	usage += fmt.Sprintf("  %s %s %s %s\n",
		Colorize(Cyan, "smaran"),
		Colorize(Cyan, "commands"),
		Colorize(Yellow, "--tag <tag_id> [--limit N]"),
		Colorize(Gray, "List raw commands for a specific tag"))
	usage += fmt.Sprintf("  %s %s %s                 %s\n",
		Colorize(Cyan, "smaran"),
		Colorize(Cyan, "delete"),
		Colorize(Yellow, "<command_id>"),
		Colorize(Gray, "Delete a raw history command by ID"))
	usage += fmt.Sprintf("  %s %s %s                      %s\n",
		Colorize(Cyan, "smaran"),
		Colorize(Cyan, "untag"),
		Colorize(Yellow, "<tag_id>"),
		Colorize(Gray, "Remove a tag and unbind commands"))
	usage += fmt.Sprintf("  %s %s %s                %s\n",
		Colorize(Cyan, "smaran"),
		Colorize(Cyan, "--no-color"),
		Colorize(Yellow, "<command>"),
		Colorize(Gray, "Disable ANSI color output"))
	usage += fmt.Sprintf("  %s %s [bash|zsh] [--install]         %s\n",
		Colorize(Cyan, "smaran"),
		Colorize(Cyan, "init"),
		Colorize(Gray, "Output or install shell hook configuration"))
	usage += fmt.Sprintf("  %s %s                             %s\n",
		Colorize(Cyan, "smaran"),
		Colorize(Cyan, "version"),
		Colorize(Gray, "Print version information"))
	fmt.Print(usage)
}
