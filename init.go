package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const bashHook = `
# --- didfix bash shell hook ---
didfix_log() {
  local cmd
  cmd=$(history 1 | sed 's/^[ ]*[0-9]*[ ]*//')
  didfix log --cmd "$cmd" --cwd "$PWD" &disown 2>/dev/null
}
if [[ ! "$PROMPT_COMMAND" =~ didfix_log ]]; then
  PROMPT_COMMAND="didfix_log; $PROMPT_COMMAND"
fi
# --- end didfix hook ---
`

const zshHook = `
# --- didfix zsh shell hook ---
didfix_zsh_log() {
  didfix log --cmd "$(fc -ln -1)" --cwd "$PWD" &disown 2>/dev/null
}
if (( ! ${precmd_functions[(Ie)didfix_zsh_log]} )); then
  precmd_functions+=(didfix_zsh_log)
fi
# --- end didfix hook ---
`

func printOrInstallInit(shell string, install bool) error {
	if shell == "" {
		// Detect shell from SHELL environment variable
		shellPath := os.Getenv("SHELL")
		if strings.Contains(shellPath, "zsh") {
			shell = "zsh"
		} else {
			shell = "bash"
		}
	}

	hookText := bashHook
	rcFileName := ".bashrc"
	if shell == "zsh" {
		hookText = zshHook
		rcFileName = ".zshrc"
	}

	if !install {
		fmt.Print(strings.TrimSpace(hookText))
		fmt.Println()
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	rcPath := filepath.Join(home, rcFileName)
	content, _ := os.ReadFile(rcPath)

	if strings.Contains(string(content), "didfix_log") || strings.Contains(string(content), "didfix_zsh_log") {
		fmt.Printf("didfix shell hook is already present in %s\n", rcPath)
		return nil
	}

	f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", rcPath, err)
	}
	defer f.Close()

	if _, err := f.WriteString("\n" + strings.TrimSpace(hookText) + "\n"); err != nil {
		return err
	}

	fmt.Printf("successfully appended didfix shell hook to %s\n", rcPath)
	fmt.Printf("Run 'source %s' or restart your terminal to activate.\n", rcPath)
	return nil
}
