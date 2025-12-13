package app_interaction

import (
	"ai-agent-go/internal/llm_tool"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// HandleAppOpenRequest processes a request 
func HandleAppOpenRequest(input string, memory *[]string) {
	prompt := fmt.Sprintf(`
You are an AI assistant. The user wants to open an application.
Request: "%s"
Here is the current memory:
%s
Your task:
1. Extract the application name from the request.
Respond ONLY with JSON in the format:
{
  "app_name": "Application Name"
}
No explanation, no markdown.
`, input, strings.Join(*memory, "\n"))

	resp, err := llm_tool.AskLLM(prompt)
	if err != nil {
		fmt.Println("❌ Error from AI:", err)
		return
	}

	var result struct {
		AppName string `json:"app_name"`
	}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		fmt.Println("❌ Failed to parse AI response:", err)
		fmt.Println("🔎 Raw response:", resp)
		return
	}

	// Check if the application exists on the user's system
	appList, err := exec.Command("ls", "/Applications").Output()
	if err != nil {
		fmt.Println("❌ Failed to list applications:", err)
		return
	}

	applications := strings.Split(string(appList), "\n")
	appNameInput := strings.ToLower(strings.TrimSpace(result.AppName))
	var matchedApp string

	for _, app := range applications {
		if strings.Contains(strings.ToLower(app), appNameInput) {
			matchedApp = strings.TrimSuffix(app, ".app")
			break
		}
	}

	if matchedApp == "" {
		fmt.Println("❌ Application not found on the system. Please check the name and try again.")
		return
	}

	fmt.Printf("🔍 Opening application: %s\n", matchedApp)

	// Execute the command to open the application
	cmd := exec.Command("open", "-a", matchedApp)
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Failed to open application: %s\n", err)
		return
	}

	fmt.Printf("✅ Successfully opened application: %s\n", matchedApp)
	// Update memory with the new application opened
	*memory = append(*memory, fmt.Sprintf("Opened application: %s", matchedApp))
}

// HandleAppCloseRequest processes a request to close an application
func HandleAppCloseRequest(input string, memory *[]string) {
	prompt := fmt.Sprintf(`
You are an AI assistant. The user wants to close an application.
Request: "%s"
Here is the current memory:
%s
Your task:	
1. Extract the application name from the request.
Respond ONLY with JSON in the format:
{
  "app_name": "Application Name"
}
No explanation, no markdown.
`, input, strings.Join(*memory, "\n"))
	resp, err := llm_tool.AskLLM(prompt)
	if err != nil {
		fmt.Println("❌ Error from AI:", err)
		return
	}
	var result struct {
		AppName string `json:"app_name"`
	}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		fmt.Println("❌ Failed to parse AI response:", err)
		fmt.Println("🔎 Raw response:", resp)
		return
	}
	// Check if the application exists on the user's system
	appList, err := exec.Command("ls", "/Applications").Output()
	if err != nil {
		fmt.Println("❌ Failed to list applications:", err)
		return
	}
	applications := strings.Split(string(appList), "\n")
	appNameInput := strings.ToLower(strings.TrimSpace(result.AppName))
	var matchedApp string
	for _, app := range applications {
		if strings.Contains(strings.ToLower(app), appNameInput) {
			matchedApp = strings.TrimSuffix(app, ".app")
			break
		}
	}
	if matchedApp == "" {
		fmt.Println("❌ Application not found on the system. Please check the name and try again.")
		return
	}
	fmt.Printf("🔍 Closing application: %s\n", matchedApp)
	// Execute the command to close the application
	cmd := exec.Command("pkill", "-f", matchedApp)
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Failed to close application: %s\n", err)
		return
	}
	fmt.Printf("✅ Successfully closed application: %s\n", matchedApp)
	*memory = append(*memory, fmt.Sprintf("Closed application: %s", matchedApp))
}
