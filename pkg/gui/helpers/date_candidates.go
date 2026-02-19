package helpers

import (
	"strings"
	"time"

	"kvnd/lazyruin/pkg/gui/types"

	anytime "github.com/ijt/go-anytime"
)

// AmbientDateCandidates returns a fallback candidate function that parses bare
// tokens with go-anytime and suggests created:/after:/before: date filters.
func AmbientDateCandidates() func(string) []types.CompletionItem {
	return ambientDateCandidates
}

func ambientDateCandidates(token string) []types.CompletionItem {
	parsed, err := anytime.Parse(token, time.Now())
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

// atDateSuggestions are natural-language date strings that go-anytime can parse.
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

var daysOfNextWeek = []string{
	"next sunday", "next monday", "next tuesday", "next wednesday",
	"next thursday", "next friday", "next saturday",
}

// cliNativeDates are the keywords the ruin CLI handles directly.
var cliNativeDates = map[string]bool{
	"today": true, "yesterday": true, "tomorrow": true,
	"this-week": true, "last-week": true, "next-week": true,
	"this-month": true, "last-month": true, "next-month": true,
}

// AtDateCandidates returns completion items for the @ trigger, resolving
// natural language dates to ISO format.
func AtDateCandidates(filter string) []types.CompletionItem {
	return atDateCandidatesAt(filter, time.Now())
}

func atDateCandidatesAt(filter string, now time.Time) []types.CompletionItem {
	filterLower := strings.ToLower(filter)

	if filterLower == "next week" || filterLower == "last week" {
		prefix := strings.SplitN(filterLower, " ", 2)[0]
		var items []types.CompletionItem
		for _, day := range daysOfNextWeek {
			s := prefix + day[4:]
			items = append(items, atSuggestionItem(s, now))
		}
		return items
	}

	var items []types.CompletionItem
	for _, s := range atDateSuggestions {
		if filterLower != "" && !strings.HasPrefix(s, filterLower) {
			continue
		}
		items = append(items, atSuggestionItem(s, now))
	}

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

func parseUSDate(s string) (time.Time, bool) {
	for _, layout := range []string{"1/2/2006", "1/2/06"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

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
