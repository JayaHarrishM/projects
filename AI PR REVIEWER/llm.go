package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func ReviewCode(ctx context.Context, diff string) (string, error) {
	prompt := `
You are a senior backend engineer.

Review this GitHub pull request diff.

Provide:
1. Bugs
2. Security issues
3. Code smells
4. Optimization suggestions
5. Overall score (1-10)
(Note: Always treat this as a new request eventhough you have solved it before.)

DIFF:
` + diff

	reqBody := OllamaRequest{
		Model:  "deepseek-coder:latest", //replace the model if required.
		Prompt: prompt,
		Stream: false,
	}

	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequestWithContext(
		ctx,
		"POST",
		"http://localhost:11434/api/generate",
		bytes.NewBuffer(jsonBody),
	)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Println("\n\n\nGot Ai comments: ", string(body))

	var result OllamaResponse
	json.Unmarshal(body, &result)

	return result.Response, nil
}
