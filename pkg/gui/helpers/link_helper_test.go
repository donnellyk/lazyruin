package helpers

import "testing"

func TestParseURLAndTags(t *testing.T) {
	h := &LinkHelper{}

	tests := []struct {
		name     string
		input    string
		wantURL  string
		wantTags []string
	}{
		{
			name:     "url only",
			input:    "https://example.com",
			wantURL:  "https://example.com",
			wantTags: nil,
		},
		{
			name:     "url with one tag",
			input:    "https://example.com #reading",
			wantURL:  "https://example.com",
			wantTags: []string{"reading"},
		},
		{
			name:     "url with multiple tags",
			input:    "https://example.com #reading #tech #ai",
			wantURL:  "https://example.com",
			wantTags: []string{"reading", "tech", "ai"},
		},
		{
			name:     "empty input",
			input:    "",
			wantURL:  "",
			wantTags: nil,
		},
		{
			name:     "only tags no url",
			input:    "#reading #tech",
			wantURL:  "",
			wantTags: []string{"reading", "tech"},
		},
		{
			name:     "bare hash ignored",
			input:    "https://example.com #",
			wantURL:  "https://example.com",
			wantTags: nil,
		},
		{
			name:     "tags before url",
			input:    "#reading https://example.com",
			wantURL:  "https://example.com",
			wantTags: []string{"reading"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, gotTags := h.parseURLAndTags(tt.input)
			if gotURL != tt.wantURL {
				t.Errorf("url = %q, want %q", gotURL, tt.wantURL)
			}
			if len(gotTags) != len(tt.wantTags) {
				t.Fatalf("tags = %v, want %v", gotTags, tt.wantTags)
			}
			for i := range gotTags {
				if gotTags[i] != tt.wantTags[i] {
					t.Errorf("tags[%d] = %q, want %q", i, gotTags[i], tt.wantTags[i])
				}
			}
		})
	}
}

func TestParseLinkContent(t *testing.T) {
	h := &LinkHelper{}

	tests := []struct {
		name        string
		content     string
		url         string
		wantTitle   string
		wantComment string
		wantTags    []string
	}{
		{
			name:        "full resolved content",
			content:     "# Example Page\n\nhttps://example.com\n\nThis is the summary.",
			url:         "https://example.com",
			wantTitle:   "Example Page",
			wantComment: "This is the summary.",
		},
		{
			name:        "no title",
			content:     "https://example.com\n\nSome comment.",
			url:         "https://example.com",
			wantTitle:   "",
			wantComment: "Some comment.",
		},
		{
			name:        "tags extracted from content",
			content:     "# Example Page\n\nhttps://example.com\n\nSummary here.\n\n#reading #tech",
			url:         "https://example.com",
			wantTitle:   "Example Page",
			wantComment: "Summary here.",
			wantTags:    []string{"reading", "tech"},
		},
		{
			name:        "url only",
			content:     "https://example.com",
			url:         "https://example.com",
			wantTitle:   "",
			wantComment: "",
		},
		{
			name:     "tags added after resolve",
			content:  "# My Title\n\nhttps://example.com\n\n#new-tag #another",
			url:      "https://example.com",
			wantTitle: "My Title",
			wantTags:  []string{"new-tag", "another"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTitle, gotComment, gotTags := h.parseLinkContent(tt.content, tt.url)
			if gotTitle != tt.wantTitle {
				t.Errorf("title = %q, want %q", gotTitle, tt.wantTitle)
			}
			if gotComment != tt.wantComment {
				t.Errorf("comment = %q, want %q", gotComment, tt.wantComment)
			}
			if len(gotTags) != len(tt.wantTags) {
				t.Fatalf("tags = %v, want %v", gotTags, tt.wantTags)
			}
			for i := range gotTags {
				if gotTags[i] != tt.wantTags[i] {
					t.Errorf("tags[%d] = %q, want %q", i, gotTags[i], tt.wantTags[i])
				}
			}
		})
	}
}
