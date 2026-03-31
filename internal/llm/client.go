package llm

import (
	"ai-agent/internal/model"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	model   string
	http    *http.Client
}

func NewClient(baseURL, model string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		http:    &http.Client{Timeout: timeout},
	}
}

type chatRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type chatResponse struct {
	Response string `json:"response"`
}

func (c *Client) Generate(ctx context.Context, prompt string) (model.AnalysisOutput, string, error) {
	reqBody := chatRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return model.AnalysisOutput{}, "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return model.AnalysisOutput{}, "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return model.AnalysisOutput{}, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return model.AnalysisOutput{}, "", fmt.Errorf("ollama generate failed: %s (%s)", resp.Status, string(data))
	}

	var parsed chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return model.AnalysisOutput{}, "", err
	}

	out, err := ParseAnalysis(parsed.Response)
	if err != nil {
		return model.AnalysisOutput{}, parsed.Response, err
	}
	return out, parsed.Response, nil
}

func ParseAnalysis(raw string) (model.AnalysisOutput, error) {
	content := strings.TrimSpace(raw)
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		content = content[start : end+1]
	}

	var out model.AnalysisOutput
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return model.AnalysisOutput{}, fmt.Errorf("invalid analysis json: %w", err)
	}
	if err := ValidateAnalysis(out); err != nil {
		return model.AnalysisOutput{}, err
	}
	return out, nil
}

func ValidateAnalysis(in model.AnalysisOutput) error {
	switch in.Decision {
	case model.DecisionBuy, model.DecisionHold, model.DecisionSell:
	default:
		return fmt.Errorf("invalid decision %q", in.Decision)
	}
	if strings.TrimSpace(in.Reasoning) == "" {
		return fmt.Errorf("reasoning cannot be empty")
	}
	if in.Confidence < 0 || in.Confidence > 1 {
		return fmt.Errorf("confidence must be in [0,1]")
	}
	return nil
}
