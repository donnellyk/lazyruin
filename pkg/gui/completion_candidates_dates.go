package gui

import (
	"strings"
	"time"

	"github.com/donnellyk/lazyruin/pkg/gui/types"

	anytime "github.com/ijt/go-anytime"
)

// dateLiterals are the date keywords supported natively by ruin's date filters.
var dateLiterals = []struct {
	value  string
	detail string
}{
	{"today", "today"},
	{"yesterday", "yesterday"},
	{"tomorrow", "tomorrow"},
	{"this-week", "current week"},
	{"last-week", "previous week"},
	{"next-week", "next week"},
	{"this-month", "current month"},
	{"last-month", "previous month"},
	{"next-month", "next month"},
}

// dateCandidates builds completion items for a date-prefix filter (e.g. "created:", "updated:").
// It offers literal keywords (today/yesterday/tomorrow) plus dynamic natural-language parsing
// via go-anytime, showing resolved ISO dates with a human-readable detail.
func dateCandidates(prefix, filter string) []types.CompletionItem {
	return dateCandidatesAt(prefix, filter, time.Now())
}

// dateCandidatesAt is the testable core of dateCandidates, accepting an explicit "now" time.
func dateCandidatesAt(prefix, filter string, now time.Time) []types.CompletionItem {
	filterLower := strings.ToLower(filter)
	var items []types.CompletionItem

	// Always offer literal keywords that match the filter.
	for _, s := range dateLiterals {
		if filterLower != "" && !strings.HasPrefix(s.value, filterLower) {
			continue
		}
		items = append(items, types.CompletionItem{
			Label:      prefix + s.value,
			InsertText: prefix + s.value,
			Detail:     s.detail,
		})
	}

	// If the filter is non-empty and not an exact literal match, try anytime parsing.
	if filter != "" && !isDateLiteral(filterLower) {
		if parsed, err := anytime.Parse(filter, now); err == nil {
			iso := parsed.Format("2006-01-02")
			detail := parsed.Format("Mon, Jan 02, 2006")
			items = append(items, types.CompletionItem{
				Label:      prefix + iso,
				InsertText: prefix + iso,
				Detail:     detail,
			})
		}
	}

	return items
}

// ambientDateCandidates parses a bare token (no trigger prefix) with go-anytime
// and returns a created: suggestion if it resolves to a valid date.
// Used as a FallbackCandidates function for the search popup.
func ambientDateCandidates(token string) []types.CompletionItem {
	return ambientDateCandidatesAt(token, time.Now())
}

// ambientDateCandidatesAt is the testable core of ambientDateCandidates.
func ambientDateCandidatesAt(token string, now time.Time) []types.CompletionItem {
	parsed, err := anytime.Parse(token, now)
	if err != nil {
		return nil
	}
	iso := parsed.Format("2006-01-02")
	detail := parsed.Format("Mon, Jan 02, 2006")
	return []types.CompletionItem{
		{Label: "created:" + iso, InsertText: "created:" + iso, Detail: detail},
		{Label: "after:" + iso, InsertText: "after:" + iso, Detail: detail},
		{Label: "before:" + iso, InsertText: "before:" + iso, Detail: detail},
	}
}

// atDateCandidates returns completion items for the @ trigger, resolving natural
// language dates to ISO format. Used in both capture and search popups.
func atDateCandidates(filter string) []types.CompletionItem {
	return atDateCandidatesAt(filter, time.Now())
}

// atDateSuggestions are natural-language date strings that go-anytime can parse.
// Grouped by prefix for compact filtering.
var atDateSuggestions = []string{
	"today", "yesterday", "tomorrow",
	"this-week", "last-week", "next-week",
	"this-month", "last-month", "next-month",
	"monday", "tuesday", "wednesday", "thursday",
	"friday", "saturday", "sunday",
	"next monday", "next tuesday", "next wednesday", "next thursday",
	"next friday", "next saturday", "next sunday",
	"last monday", "last tuesday", "last wednesday", "last thursday",
	"last friday", "last saturday", "last sunday",
	"next week", "last week",
	"next month", "last month",
	"january", "february", "march", "april", "may", "june",
	"july", "august", "september", "october", "november", "december",
	"next year", "last year",
}

// daysOfNextWeek returns "next sunday" through "next saturday" for expanding "next week".
var daysOfNextWeek = []string{
	"next sunday", "next monday", "next tuesday", "next wednesday",
	"next thursday", "next friday", "next saturday",
}

// atDateCandidatesAt is the testable core of atDateCandidates.
func atDateCandidatesAt(filter string, now time.Time) []types.CompletionItem {
	filterLower := strings.ToLower(filter)

	// "next week" / "last week" → expand to individual days
	if filterLower == "next week" || filterLower == "last week" {
		prefix := strings.SplitN(filterLower, " ", 2)[0] // "next" or "last"
		var items []types.CompletionItem
		for _, day := range daysOfNextWeek {
			s := prefix + day[4:] // "next" + " sunday", etc.
			items = append(items, atSuggestionItem(s, now))
		}
		return items
	}

	// Filter suggestions by prefix match
	var items []types.CompletionItem
	for _, s := range atDateSuggestions {
		if filterLower != "" && !strings.HasPrefix(s, filterLower) {
			continue
		}
		items = append(items, atSuggestionItem(s, now))
	}

	// If nothing matched from suggestions, try US date formats then anytime
	if len(items) == 0 && filter != "" {
		if parsed, ok := parseUSDate(filter); ok {
			iso := parsed.Format("2006-01-02")
			items = append(items, types.CompletionItem{
				Label:      "@" + iso,
				InsertText: "@" + iso,
				Detail:     parsed.Format("Mon, Jan 02"),
			})
		} else if parsed, err := anytime.Parse(filter, now); err == nil {
			iso := parsed.Format("2006-01-02")
			items = append(items, types.CompletionItem{
				Label:      "@" + iso,
				InsertText: "@" + iso,
				Detail:     parsed.Format("Mon, Jan 02"),
			})
		}
	}

	return items
}

// parseUSDate tries MM/DD/YYYY and MM/DD/YY formats, returning the parsed time.
func parseUSDate(s string) (time.Time, bool) {
	for _, layout := range []string{"1/2/2006", "1/2/06"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// cliNativeDates are the keywords the ruin CLI handles directly; all other
// suggestions must be resolved to @YYYY-MM-DD before insertion.
var cliNativeDates = map[string]bool{
	"today": true, "yesterday": true, "tomorrow": true,
	"this-week": true, "last-week": true, "next-week": true,
	"this-month": true, "last-month": true, "next-month": true,
}

// atSuggestionItem builds a types.CompletionItem for a natural-language date string,
// resolving it via go-anytime for the detail preview. CLI-native keywords
// (today/yesterday/tomorrow) insert as-is; everything else inserts as @ISO date.
func atSuggestionItem(s string, now time.Time) types.CompletionItem {
	parsed, err := anytime.Parse(s, now)
	if err != nil {
		return types.CompletionItem{Label: "@" + s, InsertText: "@" + s}
	}
	insert := "@" + s
	if !cliNativeDates[s] {
		insert = "@" + parsed.Format("2006-01-02")
	}
	return types.CompletionItem{
		Label:      "@" + s,
		InsertText: insert,
		Detail:     parsed.Format("Mon, Jan 02"),
	}
}

// isDateLiteral returns true if s exactly matches a supported literal keyword.
func isDateLiteral(s string) bool {
	for _, lit := range dateLiterals {
		if s == lit.value {
			return true
		}
	}
	return false
}

// betweenCandidates returns between: filter suggestions with computed ISO date ranges.
func (gui *Gui) betweenCandidates(filter string) []types.CompletionItem {
	return betweenCandidatesAt(filter, time.Now())
}

// betweenCandidatesAt is the testable core of betweenCandidates.
func betweenCandidatesAt(filter string, now time.Time) []types.CompletionItem {
	today := now.Format("2006-01-02")
	weekAgo := now.AddDate(0, 0, -7).Format("2006-01-02")
	monthAgo := now.AddDate(0, -1, 0).Format("2006-01-02")
	yearAgo := now.AddDate(-1, 0, 0).Format("2006-01-02")

	shortcuts := []types.CompletionItem{
		{Label: "between:" + weekAgo + "," + today, InsertText: "between:" + weekAgo + "," + today, Detail: "last 7 days"},
		{Label: "between:" + monthAgo + "," + today, InsertText: "between:" + monthAgo + "," + today, Detail: "last 30 days"},
		{Label: "between:" + yearAgo + "," + today, InsertText: "between:" + yearAgo + "," + today, Detail: "last year"},
	}

	if filter == "" {
		return shortcuts
	}

	filterLower := strings.ToLower(filter)
	var items []types.CompletionItem
	for _, s := range shortcuts {
		suffix := strings.TrimPrefix(s.InsertText, "between:")
		if strings.Contains(suffix, filterLower) || strings.Contains(s.Detail, filterLower) {
			items = append(items, s)
		}
	}

	// Try anytime parsing: offer between:<parsed>,<today>
	if len(items) == 0 {
		if parsed, err := anytime.Parse(filter, now); err == nil {
			iso := parsed.Format("2006-01-02")
			detail := parsed.Format("Mon, Jan 02, 2006") + " to today"
			items = append(items, types.CompletionItem{
				Label:      "between:" + iso + "," + today,
				InsertText: "between:" + iso + "," + today,
				Detail:     detail,
			})
		}
	}

	return items
}
