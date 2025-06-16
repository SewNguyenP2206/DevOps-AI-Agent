package git_func

import (
	"ai-agent-go/internal/llm_tool"
	"ai-agent-go/internal/memory_func"
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func HandleGitCloneRequest(input string, memory *[]string) {
	reader := bufio.NewReader(os.Stdin)

	prompt := fmt.Sprintf(`
You are an AI assistant. The user wants to clone a Git repository.
Request: "%s"
Here is the current memory:
%s

Your task:
1. Identify the Git repository URL from the memory.
2. Identify the directory where the user wants to clone the repo.

Respond ONLY with JSON in the format:
{
  "url": "https://...",
  "directory": "/path/to/clone"
}
No explanation, no markdown.
`, input, strings.Join(*memory, "\n"))

	resp, err := llm_tool.AskLLM(prompt)
	if err != nil {
		fmt.Println("❌ Error from AI:", err)
		return
	}

	var result struct {
		URL       string `json:"url"`
		Directory string `json:"directory"`
	}

	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		fmt.Println("❌ Failed to parse AI response:", err)
		fmt.Println("🔎 Raw response:", resp)
		return
	}

	// Ask for repo URL if missing
	if result.URL == "" {
		fmt.Println("🤔 I don’t have the Git repository URL in memory. Please enter it:")
		fmt.Print(">>> ")
		url, _ := reader.ReadString('\n')
		result.URL = strings.TrimSpace(url)
		if result.URL == "" {
			fmt.Println("❌ No URL provided. Cannot proceed.")
			return
		}
	}

	// Ask for directory if missing
	if result.Directory == "" {
		fmt.Println("🤔 I don’t have the target directory in memory. Please enter it:")
		fmt.Print(">>> ")
		dir, _ := reader.ReadString('\n')
		result.Directory = strings.TrimSpace(dir)
		if result.Directory == "" {
			fmt.Println("❌ No directory provided. Cannot proceed.")
			return
		}
	}

	// Create directory if not exists
	if _, err := os.Stat(result.Directory); os.IsNotExist(err) {
		fmt.Printf("❗ Directory '%s' does not exist. Create it? (yes/no)\n>>> ", result.Directory)
		confirm, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(confirm)) != "yes" {
			fmt.Println("❌ Cannot proceed without directory.")
			return
		}
		if err := os.MkdirAll(result.Directory, 0755); err != nil {
			fmt.Println("❌ Failed to create directory:", err)
			return
		}
	}

	// Run git clone
	fmt.Println("🔄 Cloning repository...")
	cmd := exec.Command("git", "clone", result.URL, result.Directory)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("❌ Git clone failed:", err)
		return
	}

	fmt.Println("✅ Successfully cloned repository to:", result.Directory)

	// Update memory
	newFact := fmt.Sprintf("Cloned repository %s to %s", result.URL, result.Directory)
	*memory = append(*memory, newFact)
	if err := memory_func.SaveMemory("devOpsmemory.txt", *memory); err != nil {
		fmt.Println("⚠️ Failed to save memory:", err)
	} else {
		fmt.Println("🧠 Memory updated and saved.")
	}
}
