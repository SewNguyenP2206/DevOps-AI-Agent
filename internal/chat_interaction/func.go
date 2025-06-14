package chat_interaction

import (
	"ai-agent-go/internal/llm_tool"
	"ai-agent-go/internal/memory_func"
	"bufio"
	"fmt"
	"os"
	"strings"
)

func HandleQuestion(input string, memory []string, reader *bufio.Reader) {
	// Bước 1: Yêu cầu AI tìm câu trả lời từ memory
	prompt := fmt.Sprintf(`
You are an AI assistant. The user asked: "%s"

This is currently stored in memory:
%s

Answer the question using only the information in memory.

`, input, strings.Join(memory, "\n"))

	resp, err := llm_tool.AskLLM(prompt)
	if err != nil {
		fmt.Println("❌ Error from AI:", err)
		return
	}
	resp = strings.TrimSpace(resp)

	// Nếu câu trả lời KHÔNG kết thúc bằng "is", coi như AI đã trả lời được
	if !strings.HasSuffix(resp, "is") && !strings.HasSuffix(resp, "is ") {
		fmt.Println("✅ Answer from memory:", resp)
		return
	}

	// Nếu chưa có trong memory → bổ sung
	fmt.Println("🤔 I don’t have this information in memory. Let’s add it!")
	fmt.Print(">>> Please enter the missing info: ")
	newInfo, _ := reader.ReadString('\n')
	newInfo = strings.TrimSpace(newInfo)

	fact := resp + " " + newInfo
	memory = append(memory, fact)

	fmt.Println("🧠 Memory updated with:", fact)

	if err := memory_func.SaveMemory("memory.txt", memory); err != nil {
		fmt.Println("❌ Failed to save memory:", err)
	}
}

// HandleUpdate: ask LLM to generate a new fact and replace old one in memory
func HandleUpdate(input string, memory []string) (string, error) {
	// For simplicity, just ask user for the new fact and replace the old one if found
	fmt.Println("What is the updated information?")
	reader := bufio.NewReader(os.Stdin)
	newFact, _ := reader.ReadString('\n')
	newFact = strings.TrimSpace(newFact)
	if newFact == "" {
		return "", fmt.Errorf("No update provided")
	}

	// Try to find and replace old fact
	updated := false
	for i, line := range memory {
		if strings.Contains(strings.ToLower(line), strings.ToLower(input)) {
			memory[i] = newFact
			updated = true
			break
		}
	}
	if !updated {
		memory = append(memory, newFact)
	}
	_ = memory_func.SaveMemory("memory.txt", memory)
	return "Updated: " + newFact, nil
}

// HandleAdd: extract fact(s) from input and append to memory as natural language
func HandleAdd(input string, memory *[]string) {
	prompt := fmt.Sprintf(`
You are an AI assistant that extracts facts from user input.
Extract all useful facts as natural language sentences, one per line.
If nothing useful, return an empty string.

User input:
%s

Example output:
Ip of Revoland ec2 instance is 123.123.123.123
The directory of Linux folder is /Users/sewn/Linux
`, input)

	resp, err := llm_tool.AskLLM(prompt)
	if err != nil {
		fmt.Println("Error extracting info:", err)
		return
	}

	lines := strings.Split(resp, "\n")
	added := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			*memory = append(*memory, line)
			fmt.Println("Memory updated:", line)
			added = true
		}
	}
	if added {
		_ = memory_func.SaveMemory("memory.txt", *memory)
	} else {
		fmt.Println("Nothing useful to store.")
	}
}
