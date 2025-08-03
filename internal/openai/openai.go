// Package openai provides a client for interacting with the OpenAI API
package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type OpenAI struct {
	apiKey             string
	transcriptionModel string
	speechModel        string
	chatModel          string
	voice              string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type AudioRequest struct {
	Model        string `json:"model"`
	Input        string `json:"input"`
	Voice        string `json:"voice"`
	Instructions string `json:"instructions,omitempty"`
}

type TranscriptionResponse struct {
	Text string `json:"text"`
}

func New(apiKey string, transcriptionModel string, speechModel string, chatModel string, voice string) *OpenAI {
	return &OpenAI{
		apiKey:             apiKey,
		transcriptionModel: transcriptionModel,
		speechModel:        speechModel,
		chatModel:          chatModel,
		voice:              voice,
	}
}

func (o *OpenAI) Transcribe(audioFile io.Reader) (string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add model field
	if err := writer.WriteField("model", o.transcriptionModel); err != nil {
		return "", fmt.Errorf("failed to write model field: %w", err)
	}

	formFile, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(formFile, audioFile); err != nil {
		return "", fmt.Errorf("failed to copy audio data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", &buf)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response body: %w", err)
		}
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var transcriptionResp TranscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&transcriptionResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close response body: %w", err)
	}

	return transcriptionResp.Text, nil
}

func (o *OpenAI) GenerateAudio(text string, instructions string) (io.Reader, error) {
	audioReq := AudioRequest{
		Model:        o.speechModel,
		Input:        text,
		Voice:        o.voice,
		Instructions: instructions,
	}

	jsonData, err := json.Marshal(audioReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/speech", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		err = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close response body: %w", err)
		}
		return nil, fmt.Errorf("API request failed with status: %d: %s", resp.StatusCode, string(body))
	}

	return resp.Body, nil
}

func (o *OpenAI) Chat(messages []Message) (string, error) {
	chatReq := ChatRequest{
		Model:    o.chatModel,
		Messages: messages,
	}

	jsonData, err := json.Marshal(chatReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var chatResp ChatResponse
	err = json.NewDecoder(resp.Body).Decode(&chatResp)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	err = resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close response body: %w", err)
	}

	return chatResp.Choices[0].Message.Content, nil
}
