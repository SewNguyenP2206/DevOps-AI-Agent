package folder_func

import (
	"ai-agent-go/internal/llm_tool"
	"ai-agent-go/internal/memory_func"
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func HandleCreateFolder(input string, reader *bufio.Reader, memory *[]string) {
	// Try to extract folder name and location using LLM
	prompt := fmt.Sprintf(`
You are an AI assistant. Extract the folder name and location from this message:

"%s"

Return a JSON object like:
{
  "folder_name": "the-folder-name",
  "location": "the-location-keyword-or-null"
}

If location is not specified or not found in memory, use null. If folder name is missing, leave it blank.
Return only JSON, no explanation.
`, input)

	resp, err := llm_tool.AskLLM(prompt)
	if err != nil {
		fmt.Println("Error extracting folder info:", err)
		return
	}

	print("AI response:", resp)

	var folderName, location string
	type folderExtract struct {
		FolderName string `json:"folder_name"`
		Location   string `json:"location"`
	}
	var result folderExtract
	if err := json.Unmarshal([]byte(resp), &result); err == nil {
		folderName = strings.TrimSpace(result.FolderName)
		location = strings.TrimSpace(result.Location)
	}

	if location == "" || location == "null" {
		fmt.Println("❗ Location not specified. What is the name of the location you want to create the folder in ?")
		fmt.Print(">>> Enter the location keyword: ")
		locationInput, _ := reader.ReadString('\n')
		location = strings.TrimSpace(locationInput)
		if location == "" {
			fmt.Println("❌ Location cannot be empty.")
			return
		}
	}

	// Resolve location to absolute path from memory
	absPath := ""
	if location != "" && location != "null" {
		for _, line := range *memory {
			if strings.HasSuffix(strings.ToLower(line), "directory is "+strings.ToLower(location)) ||
				strings.HasSuffix(strings.ToLower(line), "folder is "+strings.ToLower(location)) ||
				strings.Contains(strings.ToLower(line), "directory of "+strings.ToLower(location)+" is") ||
				strings.Contains(strings.ToLower(line), "folder of "+strings.ToLower(location)+" is") {
				absPath = extractValueFromFact(line)
				break
			}
		}
	}

	// Nếu chưa có absolute path, hỏi user
	if absPath == "" {
		fmt.Printf("❓ I don't know where \"%s\" is. Please provide the full path: ", location)
		inputPath, _ := reader.ReadString('\n')
		inputPath = strings.TrimSpace(inputPath)
		if inputPath == "" {
			fmt.Println("❌ Path cannot be empty.")
			return
		}
		absPath = inputPath
		// Lưu vào memory
		fact := fmt.Sprintf("The directory of %s is %s", location, absPath)
		*memory = append(*memory, fact)
		_ = memory_func.SaveMemory("memory.txt", *memory)
		fmt.Printf("✅ \"%s\" path saved to memory.\n", location)
	}

	// Nếu thiếu folder name, hỏi user
	if folderName == "" || folderName == "null" {
		fmt.Print("📂 What should the folder be called? ")
		folderName, _ = reader.ReadString('\n')
		folderName = strings.TrimSpace(folderName)
		if folderName == "" {
			fmt.Println("❌ Folder name cannot be empty.")
			return
		}
	}

	fullPath := filepath.Join(absPath, folderName)
	fmt.Println("📁 Full path to create:", fullPath)

	// Kiểm tra xem đã có fact này trong memory chưa
	fact := fmt.Sprintf("The directory of %s is %s", folderName, fullPath)
	existed := false
	for _, line := range *memory {
		if strings.TrimSpace(line) == fact {
			existed = true
			break
		}
	}

	// Tạo folder nếu chưa tồn tại trên ổ đĩa
	if _, err := os.Stat(fullPath); err == nil {
		fmt.Println("⚠️ (AI) Folder already exists at:", fullPath)
	} else {
		cmd := exec.Command("mkdir", "-p", fullPath)
		if err := cmd.Run(); err != nil {
			fmt.Println("❌ Failed to create folder:", err)
			return
		}
		fmt.Println("✅ Folder created successfully at:", fullPath)
	}

	// Nếu chưa có trong memory thì mới lưu
	if !existed {
		*memory = append(*memory, fact)
		if err := memory_func.SaveMemory("memory.txt", *memory); err != nil {
			fmt.Println("❌ Error saving folder info to memory:", err)
		} else {
			fmt.Printf("✅ Memory updated: %s\n", fact)
		}
	}
}

func HandleDeleteFolder(input string, reader *bufio.Reader, memory *[]string) {
	// Try to extract folder name and location using LLM
	prompt := fmt.Sprintf(`
You are an AI assistant. Extract the folder name and location from this message:
"%s"
Return a JSON object like:
{
  "folder_name": "the-folder-name",
  "location": "the-location-keyword-or-null"
}																							
If location is not specified or not found in memory, use null. If folder name is missing, leave it blank.
Return only JSON, no explanation.
`, input)
	resp, err := llm_tool.AskLLM(prompt)
	if err != nil {
		fmt.Println("Error extracting folder info:", err)
		return
	}
	print("AI response:", resp)
	var folderName, location string
	type folderExtract struct {
		FolderName string `json:"folder_name"`
		Location   string `json:"location"`
	}
	var result folderExtract
	if err := json.Unmarshal([]byte(resp), &result); err == nil {
		folderName = strings.TrimSpace(result.FolderName)
		location = strings.TrimSpace(result.Location)
	}
	if location == "" || location == "null" {
		fmt.Println("❗ Location not specified. What is the name of the location you want to delete the folder from?")
		fmt.Print(">>> Enter the location keyword: ")
		locationInput, _ := reader.ReadString('\n')
		location = strings.TrimSpace(locationInput)
		if location == "" {
			fmt.Println("❌ Location cannot be empty.")
			return
		}
	}
	// Resolve location to absolute path from memory
	absPath := ""
	if location != "" && location != "null" {
		for _, line := range *memory {
			if strings.HasSuffix(strings.ToLower(line), "directory is "+strings.ToLower(location)) ||
				strings.HasSuffix(strings.ToLower(line), "folder is "+strings.ToLower(location)) ||
				strings.Contains(strings.ToLower(line), "directory of "+strings.ToLower(location)+" is") ||
				strings.Contains(strings.ToLower(line), "folder of "+strings.ToLower(location)+" is") {
				absPath = extractValueFromFact(line)
				break
			}
		}
	}
	// Nếu chưa có absolute path, hỏi user
	if absPath == "" {
		fmt.Printf("❓ I don't know where \"%s\" is. Please provide the full path: ", location)
		inputPath, _ := reader.ReadString('\n')
		inputPath = strings.TrimSpace(inputPath)
		if inputPath == "" {
			fmt.Println("❌ Path cannot be empty.")
			return
		}
		absPath = inputPath
		// Lưu vào memory
		fact := fmt.Sprintf("The directory of %s is %s", location, absPath)
		*memory = append(*memory, fact)
		if err := memory_func.SaveMemory("memory.txt", *memory); err != nil {
			fmt.Println("❌ Error saving folder info to memory:", err)
		} else {
			fmt.Printf("✅ \"%s\" path saved to memory.\n", location)
		}
	}
	// Nếu thiếu folder name, hỏi user
	if folderName == "" || folderName == "null" {
		fmt.Print("📂 What is the name of the folder to delete? ")
		folderName, _ = reader.ReadString('\n')
		folderName = strings.TrimSpace(folderName)
		if folderName == "" {
			fmt.Println("❌ Folder name cannot be empty.")
			return
		}
	}
	fullPath := filepath.Join(absPath, folderName)
	fmt.Println("📁 Full path to delete:", fullPath)
	// Kiểm tra xem đã có fact này trong memory chưa
	fact := fmt.Sprintf("The directory of %s is %s", folderName, fullPath)
	existed := false
	for _, line := range *memory {
		if strings.TrimSpace(line) == fact {
			existed = true
			break
		}
	}
	// Xóa folder nếu tồn tại trên ổ đĩa
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		fmt.Println("⚠️ (AI) Folder does not exist at:", fullPath)
	} else {
		cmd := exec.Command("rm", "-rf", fullPath)
		if err := cmd.Run(); err != nil {
			fmt.Println("❌ Failed to delete folder:", err)
			return
		}
		fmt.Println("✅ Folder deleted successfully at:", fullPath)
	}
	// Nếu đã có trong memory thì xóa
	if existed {
		for i, line := range *memory {
			if strings.TrimSpace(line) == fact {
				*memory = append((*memory)[:i], (*memory)[i+1:]...)
				break
			}
		}
		if err := memory_func.SaveMemory("memory.txt", *memory); err != nil {
			fmt.Println("❌ Error saving updated memory:", err)
		} else {
			fmt.Printf("✅ Memory updated: %s deleted\n", fact)
		}
	} else {
		fmt.Println("❗ Folder info not found in memory, no update needed.")
	}
}

// Helper: extract value (after "is") from a fact sentence
func extractValueFromFact(line string) string {
	parts := strings.Split(line, " is ")
	if len(parts) > 1 {
		return strings.TrimSpace(parts[len(parts)-1])
	}
	return ""
}
