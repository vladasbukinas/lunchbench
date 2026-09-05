package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ardanlabs/kronk/sdk/kronk/model"
)

const chatPath = "/v1/chat/completions"

const systemPrompt = `You are a lunch concierge. You are given today's restaurant
menu data, including how far each restaurant is (walking minutes), and a task.
Answer only from the provided data, be concise, and do not invent restaurants,
prices, or distances. When asked to decide, weigh both the food and the
walking distance.`

type task struct {
	name string
	ask  string
}

func tasks() []task {
	return []task{
		{name: "budget-list", ask: "List every lunch special priced under $12.00 with its restaurant name and exact price."},
		{name: "cheapest", ask: "Which single lunch special across all restaurants is the cheapest, and what does it include?"},
		{name: "vegetarian", ask: "Recommend the best vegetarian-friendly lunch under $11.00, weighing the food and the walking distance, in exactly three sentences."},
		{name: "summary", ask: "Summarize each restaurant in one short line including its price range from cheapest to priciest special and its walking time."},
		{name: "quick-pick", ask: "You have only 30 minutes for lunch today. Pick one restaurant and one special, and explain in two sentences how distance and price shaped your decision."},
	}
}

type bench struct {
	url     string
	modelID string
	bearer  string
	think   bool
	menuTxt string
}

func newBench(server string, bearer string, modelID string, think bool, menuTxt string) *bench {
	return &bench{
		url:     strings.TrimRight(server, "/") + chatPath,
		modelID: modelID,
		bearer:  bearer,
		think:   think,
		menuTxt: menuTxt,
	}
}

type runResult struct {
	task             string
	model            string
	wall             time.Duration
	ttft             time.Duration
	promptTokens     int
	completionTokens int
	reasoningTokens  int
	totalTokens      int
	serverTPS        float64
	serverTTFTMS     float64
	answer           string
}

func (b *bench) run(ctx context.Context, t task, maxTokens int) (runResult, error) {
	res := runResult{task: t.name}

	messages := []model.D{
		{"role": "system", "content": systemPrompt},
		{"role": "user", "content": "MENU DATA:\n" + b.menuTxt + "\nTASK: " + t.ask},
	}

	reqBody, err := json.Marshal(model.D{
		"model":          b.modelID,
		"messages":       messages,
		"max_tokens":     maxTokens,
		"temperature":    0.2,
		"stream":         true,
		"stream_options": model.D{"include_usage": true},

		"chat_template_kwargs": model.D{"enable_thinking": b.think},
	})
	if err != nil {
		return res, fmt.Errorf("run: encoding request: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.url, strings.NewReader(string(reqBody)))
	if err != nil {
		return res, fmt.Errorf("run: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if b.bearer != "" {
		req.Header.Set("Authorization", "Bearer "+b.bearer)
	}

	start := time.Now()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return res, fmt.Errorf("run: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return res, fmt.Errorf("run: server returned status %d", resp.StatusCode)
	}

	var answer strings.Builder
	var reasoning strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	first := true

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			break
		}

		var cr model.ChatResponse
		if err := json.Unmarshal([]byte(payload), &cr); err != nil {
			return res, fmt.Errorf("run: decoding SSE chunk: %w", err)
		}

		res.model = cr.Model

		if len(cr.Choices) == 0 {
			if cr.Usage != nil {
				res.promptTokens = cr.Usage.PromptTokens
				res.completionTokens = cr.Usage.CompletionTokens
				res.reasoningTokens = cr.Usage.CompletionTokensDetails.ReasoningTokens
				res.totalTokens = cr.Usage.TotalTokens
				res.serverTPS = cr.Usage.TokensPerSecond
				res.serverTTFTMS = cr.Usage.TimeToFirstTokenMS
			}
			continue
		}

		c := cr.Choices[0]

		switch c.FinishReason() {
		case model.FinishReasonError:
			return res, fmt.Errorf("run: model error: %s", c.Delta.Content)

		case model.FinishReasonStop, model.FinishReasonLength, model.FinishReasonTool:
			continue
		}

		if c.Delta == nil {
			continue
		}

		if first && (c.Delta.Content != "" || c.Delta.Reasoning != "") {
			res.ttft = time.Since(start)
			first = false
		}

		if c.Delta.Reasoning != "" {
			reasoning.WriteString(c.Delta.Reasoning)
		}

		if c.Delta.Content != "" {
			answer.WriteString(c.Delta.Content)
		}
	}

	if err := scanner.Err(); err != nil {
		return res, fmt.Errorf("run: reading stream: %w", err)
	}

	res.wall = time.Since(start)
	res.answer = strings.TrimSpace(answer.String())
	if res.answer == "" {
		if txt := strings.TrimSpace(reasoning.String()); txt != "" {
			if len(txt) > 600 {
				txt = "..." + txt[len(txt)-600:]
			}
			res.answer = "(thinking only) " + txt
		}
	}

	return res, nil
}
