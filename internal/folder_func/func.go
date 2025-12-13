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
	"strconv"
	"strings"
)

func HandleCreateFolder(input string, reader *bufio.Reader, memory *[]string) {
	// Step 1 — AI extract folder and location
	prompt := fmt.Sprintf(`
You are an AI assistant. Extract the folder name and location from this message:

"%s"

Return a JSON object like:
{
  "folder_name": "the-folder-name",
  "location": "the-location-keyword-or-null"
}

If folder name not found → blank.
If location missing → null.
Return ONLY JSON.
`, input)

	resp, err := llm_tool.AskLLM(prompt)
	if err != nil {
		fmt.Println("Error extracting folder info:", err)
		return
	}

	type Extract struct {
		FolderName string `json:"folder_name"`
		Location   string `json:"location"`
	}

	var ex Extract
	if err := json.Unmarshal([]byte(resp), &ex); err != nil {
		fmt.Println("❌ AI JSON parse failed:", err)
		return
	}

	folderName := strings.TrimSpace(ex.FolderName)
	location := strings.TrimSpace(ex.Location)

	// Ask location if AI couldn't detect
	if location == "" || location == "null" {
		fmt.Println("❗ Location not specified. What is the name of the location you want to create the folder in?")
		fmt.Print(">>> Enter location keyword: ")
		locInput, _ := reader.ReadString('\n')
		location = strings.TrimSpace(locInput)
		if location == "" {
			fmt.Println("❌ Location cannot be empty.")
			return
		}
	}

	// -------------------------------------------------------------
	// NEW MEMORY SYSTEM — folder:<name>|parent:<parent>|path:<path>
	// -------------------------------------------------------------

	// Helper: parse memory entry
	type FolderFact struct {
		Folder string
		Parent string
		Path   string
	}

	parseFact := func(line string) (*FolderFact, bool) {
		parts := strings.Split(line, "|")
		if len(parts) != 3 {
			return nil, false
		}
		f := &FolderFact{}
		for _, p := range parts {
			kv := strings.SplitN(p, ":", 2)
			if len(kv) != 2 {
				continue
			}
			k := strings.TrimSpace(kv[0])
			v := strings.TrimSpace(kv[1])
			switch k {
			case "folder":
				f.Folder = v
			case "parent":
				f.Parent = v
			case "path":
				f.Path = v
			}
		}
		return f, true
	}

	// Find all folder matches by name
	location = strings.TrimSpace(strings.Trim(location, "\""))
	resolveLocation := func(name string) []*FolderFact {
		var matches []*FolderFact
		for _, line := range *memory {
			if fact, ok := parseFact(line); ok {
				if strings.EqualFold(fact.Folder, name) {
					matches = append(matches, fact)
				}
			}
		}
		return matches
	}

	// Ask user to choose from multiple matches
	chooseFolder := func(matches []*FolderFact) *FolderFact {
		fmt.Println("❓ Multiple folders found with name:", matches[0].Folder)
		for i, f := range matches {
			fmt.Printf("%d) %s/%s → %s\n", i+1, f.Parent, f.Folder, f.Path)
		}
		fmt.Print(">>> Choose number: ")
		raw, _ := reader.ReadString('\n')
		raw = strings.TrimSpace(raw)
		index, _ := strconv.Atoi(raw)
		if index < 1 || index > len(matches) {
			fmt.Println("❌ Invalid choice.")
			return nil
		}
		return matches[index-1]
	}

	// --- Resolve parent folder path ---
	var absPath string
	matches := resolveLocation(location)

	if len(matches) == 0 {
		// Not found — ask user full path
		fmt.Printf("❓ I don't know where \"%s\" is. Please provide the full absolute path: ", location)
		pathInput, _ := reader.ReadString('\n')
		pathInput = strings.TrimSpace(pathInput)
		if pathInput == "" {
			fmt.Println("❌ Path cannot be empty.")
			return
		}

		absPath = pathInput

		// Save new memory
		fact := fmt.Sprintf("folder:%s|parent:%s|path:%s", location, "__root__", absPath)
		*memory = append(*memory, fact)
		memory_func.SaveMemory("memory.txt", *memory)
		fmt.Println("✅ Saved location to memory.")
	} else if len(matches) == 1 {
		absPath = matches[0].Path
	} else {
		chosen := chooseFolder(matches)
		if chosen == nil {
			return
		}
		absPath = chosen.Path
	}

	// Ask for folder name if missing
	if folderName == "" {
		fmt.Print("📂 What should the new folder be called? ")
		nameInput, _ := reader.ReadString('\n')
		folderName = strings.TrimSpace(nameInput)
		if folderName == "" {
			fmt.Println("❌ Folder name cannot be empty.")
			return
		}
	}

	// --- Create folder ---
	fullPath := filepath.Join(absPath, folderName)
	fmt.Println("📁 Full path:", fullPath)

	// Check existence
	if _, err := os.Stat(fullPath); err == nil {
		fmt.Println("⚠️ Folder already exists.")
	} else {
		cmd := exec.Command("mkdir", "-p", fullPath)
		if err := cmd.Run(); err != nil {
			fmt.Println("❌ Failed to create folder:", err)
			return
		}
		fmt.Println("✅ Folder created successfully.")
	}

	// Save to memory uniquely
	fact := fmt.Sprintf("folder:%s|parent:%s|path:%s", folderName, location, fullPath)

	exists := false
	for _, line := range *memory {
		if line == fact {
			exists = true
			break
		}
	}

	if !exists {
		*memory = append(*memory, fact)
		memory_func.SaveMemory("memory.txt", *memory)
		fmt.Println("💾 Memory updated.")
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
Look for the directory of the location used mentioned,if location is not specified or not found in memory, use null for location. If folder name is missing, leave it blank.
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
