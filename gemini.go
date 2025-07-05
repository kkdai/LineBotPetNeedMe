// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// SearchCriteria represents the criteria for searching pets.
type SearchCriteria struct {
	Kind     string `json:"kind,omitempty"`
	Sex      string `json:"sex,omitempty"`
	BodyType string `json:"body_type,omitempty"`
	Age      string `json:"age,omitempty"`
	Color    string `json:"color,omitempty"`
}

var genaiClient *genai.GenerativeModel

func init() {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Println("Warning: GOOGLE_API_KEY is not set. Gemini functionality will be disabled.")
		return
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create genai client: %v", err)
	}
	genaiClient = client.GenerativeModel("gemini-1.5-flash")
}

// ParseSearchCriteriaFromQuery uses Gemini to parse the user's query.
func ParseSearchCriteriaFromQuery(query string) (*SearchCriteria, error) {
	if genaiClient == nil {
		return nil, nil // Gemini is not initialized
	}

	prompt := `
You are a pet adoption assistant. Your task is to analyze the user's request and extract search criteria for finding a pet.
The user's request is: "` + query + `"

Based on the request, identify the following criteria:
- kind: "貓" or "狗"
- sex: "公" or "母"
- body_type: "小型", "中型", or "大型"
- age: "幼年", "成年"
- color: "白", "黑", "黃", "棕", "灰", "虎斑", "三花", "其他"

Return the criteria as a JSON object. If a criterion is not mentioned, omit it from the JSON.
For example, if the user says "我想找一隻小隻的母狗", you should return:
{
  "kind": "狗",
  "sex": "母",
  "body_type": "小型"
}
If the user says "有貓嗎", you should return:
{
  "kind": "貓"
}
If the user's query is not related to finding a pet, return an empty JSON object {}.
`

	ctx := context.Background()
	resp, err := genaiClient.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, nil
	}

	jsonString := ""
	if part, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		jsonString = string(part)
	}

	// Clean the JSON string
	jsonString = cleanJSONString(jsonString)

	var criteria SearchCriteria
	if err := json.Unmarshal([]byte(jsonString), &criteria); err != nil {
		log.Printf("Failed to unmarshal JSON from Gemini: %v, raw: %s", err, jsonString)
		return nil, nil // Could not parse, treat as no criteria
	}

	// If all fields are empty, it means no criteria were found.
	if criteria.Kind == "" && criteria.Sex == "" && criteria.BodyType == "" && criteria.Age == "" {
		return nil, nil
	}

	return &criteria, nil
}

func cleanJSONString(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	return s
}
