package commands

import (
	"fmt"
	"strings"
	"testing"
)

func TestLinkCommand_New_BasicURL(t *testing.T) {
	mock := NewMockExecutor().WithLinkJSON([]byte(`{"path":"/vault/link.md","uuid":"abc-123","title":"Example","url":"https://example.com"}`))

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	result, err := ruin.Link.New("https://example.com", LinkNewOpts{})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if result.UUID != "abc-123" {
		t.Errorf("UUID = %q, want %q", result.UUID, "abc-123")
	}
	if result.URL != "https://example.com" {
		t.Errorf("URL = %q, want %q", result.URL, "https://example.com")
	}
	if result.Title != "Example" {
		t.Errorf("Title = %q, want %q", result.Title, "Example")
	}
}

func TestLinkCommand_New_AllOpts(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_, _ = ruin.Link.New("https://example.com", LinkNewOpts{
		Title:   "My Title",
		Tags:    "link,reading",
		Parent:  "parent-uuid",
		NoFetch: true,
		Comment: "saved for later",
	})

	args := cap.lastArgs()
	if args[0] != "link" || args[1] != "new" || args[2] != "https://example.com" {
		t.Fatalf("expected [link new https://example.com ...], got %v", args)
	}
	assertArgsContain(t, args, "--title")
	assertArgsContain(t, args, "My Title")
	assertArgsContain(t, args, "--tags")
	assertArgsContain(t, args, "link,reading")
	assertArgsContain(t, args, "--parent")
	assertArgsContain(t, args, "parent-uuid")
	assertArgsContain(t, args, "--no-fetch")
	assertArgsContain(t, args, "--comment")
	assertArgsContain(t, args, "saved for later")
}

func TestLinkCommand_New_NoOptionalFlags(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_, _ = ruin.Link.New("https://example.com", LinkNewOpts{})

	args := cap.lastArgs()
	if len(args) != 3 {
		t.Errorf("expected [link new <url>], got %v", args)
	}
	assertArgsNotContain(t, args, "--title")
	assertArgsNotContain(t, args, "--tags")
	assertArgsNotContain(t, args, "--parent")
	assertArgsNotContain(t, args, "--no-fetch")
	assertArgsNotContain(t, args, "--comment")
}

func TestLinkCommand_New_Error(t *testing.T) {
	mock := NewMockExecutor().WithError(errMock)

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	_, err := ruin.Link.New("https://example.com", LinkNewOpts{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLinkCommand_Resolve(t *testing.T) {
	mock := NewMockExecutor().WithLinkJSON([]byte(`{"url":"https://example.com","title":"Example Domain","summary":"An example site","resolved_via":"fetch"}`))

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	result, err := ruin.Link.Resolve("https://example.com")
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if result.URL != "https://example.com" {
		t.Errorf("URL = %q, want %q", result.URL, "https://example.com")
	}
	if result.Title != "Example Domain" {
		t.Errorf("Title = %q, want %q", result.Title, "Example Domain")
	}
	if result.Summary != "An example site" {
		t.Errorf("Summary = %q, want %q", result.Summary, "An example site")
	}
	if result.ResolvedVia != "fetch" {
		t.Errorf("ResolvedVia = %q, want %q", result.ResolvedVia, "fetch")
	}
}

func TestLinkCommand_Resolve_Args(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_, _ = ruin.Link.Resolve("https://example.com")

	args := cap.lastArgs()
	if len(args) != 3 || args[0] != "link" || args[1] != "resolve" || args[2] != "https://example.com" {
		t.Errorf("expected [link resolve https://example.com], got %v", args)
	}
}

func TestLinkCommand_Resolve_Error(t *testing.T) {
	mock := NewMockExecutor().WithError(errMock)

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	_, err := ruin.Link.Resolve("https://example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLinkCommand_New_TitleWithSpaces(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_, _ = ruin.Link.New("https://example.com", LinkNewOpts{
		Title: "A Title With Spaces",
	})

	args := cap.lastArgs()
	joined := strings.Join(args, "|")
	if !strings.Contains(joined, "--title|A Title With Spaces") {
		t.Errorf("title should be passed as single arg, got args: %v", args)
	}
}

var errMock = fmt.Errorf("mock error")
