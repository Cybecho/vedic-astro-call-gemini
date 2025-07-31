package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/genai"
)

type ChartRequest struct {
	Chart     interface{} `json:"chart"`
	Duration  float64     `json:"duration_of_response"`
	CreatedAt string      `json:"created_at"`
}

type Response struct {
	Interpretation string `json:"interpretation"`
	Success        bool   `json:"success"`
	Error          string `json:"error,omitempty"`
}

func loadPrompt() (string, error) {
	content, err := os.ReadFile("post-prompt.txt")
	if err != nil {
		return "", fmt.Errorf("failed to read post-prompt.txt: %v", err)
	}
	return string(content), nil
}

func generateInterpretation(ctx context.Context, client *genai.Client, prompt string, chartData interface{}) (string, error) {
	chartJSON, err := json.MarshalIndent(chartData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal chart data: %v", err)
	}

	fullPrompt := fmt.Sprintf("%s\n\n차트 데이터:\n%s", prompt, string(chartJSON))

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash-lite",
		genai.Text(fullPrompt),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %v", err)
	}

	return result.Text(), nil
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

	prompt, err := loadPrompt()
	if err != nil {
		response := Response{
			Success: false,
			Error:   "Failed to load prompt template",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		response := Response{
			Success: false,
			Error:   "Failed to create Gemini client",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	interpretation, err := generateInterpretation(ctx, client, prompt, chartReq.Chart)
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

	response := Response{
		Interpretation: interpretation,
		Success:        true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/interpret", handleChart)
	
	fmt.Println("베다 점성술 AI 해석 서비스가 9494 포트에서 시작되었습니다.")
	fmt.Println("엔드포인트: POST /interpret")
	
	log.Fatal(http.ListenAndServe(":9494", nil))
}