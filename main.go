package main

import (
	"ai-agent-go/internal/chat_interaction"
	"ai-agent-go/internal/command_func"
	"ai-agent-go/internal/folder_func"
	"ai-agent-go/internal/llm_tool"
	"ai-agent-go/internal/memory_func"
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	memory, err := memory_func.LoadMemory("memory.txt")
	if err != nil {
		fmt.Println("Cannot load memory:", err)
		memory = []string{}
	}

	fmt.Println("Hi user!")
	RunLoop(memory)
}

func RunLoop(memory []string) {
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
			chat_interaction.HandleAdd(input, &memory)
		case "Question":
			chat_interaction.HandleQuestion(input, memory, reader)
		case "Command":
			cmd, err := command_func.HandleCommand(input, memory)
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
			cmd, err := chat_interaction.HandleUpdate(input, memory)
			if err != nil {
				fmt.Println("Command Error:", err)
				continue
			}
			fmt.Println("Updating information...", cmd)
			continue
		case "DeleteFolder":
			folder_func.HandleDeleteFolder(input, reader, &memory)
			continue
		case "CreateFolder":
			folder_func.HandleCreateFolder(input, reader, &memory)
			continue
		default:
			fmt.Println("Agent: I didn't understand your intent.")
		}
	}
}

func ClassifyInput(input string) (string, error) {
	prompt := fmt.Sprintf(`
You are a classification bot.

Your job is to classify the user's message into **only one** of the following types:
- Question
- DeleteFolder 
- Add
- Command
- Update
- CreateFolder
- Unknown

Message:
%s

Return only the classification **as one of the exact words above**.
Do not explain.
Do not include quotes.
Do not include any tags like <think>.
Only return one word.`, input)

	resp, err := llm_tool.AskLLM(prompt)
	fmt.Println("Classifying input:", resp)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resp), nil
}
