package main

import (
	"ai-agent-go/internal/app_interaction"
	"ai-agent-go/internal/chat_interaction"
	"ai-agent-go/internal/command_func"
	"ai-agent-go/internal/folder_func"
	"ai-agent-go/internal/git_func"
	"ai-agent-go/internal/llm_tool"
	"ai-agent-go/internal/memory_func"
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {

	fmt.Println("Hi user!")
	RunLoop()
}

func RunLoop() {

	devOpsMemory, err := memory_func.LoadMemory("devOpsMemory.txt")
	if err != nil {
		fmt.Println("Cannot load devOpsMemory:", err)
		devOpsMemory = []string{}
	}

	personalMemory, err := memory_func.LoadMemory("personalMemory.txt")
	if err != nil {
		fmt.Println("Cannot load persionalMemory:", err)
		personalMemory = []string{}
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">>> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" || input == "quit" {
			fmt.Println("Bye user!")
			break
		}

		classType, err := ClassifyInput(input)
		if err != nil {
			fmt.Println("Error classifying input:", err)
			continue
		}

		switch classType {
		case "Add":
			chat_interaction.HandleAdd(input, &devOpsMemory)
		case "OperationSystemQuestion":
			chat_interaction.HandleQuestion(input, devOpsMemory, reader)
		case "Command":
			cmd, err := command_func.HandleCommand(input, devOpsMemory)
			if err != nil {
				fmt.Println("Command Error:", err)
				continue
			}
			fmt.Println("SSH Command:", cmd)
			fmt.Println("Executing SSH...")
			errSSH := command_func.OpenTerminalAndRunCommand(cmd)
			if errSSH != nil {
				fmt.Println("❌ Failed to open Terminal:", err)
			} else {
				fmt.Println("✅ SSH command sent to new Terminal window.")
			}
			continue
		case "Update":
			cmd, err := chat_interaction.HandleUpdate(input, devOpsMemory)
			if err != nil {
				fmt.Println("Command Error:", err)
				continue
			}
			fmt.Println("Updating information...", cmd)
			continue
		case "DeleteFolder":
			folder_func.HandleDeleteFolder(input, reader, &devOpsMemory)
			continue
		case "CreateFolder":
			folder_func.HandleCreateFolder(input, reader, &devOpsMemory)
			continue
		case "PersonalInformationAddition":
			chat_interaction.HandlePersionalInformationAdd(input, &personalMemory)
			continue
		case "PersonalInformationQuestion":
			chat_interaction.HandlePersionalInformationQuestion(input, personalMemory, reader)
			continue
		case "CloneGitRepositoryRequest":
			git_func.HandleGitCloneRequest(input, &devOpsMemory)
			continue
		case "PullGitRepositoryRequest":
			git_func.HanldeGitPullRequest(input, &devOpsMemory)
			continue
		case "OpenApplication":
			app_interaction.HandleAppOpenRequest(input, &devOpsMemory)
			continue
		case "CloseApplication":
			app_interaction.HandleAppCloseRequest(input, &devOpsMemory)
			continue
		default:
			fmt.Println("Agent: I didn't understand your intent.")
		}
	}
}

func ClassifyInput(input string) (string, error) {
	prompt := fmt.Sprintf(`
You are an AI assistant that classifies the user's message into only one of the following types:

### Informational (questions)
- OperationSystemQuestion → for questions about folders, files, Git repo info like "What is the URL of the Git repo?" or "Where is the Linux folder?"

### Action-based commands
- CloneGitRepositoryRequest → only if the user explicitly wants to clone a Git repository (e.g., "Clone the Git repo into X folder", "Download my project from GitHub")
- PullGitRepositoryRequest → only if the user explicitly wants to pull a Git repository (e.g., "Pull the latest changes from my Git repo", "Update my project from GitHub")

Other types:
- DeleteFolder
- OpenApplication
- CloseApplication
- Add
- Command
- Update
- CreateFolder
- PersonalInformationAddition
- PersonalInformationUpdate
- PersonalInformationQuestion
- PersonalInformationDelete
- Unknown

Classify this message:
%s

Return ONLY one of the above class names as plain text, no quotes or extra text.
`, input)

	resp, err := llm_tool.AskLLM(prompt)
	fmt.Println("Classifying input:", resp)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resp), nil
}
