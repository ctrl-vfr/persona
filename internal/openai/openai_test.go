package openai

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestClient builds a client pointed at the given test server.
// Bypasses the real OpenAI endpoint and the 60s default timeout.
func newTestClient(t *testing.T, srv *httptest.Server) *OpenAI {
	t.Helper()
	return New("test-key", "whisper-1", "tts-1", "gpt-test", "alloy",
		WithBaseURL(srv.URL),
		WithHTTPClient(&http.Client{Timeout: 5 * time.Second}),
	)
}

func TestChat_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing or wrong Authorization header: %q", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"role":"assistant","content":"hello world"}}]}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestChat_ErrorTruncatesBody(t *testing.T) {
	bigBody := strings.Repeat("x", 4096)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, bigBody)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "truncated") {
		t.Errorf("expected truncation marker in error, got %q", err.Error())
	}
	// Sanity: the full 4KB body must NOT be in the error message.
	if len(err.Error()) > 1024 {
		t.Errorf("error message too long (%d bytes), expected truncation", len(err.Error()))
	}
}

func TestChat_ContextCancellation(t *testing.T) {
	// done lets the handler exit at the end of the test so srv.Close()
	// does not block waiting for in-flight handlers. r.Context().Done()
	// alone is not always signalled when the client cancels. Defer
	// order matters: close(done) MUST run before srv.Close(), so
	// register them in this order (LIFO).
	done := make(chan struct{})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-done:
		}
	}))
	defer srv.Close()
	defer close(done)

	c := newTestClient(t, srv)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.Chat(ctx, []Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestChat_EmptyAPIKey(t *testing.T) {
	c := New("", "whisper-1", "tts-1", "gpt-test", "alloy")
	_, err := c.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected error on missing API key")
	}
}

func TestGenerateAudio_Success(t *testing.T) {
	expected := []byte{0xff, 0xfb, 0x00, 0x01, 0x02} // fake mp3 header bytes
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write(expected)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GenerateAudio(context.Background(), "hi", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(expected) {
		t.Errorf("audio bytes mismatch")
	}
}

func TestTranscribe_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"text":"transcribed"}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Transcribe(context.Background(), strings.NewReader("fake audio"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "transcribed" {
		t.Errorf("got %q, want %q", got, "transcribed")
	}
}
