package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type ChartRequest struct {
	Chart        interface{} `json:"chart"`
	Duration     float64     `json:"duration_of_response"`
	CreatedAt    string      `json:"created_at"`
	CustomPrompt string      `json:"custom_prompt,omitempty"`
}

type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Response struct {
	Interpretation string      `json:"interpretation"`
	Success        bool        `json:"success"`
	Error          string      `json:"error,omitempty"`
	TokenUsage     *TokenUsage `json:"token_usage,omitempty"`
	ProcessingTime string      `json:"processing_time,omitempty"`
}

func loadPrompt() (string, error) {
	content, err := os.ReadFile("post-prompt.txt")
	if err != nil {
		return "", fmt.Errorf("failed to read post-prompt.txt: %v", err)
	}
	return string(content), nil
}


func handleChart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response := Response{
			Success: false,
			Error:   "Failed to read request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	var chartReq ChartRequest
	if err := json.Unmarshal(body, &chartReq); err != nil {
		response := Response{
			Success: false,
			Error:   "Invalid JSON format",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 커스텀 프롬프트가 있으면 사용, 없으면 기본 프롬프트 로드
	var prompt string
	
	if chartReq.CustomPrompt != "" {
		prompt = chartReq.CustomPrompt
	} else {
		promptContent, promptErr := loadPrompt()
		if promptErr != nil {
			response := Response{
				Success: false,
				Error:   "Failed to load prompt template",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
		prompt = promptContent
	}

	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// HTTP API를 직접 사용
	interpretation, tokenUsage, err := generateInterpretationHTTP(ctx, prompt, chartReq.Chart)
	if err != nil {
		response := Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to generate interpretation: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	processingTime := time.Since(startTime)

	response := Response{
		Interpretation: interpretation,
		Success:        true,
		TokenUsage:     tokenUsage,
		ProcessingTime: fmt.Sprintf("%.2fs", processingTime.Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func generateInterpretationHTTP(ctx context.Context, prompt string, chartData interface{}) (string, *TokenUsage, error) {
	apiKey := os.Getenv("GOOGLE_AI_API_KEY")
	if apiKey == "" {
		return "", nil, fmt.Errorf("Google AI API key not provided")
	}

	chartJSON, err := json.MarshalIndent(chartData, "", "  ")
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal chart data: %v", err)
	}

	fullPrompt := fmt.Sprintf("%s\n\nChart Data:\n%s", prompt, string(chartJSON))

	// Google AI API 요청 구조
	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": fullPrompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.7,
			"maxOutputTokens": 8192,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=%s", apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to call Google AI API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// 응답 파싱
	var apiResponse struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if len(apiResponse.Candidates) == 0 || len(apiResponse.Candidates[0].Content.Parts) == 0 {
		return "", nil, fmt.Errorf("no content in API response")
	}

	interpretation := apiResponse.Candidates[0].Content.Parts[0].Text
	tokenUsage := &TokenUsage{
		PromptTokens:     apiResponse.UsageMetadata.PromptTokenCount,
		CompletionTokens: apiResponse.UsageMetadata.CandidatesTokenCount,
		TotalTokens:      apiResponse.UsageMetadata.TotalTokenCount,
	}

	return strings.TrimSpace(interpretation), tokenUsage, nil
}

func main() {
	http.HandleFunc("/interpret", handleChart)
	
	fmt.Println("베다 점성술 AI 해석 서비스가 9494 포트에서 시작되었습니다.")
	fmt.Println("엔드포인트: POST /interpret")
	
	log.Fatal(http.ListenAndServe(":9494", nil))
}