package llm

import (
	"encoding/json"
	"log"
	"strings"
)

// ParseJSONResponse parses a JSON response from an LLM, handling markdown code blocks.
func ParseJSONResponse(text string) map[string]any {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	// Strip markdown code fences
	if strings.HasPrefix(text, "```") {
		lines := strings.Split(text, "\n")
		endIdx := len(lines) - 1
		for i := len(lines) - 1; i > 0; i-- {
			if strings.TrimSpace(lines[i]) == "```" {
				endIdx = i
				break
			}
		}
		text = strings.Join(lines[1:endIdx], "\n")
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		log.Printf("Failed to parse LLM response as JSON: %v", err)
		return nil
	}

	return result
}
