package commands

import (
	"errors"
	"testing"
)

func TestEmbedEval_SearchEnvelope(t *testing.T) {
	mock := NewMockExecutor().WithEmbedJSON([]byte(`{
		"type": "search",
		"query": "#daily",
		"options": {"limit": "5"},
		"results": [
			{"uuid": "u1", "title": "Daily 1", "path": "/a.md", "tags": ["daily"]},
			{"uuid": "u2", "title": "Daily 2", "path": "/b.md", "tags": ["daily"]}
		]
	}`))
	r := NewRuinCommandWithExecutor(mock, "/v")

	res, err := r.Embed.Eval("![[search: #daily | limit=5]]")
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}

	if res.Type != EmbedTypeSearch {
		t.Errorf("Type = %q, want search", res.Type)
	}
	if res.Query != "#daily" {
		t.Errorf("Query = %q, want #daily", res.Query)
	}
	if len(res.Notes) != 2 {
		t.Fatalf("len(Notes) = %d, want 2", len(res.Notes))
	}
	if res.Notes[0].UUID != "u1" || res.Notes[1].UUID != "u2" {
		t.Errorf("UUIDs = %q, %q", res.Notes[0].UUID, res.Notes[1].UUID)
	}
	if len(mock.Calls) != 1 {
		t.Fatalf("expected 1 Execute call, got %d", len(mock.Calls))
	}
	args := mock.Calls[0]
	if len(args) < 3 || args[0] != "embed" || args[1] != "eval" {
		t.Errorf("unexpected args: %v", args)
	}
	if args[2] != "![[search: #daily | limit=5]]" {
		t.Errorf("expected verbatim embed string at args[2], got %q", args[2])
	}
}

func TestEmbedEval_QueryEnvelopeReusesNotes(t *testing.T) {
	mock := NewMockExecutor().WithEmbedJSON([]byte(`{
		"type": "query",
		"query": "saved-name",
		"results": [{"uuid": "u3", "title": "Q result", "path": "/q.md"}]
	}`))
	r := NewRuinCommandWithExecutor(mock, "/v")

	res, err := r.Embed.Eval("![[query: saved-name]]")
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if res.Type != EmbedTypeQuery {
		t.Errorf("Type = %q, want query", res.Type)
	}
	if len(res.Notes) != 1 || res.Notes[0].UUID != "u3" {
		t.Fatalf("Notes = %+v", res.Notes)
	}
}

func TestEmbedEval_PickEnvelope(t *testing.T) {
	mock := NewMockExecutor().WithEmbedJSON([]byte(`{
		"type": "pick",
		"query": "#followup",
		"results": [
			{"uuid": "u1", "title": "n1", "file": "/a.md", "matches": [{"line": 3, "content": "do thing #followup", "tags": ["followup"], "done": false}]}
		]
	}`))
	r := NewRuinCommandWithExecutor(mock, "/v")

	res, err := r.Embed.Eval("![[pick: #followup]]")
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if res.Type != EmbedTypePick {
		t.Errorf("Type = %q, want pick", res.Type)
	}
	if len(res.Picks) != 1 || res.Picks[0].UUID != "u1" {
		t.Fatalf("Picks = %+v", res.Picks)
	}
	if len(res.Picks[0].Matches) != 1 || res.Picks[0].Matches[0].Line != 3 {
		t.Errorf("expected one match on line 3, got %+v", res.Picks[0].Matches)
	}
}

func TestEmbedEval_ComposeEnvelope(t *testing.T) {
	mock := NewMockExecutor().WithEmbedJSON([]byte(`{
		"type": "compose",
		"query": "Some Note",
		"results": {
			"expanded_markdown": "# Some Note\n\nbody\n",
			"source_map": [{"uuid": "u1", "path": "/a.md", "title": "Some Note", "start_line": 1, "end_line": 3}]
		}
	}`))
	r := NewRuinCommandWithExecutor(mock, "/v")

	res, err := r.Embed.Eval("![[compose: Some Note]]")
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if res.Type != EmbedTypeCompose {
		t.Errorf("Type = %q, want compose", res.Type)
	}
	if res.Compose == nil {
		t.Fatal("Compose nil")
	}
	if res.Compose.ExpandedMarkdown == "" {
		t.Error("ExpandedMarkdown empty")
	}
	if len(res.Compose.SourceMap) != 1 || res.Compose.SourceMap[0].UUID != "u1" {
		t.Errorf("SourceMap = %+v", res.Compose.SourceMap)
	}
}

func TestEmbedEval_UnknownTypeErrors(t *testing.T) {
	mock := NewMockExecutor().WithEmbedJSON([]byte(`{"type":"unknown","results":null}`))
	r := NewRuinCommandWithExecutor(mock, "/v")

	if _, err := r.Embed.Eval("![[bogus: foo]]"); err == nil {
		t.Fatal("expected error for unknown embed type")
	}
}

func TestEmbedEval_PropagatesExecutorError(t *testing.T) {
	wantErr := errors.New("boom")
	mock := NewMockExecutor().WithError(wantErr)
	r := NewRuinCommandWithExecutor(mock, "/v")

	if _, err := r.Embed.Eval("![[search: #x]]"); !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}

func TestEmbedEval_NullResultsAreZeroValue(t *testing.T) {
	cases := []struct {
		name     string
		envelope string
		check    func(*testing.T, *EmbedResult)
	}{
		{
			"search null",
			`{"type":"search","query":"x","results":null}`,
			func(t *testing.T, r *EmbedResult) {
				if len(r.Notes) != 0 {
					t.Errorf("expected empty Notes, got %+v", r.Notes)
				}
			},
		},
		{
			"pick missing",
			`{"type":"pick","query":"x"}`,
			func(t *testing.T, r *EmbedResult) {
				if len(r.Picks) != 0 {
					t.Errorf("expected empty Picks, got %+v", r.Picks)
				}
			},
		},
		{
			"compose null",
			`{"type":"compose","query":"x","results":null}`,
			func(t *testing.T, r *EmbedResult) {
				if r.Compose != nil {
					t.Errorf("expected nil Compose, got %+v", r.Compose)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := NewMockExecutor().WithEmbedJSON([]byte(tc.envelope))
			r := NewRuinCommandWithExecutor(mock, "/v")
			res, err := r.Embed.Eval("![[" + tc.name + "]]")
			if err != nil {
				t.Fatalf("Eval: %v", err)
			}
			tc.check(t, res)
		})
	}
}
