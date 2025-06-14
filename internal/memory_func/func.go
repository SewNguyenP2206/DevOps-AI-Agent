package memory_func

import (
	"ai-agent-go/internal/llm_tool"
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadMemory loads memory from a file, one fact per line.
func LoadMemory(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var memory []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			memory = append(memory, line)
		}
	}
	return memory, scanner.Err()
}

// SaveMemory saves memory to a file, one fact per line.
func SaveMemory(filename string, memory []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, line := range memory {
		_, err := file.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

// PromptToMemory now returns a natural language fact string
func PromptToMemory(question, answer string) (string, error) {
	prompt := fmt.Sprintf(`
You are an AI assistant helping your user save memory.

The user asked this question:
"%s"

Then the user provided this answer:
"%s"

Based on both, return a single natural language fact sentence to store in memory.
Example:
"EC2 username of Convos is ec2-user"
"IP of Revoland ec2 instance is 123.123.123.123"
"Directory of Linux folder is /Users/sewn/Linux"

Return only the fact sentence, no explanation.
`, question, answer)

	resp, err := llm_tool.AskLLM(prompt)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resp), nil
}
