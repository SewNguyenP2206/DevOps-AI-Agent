// Now memory is []string, so search for server name in text lines
package command_func

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func extractServerNameFromInput(input string, memory []string) (string, error) {
	// Simple heuristic: look for a word in input that matches a word in memory lines
	words := strings.Fields(input)
	for _, line := range memory {
		for _, w := range words {
			if strings.Contains(strings.ToLower(line), strings.ToLower(w)) {
				return w, nil
			}
		}
	}
	return "", fmt.Errorf("Cannot identify server from input.")
}

// Helper: extract value (after "is") from a fact sentence
func extractValueFromFact(line string) string {
	parts := strings.Split(line, " is ")
	if len(parts) > 1 {
		return strings.TrimSpace(parts[len(parts)-1])
	}
	return ""
}

// RunSSHCommand runs an SSH command directly in the current terminal
func RunSSHCommand(sshCmd string) error {
	parts := strings.Fields(sshCmd)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func UpdateMemoryFact(memory []string, oldFact, newFact string) []string {
	for i, line := range memory {
		if strings.Contains(line, oldFact) {
			memory[i] = newFact
			return memory
		}
	}
	return append(memory, newFact)
}

// Helper function to execute a shell command
func executeShellCommand(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}
	cmd := parts[0]
	args := parts[1:]
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
