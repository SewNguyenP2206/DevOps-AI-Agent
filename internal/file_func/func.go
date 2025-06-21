package file_func

import (
	"ai-agent-go/internal/llm_tool"
	"encoding/json"
	"fmt"
	"strings"
)

// create file with custom extension with LLM
func CreateFileWithExtension(input string, memory *[]string) (string, error) {
	prompt := fmt.Sprintf(`You are an AI assistant. Extract the file name, location to create and extension from this message:
"%s"
Return a JSON object like:
{  
  "file_name": "the-file-name-or-null",
  "extension": "the-extension"
  "location": "the-location-keyword-or-null"
}
If extension is not specified, use ".txt" as default.
Return only JSON, no explanation.
`, input)
	resp, err := llm_tool.AskLLM(prompt)
	if err != nil {
		return "", fmt.Errorf("error extracting file info: %v", err)
	}
	var requestResult struct {
		FileName  string `json:"file_name"`
		Extension string `json:"extension"`
		Location  string `json:"location"`
	}
	if err := json.Unmarshal([]byte(resp), &requestResult); err != nil {
		return "", fmt.Errorf("error parsing AI response: %v", err)
	}
	fileName := strings.TrimSpace(requestResult.FileName)
	if fileName == "null" || fileName == "" {
		return "", fmt.Errorf("file name cannot be empty")
	}
	extension := strings.TrimSpace(requestResult.Extension)
	if extension == "" || extension == "null" {
		extension = ".txt" // Default to .txt if no extension provided
	}
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension // Ensure it starts with a dot
	}
	location := strings.TrimSpace(requestResult.Location)
	if location == "null" || location == "" {
		location = "." // Default to current directory if no location provided
	}

}
