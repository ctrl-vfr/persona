// Package openai provides a thin client for the subset of the OpenAI
// HTTP API used by persona: chat completions, audio transcriptions and
// audio TTS. It is intentionally hand-rolled (no SDK) to keep the
// dependency surface small.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

const (
	// defaultTimeout caps individual HTTP requests. TTS and chat both
	// usually complete well under 30s; 60s leaves headroom for slow
	// networks without letting the UI hang forever.
	defaultTimeout = 60 * time.Second

	// maxErrorBodyBytes limits how much of an error response body we
	// surface. Avoids dumping pages of HTML or large JSON into terminal
	// output and into wrapped error chains.
	maxErrorBodyBytes = 512

	// defaultBaseURL is the OpenAI API root. Overridable via WithBaseURL
	// to point tests at an httptest.Server.
	defaultBaseURL = "https://api.openai.com"
)

// OpenAI is a small client over the REST API.
type OpenAI struct {
	apiKey             string
	transcriptionModel string
	speechModel        string
	chatModel          string
	voice              string
	baseURL            string
	httpClient         *http.Client
}

// Option configures an OpenAI client.
type Option func(*OpenAI)

// WithBaseURL overrides the API root, useful in tests.
func WithBaseURL(u string) Option {
	return func(o *OpenAI) { o.baseURL = u }
}

// WithHTTPClient injects a custom HTTP client (e.g. with custom transport).
func WithHTTPClient(c *http.Client) Option {
	return func(o *OpenAI) { o.httpClient = c }
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

// New constructs an OpenAI client. apiKey is required; an empty key is
// reported back the first time a method is called, not here, so tests
// that swap out the HTTP client can still construct a client.
func New(apiKey, transcriptionModel, speechModel, chatModel, voice string, opts ...Option) *OpenAI {
	o := &OpenAI{
		apiKey:             apiKey,
		transcriptionModel: transcriptionModel,
		speechModel:        speechModel,
		chatModel:          chatModel,
		voice:              voice,
		baseURL:            defaultBaseURL,
		httpClient:         &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// errorFromResponse builds an error from a non-2xx HTTP response,
// truncating the body so we never leak large payloads into logs.
func errorFromResponse(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyBytes+1))
	snippet := body
	suffix := ""
	if len(snippet) > maxErrorBodyBytes {
		snippet = snippet[:maxErrorBodyBytes]
		suffix = "...(truncated)"
	}
	return fmt.Errorf("openai API status %d: %s%s", resp.StatusCode, string(snippet), suffix)
}

// Transcribe sends an audio reader to /v1/audio/transcriptions.
func (o *OpenAI) Transcribe(ctx context.Context, audioFile io.Reader) (string, error) {
	if o.apiKey == "" {
		return "", errors.New("openai: API key not set")
	}
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	if err := writer.WriteField("model", o.transcriptionModel); err != nil {
		return "", fmt.Errorf("write model field: %w", err)
	}
	formFile, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(formFile, audioFile); err != nil {
		return "", fmt.Errorf("copy audio data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/v1/audio/transcriptions", &buf)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", errorFromResponse(resp)
	}

	var transcriptionResp TranscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&transcriptionResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	return transcriptionResp.Text, nil
}

// GenerateAudio synthesises text to speech and returns the raw audio
// bytes. We deliberately read the body fully here rather than returning
// an io.Reader because callers historically forgot to close the body,
// leaking connections.
func (o *OpenAI) GenerateAudio(ctx context.Context, text, instructions string) ([]byte, error) {
	if o.apiKey == "" {
		return nil, errors.New("openai: API key not set")
	}
	audioReq := AudioRequest{
		Model:        o.speechModel,
		Input:        text,
		Voice:        o.voice,
		Instructions: instructions,
	}
	jsonData, err := json.Marshal(audioReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/v1/audio/speech", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, errorFromResponse(resp)
	}

	audio, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read audio body: %w", err)
	}
	return audio, nil
}

// Chat sends a conversation to /v1/chat/completions and returns the
// first choice's content.
func (o *OpenAI) Chat(ctx context.Context, messages []Message) (string, error) {
	if o.apiKey == "" {
		return "", errors.New("openai: API key not set")
	}
	chatReq := ChatRequest{Model: o.chatModel, Messages: messages}
	jsonData, err := json.Marshal(chatReq)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", errorFromResponse(resp)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return "", errors.New("openai: empty choices in response")
	}
	return chatResp.Choices[0].Message.Content, nil
}
