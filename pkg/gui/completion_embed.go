package gui

import (
	"sort"
	"strings"

	"github.com/donnellyk/lazyruin/pkg/gui/types"
)

// dynamicEmbedTypes lists the four recognized dynamic embed type prefixes.
// Order here determines dropdown order in embedCandidates (type prefixes
// first, then note titles as a fallback for static embeds).
var dynamicEmbedTypes = []struct {
	name   string
	detail string
}{
	{"search", "full-text / tag search"},
	{"pick", "extract lines by tag"},
	{"query", "run saved query"},
	{"compose", "embed composed document"},
}

// embedState describes whether the cursor is inside an unclosed `![[...` on
// the current line, and if so, what phase of the dynamic embed it's in.
type embedState struct {
	inEmbed       bool   // cursor is inside an unclosed `![[...`
	embedType     string // one of dynamicEmbedTypes.name, or "" if no type typed yet
	insideOptions bool   // cursor is past a `|` within the embed (in options phase)
}

// insideDynamicEmbed parses the current line backward from cursor to
// determine the embed state. Returns inEmbed=false if the cursor is not
// inside `![[...` on the current line, or if the embed is already closed
// by `]]` before the cursor. content is the full buffer and cursor is a
// byte offset.
func insideDynamicEmbed(content string, cursor int) embedState {
	if cursor > len(content) {
		cursor = len(content)
	}
	// Find the start of the current line
	lineStart := 0
	for i := cursor - 1; i >= 0; i-- {
		if content[i] == '\n' {
			lineStart = i + 1
			break
		}
	}
	line := content[lineStart:cursor]

	// Find the last `![[` on the line
	openIdx := strings.LastIndex(line, "![[")
	if openIdx < 0 {
		return embedState{}
	}
	// If `]]` appears after the `![[`, the embed is closed
	after := line[openIdx+3:]
	if strings.Contains(after, "]]") {
		return embedState{}
	}

	state := embedState{inEmbed: true}

	// Detect the type prefix `TYPE:`. Only recognize known types.
	if colonIdx := strings.Index(after, ":"); colonIdx > 0 {
		maybeType := after[:colonIdx]
		if isKnownEmbedType(maybeType) {
			state.embedType = maybeType
			// Determine options phase: a `|` anywhere after the type prefix
			// (on the same line, before the cursor) signals we're in options.
			rest := after[colonIdx+1:]
			if strings.Contains(rest, "|") {
				state.insideOptions = true
			}
		}
	}

	return state
}

// isKnownEmbedType returns true if name is one of the recognized dynamic
// embed type prefixes (search, pick, query, compose).
func isKnownEmbedType(name string) bool {
	for _, t := range dynamicEmbedTypes {
		if t.name == name {
			return true
		}
	}
	return false
}

// embedTrigger returns the `![[` completion trigger. Used in the capture
// popup to offer dynamic embed type prefixes alongside static note titles.
func (gui *Gui) embedTrigger() types.CompletionTrigger {
	return types.CompletionTrigger{Prefix: "![[", Candidates: gui.embedCandidates}
}

// embedCandidates returns the candidates offered inside `![[...`. The list
// starts with the four dynamic type prefixes (ContinueCompleting so that
// downstream triggers like `#` or `@` take over after a selection) and is
// followed by note titles for the static embed case.
func (gui *Gui) embedCandidates(filter string) []types.CompletionItem {
	filterLower := strings.ToLower(filter)
	var items []types.CompletionItem

	for _, t := range dynamicEmbedTypes {
		label := t.name + ":"
		if filter != "" && !strings.Contains(label, filterLower) {
			continue
		}
		// InsertText re-includes the `![[` prefix because acceptReplaceCompletion
		// deletes from cursor back to TriggerStart (the `!`), so the user's
		// typed `![[` would otherwise be stripped.
		items = append(items, types.CompletionItem{
			Label:              label,
			InsertText:         "![[" + t.name + ": ",
			Detail:             t.detail,
			ContinueCompleting: true,
		})
	}

	// Static embed fallback: offer note titles. The InsertText wraps in
	// `![[...]]` because the trigger prefix is `![[` and acceptCompletion
	// will backspace the `![[` before typing the InsertText.
	seen := make(map[string]bool)
	for _, note := range gui.contexts.Notes.Items {
		title := note.Title
		if title == "" || seen[title] {
			continue
		}
		if filter != "" && !strings.Contains(strings.ToLower(title), filterLower) {
			continue
		}
		seen[title] = true
		items = append(items, types.CompletionItem{
			Label:      title,
			InsertText: "![[" + title + "]]",
			Detail:     note.ShortDate(),
		})
	}

	return items
}

// embedOptionKeyCandidates returns option-key suggestions for the given
// dynamic embed type. Called when the cursor is past `|` in an unclosed
// embed. Keys end with `=` for value-taking options; bare flags (any, done,
// all) have no trailing `=`.
func embedOptionKeyCandidates(embedType, filter string) []types.CompletionItem {
	var items []types.CompletionItem

	switch embedType {
	case "search", "query":
		items = []types.CompletionItem{
			{Label: "format=", InsertText: "format=", Detail: "output format", ContinueCompleting: true},
			{Label: "sort=", InsertText: "sort=", Detail: "result ordering", ContinueCompleting: true},
			{Label: "limit=", InsertText: "limit=", Detail: "max results", ContinueCompleting: true},
			{Label: "tag-scope=", InsertText: "tag-scope=", Detail: "all|global|inline", ContinueCompleting: true},
			{Label: "depth=", InsertText: "depth=", Detail: "expand children", ContinueCompleting: true},
			{Label: "empty=", InsertText: "empty=", Detail: "hide or show", ContinueCompleting: true},
		}
	case "pick":
		items = []types.CompletionItem{
			{Label: "format=", InsertText: "format=", Detail: "grouped|flat", ContinueCompleting: true},
			{Label: "sort=", InsertText: "sort=", Detail: "source note ordering", ContinueCompleting: true},
			{Label: "limit=", InsertText: "limit=", Detail: "max source notes", ContinueCompleting: true},
			{Label: "filter=", InsertText: "filter=", Detail: "note-level filter", ContinueCompleting: true},
			{Label: "empty=", InsertText: "empty=", Detail: "hide or show", ContinueCompleting: true},
			{Label: "any", InsertText: "any", Detail: "OR mode for positive tags"},
			{Label: "done", InsertText: "done", Detail: "only done items"},
			{Label: "all", InsertText: "all", Detail: "include done and undone"},
		}
	case "compose":
		items = []types.CompletionItem{
			{Label: "depth=", InsertText: "depth=", Detail: "max sub-compose depth", ContinueCompleting: true},
		}
	default:
		return nil
	}

	if filter == "" {
		return items
	}

	filterLower := strings.ToLower(filter)
	var filtered []types.CompletionItem
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Label), filterLower) ||
			strings.Contains(strings.ToLower(item.Detail), filterLower) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// embedOptionValueCandidates returns value suggestions for a given
// (embedType, optionKey) pair.
//
// Return contract:
//   - nil: the key is free-form (limit, depth, filter, or unknown); the
//     caller should fall through to the standard trigger scan.
//   - non-nil (possibly empty): the key has a known enumerated value set.
//     An empty slice means the filter matched nothing; the caller should
//     NOT fall through.
func embedOptionValueCandidates(embedType, key, filter string) []types.CompletionItem {
	var items []types.CompletionItem

	switch key {
	case "format":
		if embedType == "pick" {
			items = []types.CompletionItem{
				{Label: "grouped", InsertText: "grouped", Detail: "lines grouped by note"},
				{Label: "flat", InsertText: "flat", Detail: "flat bullet list"},
			}
		} else {
			items = []types.CompletionItem{
				{Label: "content", InsertText: "content", Detail: "full note content"},
				{Label: "list", InsertText: "list", Detail: "bullet list of wiki links"},
				{Label: "summary", InsertText: "summary", Detail: "title + date + first line"},
			}
		}
	case "sort":
		items = []types.CompletionItem{
			{Label: "created:desc", InsertText: "created:desc", Detail: "newest first"},
			{Label: "created:asc", InsertText: "created:asc", Detail: "oldest first"},
			{Label: "updated:desc", InsertText: "updated:desc", Detail: "recently updated"},
			{Label: "title:asc", InsertText: "title:asc", Detail: "A-Z"},
			{Label: "title:desc", InsertText: "title:desc", Detail: "Z-A"},
			{Label: "order:asc", InsertText: "order:asc", Detail: "manual order"},
		}
	case "tag-scope":
		items = []types.CompletionItem{
			{Label: "all", InsertText: "all", Detail: "global and inline"},
			{Label: "global", InsertText: "global", Detail: "only global tags"},
			{Label: "inline", InsertText: "inline", Detail: "only inline tags"},
		}
	case "empty":
		items = []types.CompletionItem{
			{Label: "hide", InsertText: "hide", Detail: "silently remove empty embed"},
		}
	default:
		return nil
	}

	if filter == "" {
		return items
	}
	filterLower := strings.ToLower(filter)
	// Initialize non-nil so "known key, filter matched nothing" returns an
	// empty slice — the caller uses nil as the free-form signal and would
	// otherwise misinterpret a zero-match filter as "fall through".
	filtered := []types.CompletionItem{}
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Label), filterLower) ||
			strings.Contains(strings.ToLower(item.Detail), filterLower) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// savedQueryCandidates returns saved query names as completion items.
// Used inside `![[query: <partial>`.
func (gui *Gui) savedQueryCandidates(filter string) []types.CompletionItem {
	queries := gui.contexts.Queries.Queries
	if len(queries) == 0 {
		return nil
	}

	names := make([]string, 0, len(queries))
	expansions := make(map[string]string, len(queries))
	for _, q := range queries {
		names = append(names, q.Name)
		expansions[q.Name] = q.Query
	}
	sort.Strings(names)

	filterLower := strings.ToLower(filter)
	var items []types.CompletionItem
	for _, name := range names {
		if filter != "" && !strings.Contains(strings.ToLower(name), filterLower) {
			continue
		}
		detail := expansions[name]
		if len(detail) > 40 {
			detail = detail[:37] + "..."
		}
		items = append(items, types.CompletionItem{
			Label:      name,
			InsertText: name,
			Detail:     detail,
		})
	}
	return items
}

// composeRefCandidates returns candidates for `![[compose: <partial>`:
// parent bookmarks first (most common use), then note titles as a
// secondary section.
func (gui *Gui) composeRefCandidates(filter string) []types.CompletionItem {
	filterLower := strings.ToLower(filter)
	var items []types.CompletionItem

	for _, p := range gui.contexts.Queries.Parents {
		if filter != "" && !strings.Contains(strings.ToLower(p.Name), filterLower) {
			continue
		}
		detail := "bookmark"
		if p.Title != "" {
			detail = "bookmark · " + p.Title
		} else if p.File != "" {
			detail = "bookmark · " + p.File
		}
		items = append(items, types.CompletionItem{
			Label:      p.Name,
			InsertText: p.Name,
			Detail:     detail,
		})
	}

	seen := make(map[string]bool)
	for _, note := range gui.contexts.Notes.Items {
		title := note.Title
		if title == "" || seen[title] {
			continue
		}
		if filter != "" && !strings.Contains(strings.ToLower(title), filterLower) {
			continue
		}
		seen[title] = true
		items = append(items, types.CompletionItem{
			Label:      title,
			InsertText: title,
			Detail:     note.ShortDate(),
		})
	}

	return items
}

// dynamicEmbedCandidates routes to the appropriate candidate function
// based on embed state: option key/value for the options phase, saved
// queries for `query:`, bookmarks/titles for `compose:`. Returns ok=false
// to signal "no dynamic dispatch applies; fall through to normal trigger
// detection" (used for `search:` / `pick:` and the no-type-yet state).
func (gui *Gui) dynamicEmbedCandidates(content string, cursor int, es embedState) (items []types.CompletionItem, start int, ok bool) {
	if !es.inEmbed {
		return nil, 0, false
	}
	if es.insideOptions {
		filter, optKey, optStart := dynamicEmbedFilter(content, cursor, es)
		if optKey != "" {
			items := embedOptionValueCandidates(es.embedType, optKey, filter)
			if items == nil {
				// Free-form value (e.g. filter=, limit=, depth=). Fall through
				// to the standard trigger scan so `#tag`, `@date`, and field
				// prefixes still fire while the user types inside the value.
				return nil, 0, false
			}
			return items, optStart, true
		}
		return embedOptionKeyCandidates(es.embedType, filter), optStart, true
	}
	switch es.embedType {
	case "query":
		filter, _, qStart := dynamicEmbedFilter(content, cursor, es)
		return gui.savedQueryCandidates(filter), qStart, true
	case "compose":
		filter, _, cStart := dynamicEmbedFilter(content, cursor, es)
		return gui.composeRefCandidates(filter), cStart, true
	}
	return nil, 0, false
}

// withoutAbbreviationTrigger returns a copy of triggers with the `!`
// abbreviation trigger removed. Used inside dynamic embeds where `!` means
// negation, not abbreviation.
func withoutAbbreviationTrigger(triggers []types.CompletionTrigger) []types.CompletionTrigger {
	out := make([]types.CompletionTrigger, 0, len(triggers))
	for _, t := range triggers {
		if t.Prefix == "!" {
			continue
		}
		out = append(out, t)
	}
	return out
}

// dynamicEmbedFilter returns the filter text between the type prefix (and
// optional `|` for options) and the cursor, for use by non-prefix
// dispatch in updateCompletion. For `![[query: weekl|y`, returns ("weekly", start).
//
// When state.insideOptions is true, returns the filter relative to the last
// `|` and parses whether the cursor is on a key (`format`) or value
// (`format=con`), reported via `optionKey` (empty for key phase, set for
// value phase).
//
// start is the byte offset in content where the partial text begins.
func dynamicEmbedFilter(content string, cursor int, state embedState) (filter, optionKey string, start int) {
	if !state.inEmbed {
		return "", "", 0
	}
	if cursor > len(content) {
		cursor = len(content)
	}
	// Find the `![[TYPE: ` or `![[` position on the current line
	lineStart := 0
	for i := cursor - 1; i >= 0; i-- {
		if content[i] == '\n' {
			lineStart = i + 1
			break
		}
	}
	line := content[lineStart:cursor]
	openIdx := strings.LastIndex(line, "![[")
	if openIdx < 0 {
		return "", "", 0
	}
	afterOpen := line[openIdx+3:]

	// Skip past the type prefix if present
	tail := afterOpen
	tailOffset := openIdx + 3
	if state.embedType != "" {
		prefix := state.embedType + ":"
		tail = strings.TrimPrefix(afterOpen, prefix)
		tailOffset += len(prefix)
		// Skip a single leading space after the colon (the standard form)
		if strings.HasPrefix(tail, " ") {
			tail = tail[1:]
			tailOffset++
		}
	}

	if state.insideOptions {
		// The filter is everything after the last `|` (and optional spaces
		// after it). Within that, if we're past a `=`, the key is the text
		// before `=` and the filter is everything after.
		pipeIdx := strings.LastIndex(tail, "|")
		if pipeIdx < 0 {
			return "", "", 0
		}
		optRegion := tail[pipeIdx+1:]
		optOffset := tailOffset + pipeIdx + 1
		// Skip leading spaces
		for len(optRegion) > 0 && (optRegion[0] == ' ' || optRegion[0] == '\t') {
			optRegion = optRegion[1:]
			optOffset++
		}
		// The option region may contain earlier `,`-separated options; only
		// the last segment is what the user is currently typing.
		if commaIdx := strings.LastIndex(optRegion, ","); commaIdx >= 0 {
			optRegion = optRegion[commaIdx+1:]
			optOffset += commaIdx + 1
			for len(optRegion) > 0 && (optRegion[0] == ' ' || optRegion[0] == '\t') {
				optRegion = optRegion[1:]
				optOffset++
			}
		}
		if eqIdx := strings.Index(optRegion, "="); eqIdx >= 0 {
			return optRegion[eqIdx+1:], optRegion[:eqIdx], lineStart + optOffset + eqIdx + 1
		}
		return optRegion, "", lineStart + optOffset
	}

	// Non-options phase: filter is the entire tail
	return tail, "", lineStart + tailOffset
}
