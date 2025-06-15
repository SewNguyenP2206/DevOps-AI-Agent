package chat_interaction

import (
	"ai-agent-go/internal/llm_tool"
	"ai-agent-go/internal/memory_func"
	"bufio"
	"encoding/json"
	"fmt"
	"path/filepath"
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
	prompt := fmt.Sprintf(`
You are an AI assistant. The user asked: "%s"
Here is the current memory:
%s

Determine which fact in memory the user wants to update. 
Return a JSON object with two fields:
- "old_fact": the exact string in memory that should be replaced
- "new_fact": the updated version that replaces it, with the new information from the user

Respond only with the JSON. Do not include any explanation.
`, input, strings.Join(memory, "\n"))

	resp, err := llm_tool.AskLLM(prompt)
	if err != nil {
		return "", fmt.Errorf("error from AI: %w", err)
	}
	fmt.Println("🤖 AI response:", resp)

	type FactUpdate struct {
		OldFact string `json:"old_fact"`
		NewFact string `json:"new_fact"`
	}

	var update FactUpdate
	if err := json.Unmarshal([]byte(resp), &update); err != nil {
		return "", fmt.Errorf("failed to parse AI JSON: %w", err)
	}

	// Tìm và cập nhật trong memory
	updated := false
	for i, fact := range memory {
		if strings.Contains(fact, update.OldFact) {
			// Lấy tên thư mục từ path mới
			folderName := filepath.Base(update.NewFact)
			newSentence := fmt.Sprintf("The directory of %s is %s", folderName, update.NewFact)

			memory[i] = newSentence
			updated = true
			break
		}
	}

	if !updated {
		return "", fmt.Errorf("old fact not found in memory")
	}

	// Lưu lại file memory
	if err := memory_func.SaveMemory("memory.txt", memory); err != nil {
		return "", fmt.Errorf("failed to save memory: %w", err)
	}

	fmt.Println("🧠 Memory updated with new fact:", update.NewFact)
	return update.NewFact, nil
}

// HandleAdd: extract fact(s) from input and append to memory as natural language
func HandleAdd(input string, memory *[]string) {
	prompt := fmt.Sprintf(`
You are an AI assistant that extracts facts from user input.
Extract all useful facts as natural language sentences, one per line.
You are working on Mac/Linux operating system.ß
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

func HandlePersionalInformationAdd(input string, memory *[]string) {
	// Extract personal information from input
	prompt := fmt.Sprintf(`
You are an AI assistant that extracts personal information from user input.
Extract all useful personal information as natural language sentences, one per line.
User input:
%s
Example output:
User's name is John Doe 
User's email is 123@gmail.com 
`, input)
	resp, err := llm_tool.AskLLM(prompt)
	if err != nil {
		fmt.Println("Error extracting personal info:", err)
		return
	}
	lines := strings.Split(resp, "\n")
	added := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			*memory = append(*memory, line)
			fmt.Println("Memory updated with personal info:", line)
			added = true
		}
	}
	if added {
		if err := memory_func.SaveMemory("personalMemory.txt", *memory); err != nil {
			fmt.Println("❌ Failed to save personal information to memory:", err)
		} else {
			fmt.Println("✅ Personal information saved successfully.")
		}
	} else {
		fmt.Println("No useful personal information found.")
	}
}

func HandlePersionalInformationQuestion(input string, memory []string, reader *bufio.Reader) {
	// Ask LLM to answer the personal information question
	prompt := fmt.Sprintf(`
You are an AI assistant. The user asked: "%s"
This is the personal information stored in memory:
%s
Answer the question using only the information in memory. If the answer is not available, respond with: "NOT_FOUND".
`, input, strings.Join(memory, "\n"))

	resp, err := llm_tool.AskLLM(prompt)
	if err != nil {
		fmt.Println("❌ Error from AI:", err)
		return
	}

	resp = strings.TrimSpace(resp)
	if strings.EqualFold(resp, "NOT_FOUND") {
		fmt.Println("🤔 I don’t have this information in memory. Let’s add it!")
		fmt.Print(">>> Please enter the missing personal info: ")
		newInfo, _ := reader.ReadString('\n')
		newInfo = strings.TrimSpace(newInfo)

		// Ask LLM to generate a memory-friendly fact
		addPrompt := fmt.Sprintf(`
You are an AI assistant. The user asked: "%s"
They responded with: "%s"
Please turn this into a fact in the form: "The <type> is <value>"
Return only the fact sentence.
`, input, newInfo)

		fact, err := llm_tool.AskLLM(addPrompt)
		if err != nil {
			fmt.Println("❌ Failed to create memory fact:", err)
			return
		}
		fact = strings.TrimSpace(fact)
		if fact != "" {
			memory = append(memory, fact)
			fmt.Println("🧠 Personal information memory updated with:", fact)
			if err := memory_func.SaveMemory("memory.txt", memory); err != nil {
				fmt.Println("❌ Failed to save personal information memory:", err)
			} else {
				fmt.Println("✅ Personal information memory saved successfully.")
			}
		} else {
			fmt.Println("❌ Could not generate a memory fact.")
		}
		return
	}

	// Else: AI was able to answer based on memory
	fmt.Println("✅ Answer from personal information:", resp)
}
