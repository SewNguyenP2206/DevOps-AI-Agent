package command_func

import (
	"ai-agent-go/internal/llm_tool"
	"ai-agent-go/internal/memory_func"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

func resolveSSHInfoFromMemory(serverName string, memory map[string]interface{}) (string, string, string, error) {
	// Convert memory map to a text format for LLM context
	var memText strings.Builder
	for k, v := range memory {
		memText.WriteString(fmt.Sprintf("%s: %v\n", k, v))
	}

	prompt := fmt.Sprintf(`
You are a helpful AI assistant.

Here is the memory data (as plain text key-value pairs):

%v

The user wants to SSH into a server called "%s".

From the memory, identify:
1. The correct SSH private key path (the .pem file path) of the server
2. The correct EC2 IP address of the server
3. The correct username of EC2 instance

Return ONLY a JSON like this:
{
  "keyPath": "/path/to/file.pem",
  "ipAddress": "1.2.3.4",
  "username": "ec2-user"
}

Strict format: JSON only, no explanation.
`, memText.String(), serverName)

	resp, err := llm_tool.AskLLM(prompt)
	if err != nil {
		return "", "", "", err
	}

	var result struct {
		KeyPath   string `json:"keyPath"`
		IPAddress string `json:"ipAddress"`
		Username  string `json:"username"`
	}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return "", "", "", fmt.Errorf("LLM return wrong format: %v", err)
	}

	if result.KeyPath == "" || result.IPAddress == "" || result.Username == "" {
		return "", "", "", fmt.Errorf("Dont have enough info for server %s", serverName)
	}

	return result.KeyPath, result.IPAddress, result.Username, nil
}

func resolveUpdateInfoFromMemory(serverName string, input string, memory map[string]interface{}) (string, error) {
	// Convert memory map to a text format for LLM context
	var memText strings.Builder
	for k, v := range memory {
		memText.WriteString(fmt.Sprintf("%s: %v\n", k, v))
	}

	prompt := fmt.Sprintf(`
You are a helpful AI assistant.

Here is the current memory data (as plain text key-value pairs):

%v

The user wants to update information about the server or project called "%s" with the following input:
"%s"

1. Analyze the user's input and determine which key(s) in the memory should be updated, based on the context and the server/project name.
2. Return a plain text list of updates, one per line, in the format: key: value
3. If you cannot determine what to update, return an empty string.

Strict format: key: value per line, no explanation.
Example:
EC2 IP Address: new.ip.address
EC2 Username: new-username
`, memText.String(), serverName, input)

	resp, err := llm_tool.AskLLM(prompt)
	if err != nil {
		return "", err
	}

	resp = strings.TrimSpace(resp)
	if resp == "" {
		return "", fmt.Errorf("No updatable info detected or wrong format")
	}

	// Parse the plain text response and update memory
	lines := strings.Split(resp, "\n")
	updated := make(map[string]string)
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			memory[key] = value
			updated[key] = value
		}
	}

	if len(updated) == 0 {
		return "", fmt.Errorf("No updatable info detected or wrong format")
	}

	// Save updated memory
	var memoryLines []string
	for k, v := range memory {
		memoryLines = append(memoryLines, fmt.Sprintf("%s: %v", k, v))
	}
	if err := memory_func.SaveMemory("memory.txt", memoryLines); err != nil {
		return "", fmt.Errorf("Failed to save updated memory: %v", err)
	}

	return fmt.Sprintf("Updated fields: %v", updated), nil
}

// openTerminalAndRunCommand opens a new Terminal window and runs the given command (macOS only)
func OpenTerminalAndRunCommand(cmd string) error {
	script := fmt.Sprintf(`tell application "Terminal"
    activate
    do script "%s"
end tell`, cmd)

	return exec.Command("osascript", "-e", script).Run()
}

// HandleCommand: try to extract server name and build SSH command from memory facts
func HandleCommand(input string, memory []string) (string, error) {
	serverName, err := extractServerNameFromInput(input, memory)
	if err != nil {
		return "", fmt.Errorf("Error extracting server name: %v", err)
	}

	// Find facts for keyPath, ipAddress, username
	var keyPath, ipAddress, username string
	for _, line := range memory {
		l := strings.ToLower(line)
		if strings.Contains(l, strings.ToLower(serverName)) {
			if strings.Contains(l, "key") && strings.Contains(l, "pem") {
				keyPath = extractValueFromFact(line)
			}
			if strings.Contains(l, "ip") {
				ipAddress = extractValueFromFact(line)
			}
			if strings.Contains(l, "user") || strings.Contains(l, "username") {
				username = extractValueFromFact(line)
			}
		}
	}
	if keyPath == "" || ipAddress == "" || username == "" {
		return "", fmt.Errorf("Missing info for SSH: keyPath=%s, ip=%s, user=%s", keyPath, ipAddress, username)
	}

	cmd := fmt.Sprintf(`ssh -i %s %s@%s`, keyPath, username, ipAddress)
	return cmd, nil
}
