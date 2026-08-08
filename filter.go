package main

import (
	"strings"
)

var ignoredPrefixes = []string{
	"ls",
	"cd",
	"pwd",
	"clear",
	"exit",
	"history",
	"didfix",
	"wasrun",
	"top",
	"htop",
	"which",
	"whoami",
	"man",
}

func shouldIgnoreCommand(cmd string) bool {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return true
	}

	// Extract the main binary/first token of the command line
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return true
	}

	baseCmd := fields[0]
	// If command starts with environment variables like FOO=bar cmd, find actual command
	for i, f := range fields {
		if strings.Contains(f, "=") && !strings.HasPrefix(f, "-") && !strings.HasPrefix(f, "./") && !strings.HasPrefix(f, "/") {
			continue
		}
		baseCmd = fields[i]
		break
	}

	// Strip leading path if present (e.g. /bin/ls -> ls)
	if idx := strings.LastIndex(baseCmd, "/"); idx != -1 {
		baseCmd = baseCmd[idx+1:]
	}

	for _, ignored := range ignoredPrefixes {
		if baseCmd == ignored {
			return true
		}
	}

	return false
}
