package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const bashHook = `
# --- smaran bash shell hook ---
smaran_log() {
  local cmd
  cmd=$(history 1 | sed 's/^[ ]*[0-9]*[ ]*//')
  smaran log --cmd "$cmd" --cwd "$PWD" &disown 2>/dev/null
}
if [[ ! "$PROMPT_COMMAND" =~ smaran_log ]]; then
  PROMPT_COMMAND="smaran_log; $PROMPT_COMMAND"
fi
# --- end smaran hook ---
`

const zshHook = `
# --- smaran zsh shell hook ---
smaran_zsh_log() {
  smaran log --cmd "$(fc -ln -1)" --cwd "$PWD" &disown 2>/dev/null
}
if (( ! ${precmd_functions[(Ie)smaran_zsh_log]} )); then
  precmd_functions+=(smaran_zsh_log)
fi
# --- end smaran hook ---
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

	if strings.Contains(string(content), "smaran_log") || strings.Contains(string(content), "smaran_zsh_log") {
		fmt.Print(Colorize(Yellow, fmt.Sprintf("smaran shell hook is already present in %s\n", rcPath)))
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

	fmt.Print(successText("successfully appended smaran shell hook to %s\n", rcPath))
	fmt.Print(successText("Run 'source %s' or restart your terminal to activate.\n", rcPath))
	return nil
}
