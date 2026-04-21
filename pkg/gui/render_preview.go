package gui

import (
	"fmt"
	"os"
	"strings"

	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
	"github.com/muesli/reflow/wordwrap"
)

// dimLine wraps a rendered line in AnsiDim. Re-applies dim after every
// ANSI reset so chroma's mid-line resets don't cancel the effect.
func dimLine(text string) string {
	patched := strings.ReplaceAll(text, AnsiReset, AnsiReset+AnsiDim)
	return AnsiDim + patched + AnsiReset
}

// lineIdentity pairs a source note's UUID, content line number, and file path.
// Used internally by BuildCardContent to tag each rendered line with its origin.
type lineIdentity struct {
	uuid    string
	lineNum int
	path    string
}

func (gui *Gui) RenderPreview() {
	v := gui.views.Preview
	if v == nil {
		return
	}

	v.Clear()

	ctx := gui.contexts.ActivePreview()
	ns := ctx.NavState()

	// Snapshot and clear link highlight — it only survives a single render
	// cycle. highlightNextLink/highlightPrevLink set it right before calling
	// renderPreview, so it's visible for this render but auto-clears for any
	// subsequent render triggered by other navigation.
	ns.RenderedLink = ns.HighlightedLink
	ns.HighlightedLink = -1

	switch gui.contexts.ActivePreviewKey {
	case "pickResults":
		pr := gui.contexts.PickResults
		gui.renderPickResults(v, pr.Results, ns, pr.SelectedCardIdx, gui.isPreviewActive())
	case "compose":
		gui.renderSeparatorCards(v, []models.Note{gui.contexts.Compose.Note}, ns, nil)
	case "datePreview":
		dp := gui.contexts.DatePreview
		gui.renderDatePreview(v, dp, ns, gui.isPreviewActive())
	default:
		cl := gui.contexts.CardList
		cards := cl.Cards
		if cl.DisplayState().ShowCompose && len(cl.ComposedCards) == len(cl.Cards) {
			cards = make([]models.Note, len(cl.Cards))
			for i, raw := range cl.Cards {
				if cl.ComposedCards[i] != nil {
					cards[i] = *cl.ComposedCards[i]
				} else {
					cards[i] = raw
				}
			}
		}
		gui.renderSeparatorCards(v, cards, ns, cl.TemporarilyMoved)
	}
}

// stripAnsi removes ANSI escape sequences from a string.
func stripAnsi(s string) string {
	var sb strings.Builder
	inEsc := false
	for _, r := range s {
		if inEsc {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
			}
			continue
		}
		if r == '\x1b' {
			inEsc = true
			continue
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

// visibleWidth returns the number of visible runes in a string, ignoring ANSI escape sequences.
func visibleWidth(s string) int {
	return len([]rune(stripAnsi(s)))
}

// splitLeadingChar splits a string into its first visible character (plus any
// leading ANSI escapes) and the remainder.  If the string is empty or has no
// visible characters, it returns ("", line).
func splitLeadingChar(line string) (string, string) {
	runes := []rune(line)
	i := 0
	for i < len(runes) {
		if runes[i] == '\x1b' {
			// skip ANSI escape sequence
			i++
			for i < len(runes) {
				r := runes[i]
				i++
				if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
					break
				}
			}
			continue
		}
		// First visible character — split after it
		return string(runes[:i+1]), string(runes[i+1:])
	}
	return "", line
}

// isHeaderLine checks whether a rendered line is a markdown ATX header.
// Requires "# " (hash + space) to distinguish from #tag lines.
func isHeaderLine(line string) bool {
	trimmed := strings.TrimLeft(stripAnsi(line), " ")
	// Strip leading '#' characters, then check for a space (ATX heading spec).
	rest := strings.TrimLeft(trimmed, "#")
	return len(rest) < len(trimmed) && len(rest) > 0 && rest[0] == ' '
}

// fprintPreviewLine writes a line to the preview view, applying a dim background
// highlight across the full view width when lineNum matches the current CursorLine.
// When a link is highlighted (HighlightedLink >= 0), only the link span is highlighted
// instead of the full line.
func (gui *Gui) fprintPreviewLine(v *gocui.View, line string, lineNum int, highlight bool, ns *context.PreviewNavState) {
	if !highlight || lineNum != ns.CursorLine {
		fmt.Fprintln(v, line)
		return
	}

	// Check for link-only highlight (set by renderPreview snapshot)
	hl := ns.RenderedLink
	if hl >= 0 && hl < len(ns.Links) {
		link := ns.Links[hl]
		if link.Line == lineNum {
			fmt.Fprintln(v, highlightSpan(line, link.Col, link.Len))
			return
		}
	}

	// Full-line highlight — inset the background by 1 on each side so it
	// does not overlap the view frame border.  Peel the first visible
	// character off the line so it renders before the background starts,
	// keeping the text at its original position.
	width, _ := v.InnerSize()
	leading, hlLine := splitLeadingChar(line)
	pad := max(
		// -1 for trailing inset
		width-visibleWidth(line)-1, 0)
	// Re-apply background after every ANSI reset so chroma formatting
	// doesn't clear our highlight mid-line.
	patched := strings.ReplaceAll(hlLine, AnsiReset, AnsiReset+AnsiDimBg)
	// Use AnsiBgReset (not AnsiReset) so we only clear the background we
	// added.  A full reset would wipe foreground colors that chroma leaves
	// active across line boundaries, causing subsequent lines to lose color.
	fmt.Fprintf(v, "%s%s%s%s%s\n", leading, AnsiDimBg, patched, strings.Repeat(" ", pad), AnsiBgReset)
}

// highlightSpan applies AnsiDimBg to a span of visible characters in an ANSI-decorated
// string. col and length are in visible-character units (ignoring ANSI escapes).
func highlightSpan(line string, col, length int) string {
	var sb strings.Builder
	runes := []rune(line)
	visPos := 0
	inEsc := false
	spanStart := col
	spanEnd := col + length

	for i := range runes {
		r := runes[i]
		if inEsc {
			sb.WriteRune(r)
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
			}
			continue
		}
		if r == '\x1b' {
			sb.WriteRune(r)
			inEsc = true
			continue
		}

		// Visible character
		if visPos == spanStart {
			sb.WriteString(AnsiDimBg)
		}
		sb.WriteRune(r)
		visPos++
		if visPos == spanEnd {
			sb.WriteString(AnsiReset)
		}
	}
	// Safety: close highlight if line ended before spanEnd
	if visPos < spanEnd && visPos >= spanStart {
		sb.WriteString(AnsiReset)
	}
	return sb.String()
}

// collapseConsecutiveBlanks returns lines with runs of visually-blank
// entries (ANSI stripped) at index ≥ start collapsed to a single blank.
// Used only when HideDone strips source lines; without this the blank
// lines that surrounded a hidden line stack up and look like dead space.
func collapseConsecutiveBlanks(lines []types.SourceLine, start int) []types.SourceLine {
	if start >= len(lines) {
		return lines
	}
	out := make([]types.SourceLine, 0, len(lines))
	out = append(out, lines[:start]...)
	prevBlank := false
	for _, l := range lines[start:] {
		blank := strings.TrimSpace(stripAnsi(l.Text)) == ""
		if blank && prevBlank {
			continue
		}
		out = append(out, l)
		prevBlank = blank
	}
	return out
}

// bodyHasNoVisibleContent reports whether every line at or after start is
// visually blank once ANSI escapes are stripped.
func bodyHasNoVisibleContent(lines []types.SourceLine, start int) bool {
	for i := start; i < len(lines); i++ {
		if strings.TrimSpace(stripAnsi(lines[i].Text)) != "" {
			return false
		}
	}
	return true
}

// firstNonBlankLine returns the index of the first non-whitespace line in
// contentLines, or (-1, false) if none exists.
func firstNonBlankLine(contentLines []string) (int, bool) {
	for i, l := range contentLines {
		if strings.TrimSpace(l) != "" {
			return i, true
		}
	}
	return -1, false
}

// buildFallbackLine produces a dimmed SourceLine used by hide-safety (D7)
// to keep a card visible when all body content would otherwise hide.
func buildFallbackLine(line string, id lineIdentity, contentWidth int) types.SourceLine {
	wrapped := wordwrap.String(line, contentWidth)
	first := wrapped
	if nl := strings.IndexByte(wrapped, '\n'); nl >= 0 {
		first = wrapped[:nl]
	}
	return types.SourceLine{
		Text:    dimLine(" " + first),
		UUID:    id.uuid,
		LineNum: id.lineNum,
		Path:    id.path,
	}
}

// cardListComposeMapForNote returns the per-card compose source map for the
// card whose UUID matches note.UUID, or nil when compose output isn't cached
// for it. Match by UUID (not by card index) so this works even when the
// caller doesn't thread the index through render.
func cardListComposeMapForNote(cl *context.CardListContext, note models.Note) []models.SourceMapEntry {
	if len(cl.ComposedCards) != len(cl.Cards) {
		return nil
	}
	for i, c := range cl.Cards {
		if c.UUID == note.UUID && cl.ComposedCards[i] != nil {
			return cl.ComposedSourceMaps[i]
		}
	}
	return nil
}

// BuildCardContent returns the rendered lines for a single card's body content.
// Each SourceLine carries both the rendered text and the source identity
// (UUID, 1-indexed content line number, file path), so callers never need
// to re-derive which source line a visual line came from.
func (gui *Gui) BuildCardContent(note models.Note, contentWidth int) []types.SourceLine {
	content := note.Content
	if content == "" {
		content, _ = gui.loadNoteContent(note.Path)
	}

	ds := gui.contexts.ActivePreview().DisplayState()
	var lines []types.SourceLine

	// Frontmatter display lines — non-content (LineNum=0)
	if ds.ShowFrontmatter {
		if fm, err := gui.loadNoteFrontmatter(note.Path); err == nil && fm != "" {
			lines = append(lines, types.SourceLine{Text: " " + AnsiDim + "---" + AnsiReset})
			for fl := range strings.SplitSeq(fm, "\n") {
				lines = append(lines, types.SourceLine{Text: " " + AnsiDim + fl + AnsiReset})
			}
			lines = append(lines, types.SourceLine{Text: " " + AnsiDim + "---" + AnsiReset})
		}
	}

	// Build a per-content-line identity mapping. In compose mode, each line
	// may belong to a different child note; otherwise lines map to the note itself.
	contentLines := strings.Split(content, "\n")
	identities := make([]lineIdentity, len(contentLines))

	cl := gui.contexts.CardList
	cardComposeMap := cardListComposeMapForNote(cl, note)
	switch {
	case gui.contexts.ActivePreviewKey == "compose" && len(gui.contexts.Compose.SourceMap) > 0:
		gui.buildComposeLineMap(contentLines, gui.contexts.Compose.SourceMap, identities)
	case gui.contexts.ActivePreviewKey != "compose" && cl.DisplayState().ShowCompose && len(cardComposeMap) > 0:
		gui.buildComposeLineMap(contentLines, cardComposeMap, identities)
	default:
		rawLineMap := gui.buildRawLineMap(note)
		for i := range contentLines {
			identities[i] = lineIdentity{uuid: note.UUID, lineNum: rawLineMap[i], path: note.Path}
		}
	}

	// Pre-compute section-level done-ness once per card so every visual
	// line can cheaply answer "am I dim/hidden." D6/D7 apply here.
	doneSection := computeDoneSections(contentLines)
	srcLineDone := func(srcIdx int) bool {
		if srcIdx < 0 || srcIdx >= len(contentLines) {
			return false
		}
		line := contentLines[srcIdx]
		return doneSection[srcIdx] ||
			models.HasDoneTag(line) ||
			models.IsCheckedTodo(line)
	}

	bodyStart := len(lines)

	// Render content lines and tag each with its source identity.
	// Both paths process line-by-line so each visual line can be tagged
	// with the source content line it came from.
	if ds.RenderMarkdown {
		// Highlight the full content (chroma needs context for multi-line constructs),
		// then split by source lines and wrap each individually.
		highlighted := gui.highlightMarkdown(content)
		highlightedLines := strings.Split(highlighted, "\n")
		for srcIdx, hl := range highlightedLines {
			done := srcLineDone(srcIdx)
			if ds.HideDone && done {
				continue
			}
			id := identities[srcIdx]
			isDone := ds.DimDone && done
			wrapped := wordwrap.String(hl, contentWidth)
			for wl := range strings.SplitSeq(wrapped, "\n") {
				text := " " + wl
				if isDone {
					text = dimLine(text)
				}
				lines = append(lines, types.SourceLine{
					Text:    text,
					UUID:    id.uuid,
					LineNum: id.lineNum,
					Path:    id.path,
				})
			}
		}
	} else {
		for srcIdx, l := range contentLines {
			done := srcLineDone(srcIdx)
			if ds.HideDone && done {
				continue
			}
			id := identities[srcIdx]
			isDone := ds.DimDone && done
			wrapped := wordwrap.String(l, contentWidth)
			for wl := range strings.SplitSeq(wrapped, "\n") {
				text := " " + wl
				if isDone {
					text = dimLine(text)
				}
				lines = append(lines, types.SourceLine{
					Text:    text,
					UUID:    id.uuid,
					LineNum: id.lineNum,
					Path:    id.path,
				})
			}
		}
	}

	// When HideDone removes a line with blank lines on either side, the
	// body is left with runs of consecutive blanks. Collapse them so the
	// rendered card stays tight.
	if ds.HideDone {
		lines = collapseConsecutiveBlanks(lines, bodyStart)
	}

	// D7 hide safety: if HideDone collapsed the body to nothing, keep the
	// first non-blank source line (dimmed) so the card is still locatable.
	if ds.HideDone && bodyHasNoVisibleContent(lines, bodyStart) {
		if srcIdx, ok := firstNonBlankLine(contentLines); ok {
			lines = append(lines, buildFallbackLine(contentLines[srcIdx], identities[srcIdx], contentWidth))
		}
	}

	// Trim visually empty lines from start and end (strip ANSI before checking,
	// since rendered markdown lines contain escape codes even when visually blank)
	for len(lines) > 0 && strings.TrimSpace(stripAnsi(lines[0].Text)) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(stripAnsi(lines[len(lines)-1].Text)) == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}

// buildRawLineMap returns a mapping from content line index (0-indexed in
// note.Content) to 1-indexed raw content line number (after frontmatter in
// the source file). This handles the gap created by --strip-title and
// --strip-global-tags.
func (gui *Gui) buildRawLineMap(note models.Note) map[int]int {
	result := make(map[int]int)
	content := note.Content
	if content == "" {
		content, _ = gui.loadNoteContent(note.Path)
	}

	contentLines := strings.Split(content, "\n")

	// Try to read the raw file to determine true content line numbers
	data, err := os.ReadFile(note.Path)
	if err != nil {
		// Can't read file; use sequential numbering as fallback
		for i := range contentLines {
			result[i] = i + 1
		}
		return result
	}

	fileLines := strings.Split(string(data), "\n")
	contentStart := skipFrontmatter(fileLines)

	// Forward scan: for each stripped content line, find its position in the raw file
	rawIdx := contentStart
	for srcIdx, srcLine := range contentLines {
		trimmed := strings.TrimSpace(srcLine)
		for rawIdx < len(fileLines) {
			if strings.TrimSpace(fileLines[rawIdx]) == trimmed {
				result[srcIdx] = rawIdx - contentStart + 1 // 1-indexed
				rawIdx++
				break
			}
			rawIdx++
		}
		if _, ok := result[srcIdx]; !ok {
			// Fallback: sequential from last known position
			result[srcIdx] = srcIdx + 1
		}
	}

	return result
}

// buildComposeLineMap populates identities for compose mode. Each composed
// content line is mapped to the child note that owns it (via the source map)
// and matched against the child's raw file to determine the child-relative
// 1-indexed content line number. Header normalization (# → ## etc.) is
// handled by comparing lines with leading '#' characters stripped.
func (gui *Gui) buildComposeLineMap(contentLines []string, sourceMap []models.SourceMapEntry, identities []lineIdentity) {
	// Pre-load each child's content lines (after frontmatter)
	type childData struct {
		lines []string // raw file lines after frontmatter
	}
	children := make(map[string]*childData)
	for _, entry := range sourceMap {
		if _, ok := children[entry.UUID]; ok {
			continue
		}
		data, err := os.ReadFile(entry.Path)
		if err != nil {
			children[entry.UUID] = &childData{}
			continue
		}
		fileLines := strings.Split(string(data), "\n")
		cs := skipFrontmatter(fileLines)
		children[entry.UUID] = &childData{lines: fileLines[cs:]}
	}

	for _, entry := range sourceMap {
		child := children[entry.UUID]
		if child == nil {
			continue
		}

		// Forward-scan the child's content to match composed lines
		childIdx := 0
		for srcIdx := entry.StartLine - 1; srcIdx < entry.EndLine && srcIdx < len(contentLines); srcIdx++ {
			composedTrimmed := strings.TrimSpace(contentLines[srcIdx])
			if composedTrimmed == "" {
				// Blank line — set child UUID/Path but LineNum=0 (non-resolvable)
				identities[srcIdx] = lineIdentity{uuid: entry.UUID, path: entry.Path}
				continue
			}

			composedNorm := normalizeLineForMatch(composedTrimmed)
			matched := false
			for childIdx < len(child.lines) {
				childTrimmed := strings.TrimSpace(child.lines[childIdx])
				childNorm := normalizeLineForMatch(childTrimmed)
				if composedNorm != "" && composedNorm == childNorm {
					identities[srcIdx] = lineIdentity{
						uuid:    entry.UUID,
						lineNum: childIdx + 1, // 1-indexed content line in child file
						path:    entry.Path,
					}
					childIdx++
					matched = true
					break
				}
				childIdx++
			}
			if !matched {
				// No match found — set child UUID/Path but LineNum=0
				identities[srcIdx] = lineIdentity{uuid: entry.UUID, path: entry.Path}
			}
		}
	}
}

// normalizeLineForMatch strips leading markdown decoration (header hashes,
// list bullet, blockquote marker, task checkbox) from a trimmed line so
// composed output can be forward-matched against its source note file.
// Compose normalizes headers (# → ##), pick prefixes extracted lines with
// "- ", and task-list renderers add "[ ]"/"[x]" — all of these decorate
// the source text without changing its identity. Applying symmetrically
// to both composed and source lines keeps the comparison stable.
func normalizeLineForMatch(s string) string {
	s = stripHeaderPrefix(s)
	s = stripListPrefix(s)
	s = stripTaskCheckbox(s)
	return s
}

// stripHeaderPrefix removes leading '#' characters and the following space
// from a string, normalizing markdown headers for cross-level comparison
// (e.g. "## Title" and "### Title" both become "Title").
func stripHeaderPrefix(s string) string {
	i := 0
	for i < len(s) && s[i] == '#' {
		i++
	}
	if i == 0 {
		return s
	}
	return strings.TrimLeft(s[i:], " ")
}

// stripListPrefix removes a single leading list bullet ("-", "*", "+") or
// blockquote marker (">") plus the following space. Ordered-list markers
// like "1." are left alone — they carry identity (the number) and are
// rarely injected as decoration.
func stripListPrefix(s string) string {
	if len(s) < 2 {
		return s
	}
	switch s[0] {
	case '-', '*', '+', '>':
		if s[1] == ' ' {
			return s[2:]
		}
	}
	return s
}

// stripTaskCheckbox removes a leading "[ ]" or "[x]" (case-insensitive)
// followed by a space, which markdown renderers inject for task lists.
func stripTaskCheckbox(s string) string {
	if len(s) < 4 || s[0] != '[' || s[2] != ']' || s[3] != ' ' {
		return s
	}
	switch s[1] {
	case ' ', 'x', 'X':
		return s[4:]
	}
	return s
}

// renderCardInto renders a single note card (upper separator, body, lower separator)
// and fills in ns.CardLineRanges[cardIdx]. Appends SourceLines to ns.Lines.
// Returns the updated currentLine.
func (gui *Gui) renderCardInto(v *gocui.View, note models.Note, cardIdx int,
	ns *context.PreviewNavState, currentLine int, isActive bool,
	selectedIdx int, temporarilyMoved map[int]bool, width int, contentWidth int) int {

	selected := isActive && cardIdx == selectedIdx
	ns.CardLineRanges[cardIdx][0] = currentLine

	emit := func(text string, sl types.SourceLine) {
		gui.fprintPreviewLine(v, text, currentLine, isActive, ns)
		sl.Text = text
		ns.Lines = append(ns.Lines, sl)
		currentLine++
	}

	title := note.Title
	if title == "" {
		title = "Untitled"
	}
	upperRight := ""
	if temporarilyMoved != nil && temporarilyMoved[cardIdx] {
		upperRight = " Temporarily Moved "
	}
	emit(gui.buildSeparatorLine(true, " "+title+" ", upperRight, width, selected), types.SourceLine{})

	for _, sl := range gui.BuildCardContent(note, contentWidth) {
		if isHeaderLine(sl.Text) {
			ns.HeaderLines = append(ns.HeaderLines, currentLine)
		}
		emit(sl.Text, sl)
	}

	var parentLabel string
	if note.Parent != "" {
		parentLabel = gui.resolveParentLabel(note.Parent)
	}
	rightText := ""
	if meta := models.JoinDot(note.ShortDate(), note.GlobalTagsString(), parentLabel); meta != "" {
		rightText = " " + meta + " "
	}
	emit(gui.buildSeparatorLine(false, "", rightText, width, selected), types.SourceLine{})

	ns.CardLineRanges[cardIdx][1] = currentLine
	return currentLine
}

// renderPickGroupInto renders a single pick result group (upper separator, matches, lower separator)
// and fills in ns.CardLineRanges[cardIdx]. Appends SourceLines to ns.Lines.
// Returns the updated currentLine.
func (gui *Gui) renderPickGroupInto(v *gocui.View, result models.PickResult, cardIdx int,
	ns *context.PreviewNavState, currentLine int, isActive bool,
	selectedIdx int, width int, contentWidth int) int {

	selected := isActive && cardIdx == selectedIdx
	ns.CardLineRanges[cardIdx][0] = currentLine

	emit := func(text string, sl types.SourceLine) {
		gui.fprintPreviewLine(v, text, currentLine, isActive, ns)
		sl.Text = text
		ns.Lines = append(ns.Lines, sl)
		currentLine++
	}

	title := result.Title
	if title == "" {
		title = "Untitled"
	}
	emit(gui.buildSeparatorLine(true, " "+title+" ", "", width, selected), types.SourceLine{})

	dimDone := gui.contexts.ActivePreview().DisplayState().DimDone
	for _, match := range result.Matches {
		lineNum := fmt.Sprintf("%02d", match.Line)
		prefix := fmt.Sprintf("  L%s: ", lineNum)
		prefixLen := len(prefix)
		highlighted := gui.highlightMarkdown(match.Content)
		wrapped := wordwrap.String(highlighted, contentWidth-prefixLen)
		indent := strings.Repeat(" ", prefixLen)
		src := types.SourceLine{UUID: result.UUID, LineNum: match.Line, Path: result.File}
		isDone := dimDone && match.Done
		for j, line := range strings.Split(strings.TrimRight(wrapped, "\n"), "\n") {
			var formatted string
			if j == 0 {
				formatted = fmt.Sprintf("  %sL%s:%s %s", AnsiDim, lineNum, AnsiReset, line)
			} else {
				formatted = indent + line
			}
			if isDone {
				formatted = dimLine(formatted)
			}
			emit(formatted, src)
		}
	}

	matchCount := fmt.Sprintf(" %d matches ", len(result.Matches))
	emit(gui.buildSeparatorLine(false, "", matchCount, width, selected), types.SourceLine{})

	ns.CardLineRanges[cardIdx][1] = currentLine
	return currentLine
}

// renderSeparatorCards renders cards using separator lines instead of frames
func (gui *Gui) renderSeparatorCards(v *gocui.View, cards []models.Note, ns *context.PreviewNavState, temporarilyMoved map[int]bool) {
	if len(cards) == 0 {
		fmt.Fprintln(v, "No matching notes.")
		return
	}

	ctx := gui.contexts.ActivePreview()
	width, _ := v.InnerSize()
	if width < 10 {
		width = 40
	}
	contentWidth := types.PreviewContentWidth(v)

	isActive := gui.isPreviewActive()
	currentLine := 0
	ns.CardLineRanges = make([][2]int, len(cards))
	ns.HeaderLines = ns.HeaderLines[:0]
	ns.Lines = ns.Lines[:0]

	for i, note := range cards {
		currentLine = gui.renderCardInto(v, note, i, ns, currentLine, isActive, ctx.SelectedCardIndex(), temporarilyMoved, width, contentWidth)
		if i < len(cards)-1 {
			gui.fprintPreviewLine(v, "", currentLine, isActive, ns)
			ns.Lines = append(ns.Lines, types.SourceLine{Text: ""})
			currentLine++
		}
	}

	// Scroll to keep cursor/card visible
	_, viewHeight := v.InnerSize()
	originY := ns.ScrollOffset
	if isActive {
		cl := ns.CursorLine
		idx := ctx.SelectedCardIndex()
		showFrom := cl
		showTo := cl
		if idx < len(ns.CardLineRanges) {
			r := ns.CardLineRanges[idx]
			if cl == r[0]+1 {
				showFrom = r[0]
			}
			if cl == r[1]-2 {
				showTo = r[1] - 1
			}
		}
		if showFrom < originY {
			originY = showFrom
		} else if showTo >= originY+viewHeight {
			originY = showTo - viewHeight + 1
		}
	} else {
		idx := ctx.SelectedCardIndex()
		if idx < len(ns.CardLineRanges) {
			r := ns.CardLineRanges[idx]
			if r[0] < originY {
				originY = r[0]
			} else if r[1] > originY+viewHeight {
				originY = r[1] - viewHeight
			}
		}
	}
	ns.ScrollOffset = originY
	v.SetOrigin(0, originY)
}

// renderSectionHeader renders a section divider and a blank spacer line.
// Records the divider's line number in dp.SectionHeaderLines.
// Returns the updated currentLine.
func (gui *Gui) renderSectionHeader(v *gocui.View, label string, width int,
	currentLine int, ns *context.PreviewNavState, dp *context.DatePreviewState, isActive bool) int {

	dp.SectionHeaderLines = append(dp.SectionHeaderLines, currentLine)
	line := gui.buildStraightSeparator(" "+label+" ", width)
	gui.fprintPreviewLine(v, line, currentLine, isActive, ns)
	ns.Lines = append(ns.Lines, types.SourceLine{Text: line})
	currentLine++
	gui.fprintPreviewLine(v, "", currentLine, isActive, ns)
	ns.Lines = append(ns.Lines, types.SourceLine{Text: ""})
	currentLine++
	return currentLine
}

// renderDatePreview renders three sections (inline tags, todos, notes) into a unified line space.
func (gui *Gui) renderDatePreview(v *gocui.View, dp *context.DatePreviewContext, ns *context.PreviewNavState, isActive bool) {
	width, _ := v.InnerSize()
	if width < 10 {
		width = 40
	}
	contentWidth := types.PreviewContentWidth(v)

	totalCards := len(dp.TagPicks) + len(dp.TodoPicks) + len(dp.Notes)
	if totalCards == 0 {
		fmt.Fprintln(v, " "+AnsiDim+"No activity on "+dp.TargetDate+AnsiReset)
		ns.CardLineRanges = nil
		ns.Lines = []types.SourceLine{{Text: "No activity on " + dp.TargetDate}}
		return
	}

	currentLine := 0
	ns.CardLineRanges = make([][2]int, totalCards)
	ns.HeaderLines = ns.HeaderLines[:0]
	ns.Lines = ns.Lines[:0]
	dp.SectionHeaderLines = dp.SectionHeaderLines[:0]

	cardIdx := 0

	// --- Section 1: Inline Tags ---
	sectionLineStart := currentLine
	currentLine = gui.renderSectionHeader(v, "Inline Tags", width, currentLine, ns, dp.DatePreviewState, isActive)
	tagStart := cardIdx
	if len(dp.TagPicks) == 0 {
		gui.fprintPreviewLine(v, " "+AnsiDim+"No tagged lines"+AnsiReset, currentLine, isActive, ns)
		ns.Lines = append(ns.Lines, types.SourceLine{Text: " No tagged lines"})
		currentLine++
	} else {
		for i, result := range dp.TagPicks {
			currentLine = gui.renderPickGroupInto(v, result, cardIdx, ns, currentLine, isActive, dp.SelectedCardIdx, width, contentWidth)
			if i < len(dp.TagPicks)-1 {
				gui.fprintPreviewLine(v, "", currentLine, isActive, ns)
				ns.Lines = append(ns.Lines, types.SourceLine{Text: ""})
				currentLine++
			}
			cardIdx++
		}
	}
	dp.SectionRanges[0] = [2]int{tagStart, cardIdx}
	// Blank line after section
	gui.fprintPreviewLine(v, "", currentLine, isActive, ns)
	ns.Lines = append(ns.Lines, types.SourceLine{Text: ""})
	currentLine++
	dp.SectionLineRanges[0] = [2]int{sectionLineStart, currentLine}

	// --- Section 2: Todos ---
	sectionLineStart = currentLine
	currentLine = gui.renderSectionHeader(v, "Todos", width, currentLine, ns, dp.DatePreviewState, isActive)
	todoStart := cardIdx
	if len(dp.TodoPicks) == 0 {
		gui.fprintPreviewLine(v, " "+AnsiDim+"No todos"+AnsiReset, currentLine, isActive, ns)
		ns.Lines = append(ns.Lines, types.SourceLine{Text: " No todos"})
		currentLine++
	} else {
		for i, result := range dp.TodoPicks {
			currentLine = gui.renderPickGroupInto(v, result, cardIdx, ns, currentLine, isActive, dp.SelectedCardIdx, width, contentWidth)
			if i < len(dp.TodoPicks)-1 {
				gui.fprintPreviewLine(v, "", currentLine, isActive, ns)
				ns.Lines = append(ns.Lines, types.SourceLine{Text: ""})
				currentLine++
			}
			cardIdx++
		}
	}
	dp.SectionRanges[1] = [2]int{todoStart, cardIdx}
	gui.fprintPreviewLine(v, "", currentLine, isActive, ns)
	ns.Lines = append(ns.Lines, types.SourceLine{Text: ""})
	currentLine++
	dp.SectionLineRanges[1] = [2]int{sectionLineStart, currentLine}

	// --- Section 3: Notes ---
	sectionLineStart = currentLine
	currentLine = gui.renderSectionHeader(v, "Notes", width, currentLine, ns, dp.DatePreviewState, isActive)
	noteStart := cardIdx
	if len(dp.Notes) == 0 {
		gui.fprintPreviewLine(v, " "+AnsiDim+"No notes"+AnsiReset, currentLine, isActive, ns)
		ns.Lines = append(ns.Lines, types.SourceLine{Text: " No notes"})
		currentLine++
	} else {
		for i, note := range dp.Notes {
			currentLine = gui.renderCardInto(v, note, cardIdx, ns, currentLine, isActive, dp.SelectedCardIdx, nil, width, contentWidth)
			if i < len(dp.Notes)-1 {
				gui.fprintPreviewLine(v, "", currentLine, isActive, ns)
				ns.Lines = append(ns.Lines, types.SourceLine{Text: ""})
				currentLine++
			}
			cardIdx++
		}
	}
	dp.SectionRanges[2] = [2]int{noteStart, cardIdx}
	dp.SectionLineRanges[2] = [2]int{sectionLineStart, currentLine}

	// Scroll management
	_, viewHeight := v.InnerSize()
	originY := ns.ScrollOffset
	if isActive {
		cl := ns.CursorLine
		idx := dp.SelectedCardIdx
		showFrom, showTo := cl, cl
		if idx < len(ns.CardLineRanges) {
			r := ns.CardLineRanges[idx]
			if cl == r[0]+1 {
				showFrom = r[0]
			}
			if cl == r[1]-2 {
				showTo = r[1] - 1
			}
		}
		if showFrom < originY {
			originY = showFrom
		} else if showTo >= originY+viewHeight {
			originY = showTo - viewHeight + 1
		}
	} else {
		if len(ns.CardLineRanges) > 0 {
			if ns.CardLineRanges[0][0] < originY {
				originY = 0
			}
		}
	}
	ns.ScrollOffset = originY
	v.SetOrigin(0, originY)
}

// truncWithEllipsis truncates s to at most maxRunes runes. If truncation is
// needed, the result ends in "…" (preserving a trailing space when s had
// one, so " Long title " becomes " Long…  "). Returns s unchanged when it
// already fits, and an empty string when maxRunes < 1.
func truncWithEllipsis(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	if maxRunes < 1 {
		return ""
	}
	trailing := ""
	if runes[len(runes)-1] == ' ' {
		trailing = " "
	}
	// Keep: (maxRunes - 1 for "…" - len(trailing)) head runes + "…" + trailing
	keep := maxRunes - 1 - len([]rune(trailing))
	if keep < 0 {
		return "…"
	}
	return string(runes[:keep]) + "…" + trailing
}

// buildSeparatorLine creates a separator line with optional left and right text
func (gui *Gui) buildStraightSeparator(label string, width int) string {
	sep := "─"
	labelLen := len([]rune(label))
	fillLen := max(width-labelLen-2, 0)
	var sb strings.Builder
	sb.WriteString(AnsiDim)
	sb.WriteString(sep)
	sb.WriteString(label)
	for i := 0; i < fillLen; i++ {
		sb.WriteString(sep)
	}
	sb.WriteString(sep)
	sb.WriteString(AnsiReset)
	return sb.String()
}

func (gui *Gui) buildSeparatorLine(upper bool, leftText, rightText string, width int, highlight bool) string {
	dim := AnsiDim
	green := AnsiGreen
	reset := AnsiReset

	sep := "─"
	rightLen := len([]rune(rightText))

	// Trim leftText so the whole separator fits in width. Fixed overhead is
	// 4 runes (corner + sep on each side). Everything else is budget; any
	// overflow makes the terminal wrap the frame to the next row, mangling
	// card borders — so we truncate the left text with an "…" marker.
	leftBudget := width - 4 - rightLen
	leftText = truncWithEllipsis(leftText, leftBudget)
	leftLen := len([]rune(leftText))

	// Calculate fill length
	fillLen := max(width-leftLen-rightLen-4, 0)

	var sb strings.Builder
	sb.WriteString(reset) // Clear any leftover foreground color from content lines
	if highlight {
		sb.WriteString(green)
	}
	sb.WriteString(dim)
	if upper {
		sb.WriteString("╭")
	} else {
		sb.WriteString("╰")
	}
	sb.WriteString(sep)
	sb.WriteString(leftText)
	for i := 0; i < fillLen; i++ {
		sb.WriteString(sep)
	}
	sb.WriteString(rightText)
	sb.WriteString(sep)
	if upper {
		sb.WriteString("╮")
	} else {
		sb.WriteString("╯")
	}
	sb.WriteString(reset)

	return sb.String()
}

// resolveParentLabel returns a display name for a parent UUID by checking
// loaded parent bookmarks, then loaded notes, then the title cache, then
// falling back to a truncated UUID.
func (gui *Gui) resolveParentLabel(uuid string) string {
	for _, bm := range gui.contexts.Queries.Parents {
		if bm.UUID == uuid {
			return bm.Name
		}
	}
	for _, note := range gui.contexts.Notes.Items {
		if note.UUID == uuid {
			return note.Title
		}
	}
	if gui.helpers != nil {
		if t, ok := gui.helpers.TitleCache().Get(uuid); ok {
			return t
		}
	}
	// Fallback: show truncated UUID
	if len(uuid) > 8 {
		return uuid[:8] + "..."
	}
	return uuid
}

// renderPickResults renders line-level pick results grouped by note title.
// selectedCardIdx and isActive are passed explicitly so both the main preview
// and the pick dialog overlay can share this rendering logic.
func (gui *Gui) renderPickResults(v *gocui.View, results []models.PickResult, ns *context.PreviewNavState, selectedCardIdx int, isActive bool) {
	if len(results) == 0 {
		fmt.Fprintln(v, "No matching lines.")
		return
	}

	width, _ := v.InnerSize()
	if width < 10 {
		width = 40
	}
	contentWidth := types.PreviewContentWidth(v)

	currentLine := 0
	ns.CardLineRanges = make([][2]int, len(results))
	ns.HeaderLines = ns.HeaderLines[:0]
	ns.Lines = ns.Lines[:0]

	for i, result := range results {
		currentLine = gui.renderPickGroupInto(v, result, i, ns, currentLine, isActive, selectedCardIdx, width, contentWidth)
		if i < len(results)-1 {
			gui.fprintPreviewLine(v, "", currentLine, isActive, ns)
			ns.Lines = append(ns.Lines, types.SourceLine{Text: ""})
			currentLine++
		}
	}

	// Scroll to keep cursor/group visible
	_, viewHeight := v.InnerSize()
	originY := ns.ScrollOffset
	if isActive {
		cl := ns.CursorLine
		if cl < originY {
			originY = cl
		} else if cl >= originY+viewHeight {
			originY = cl - viewHeight + 1
		}
	} else {
		if selectedCardIdx < len(ns.CardLineRanges) {
			r := ns.CardLineRanges[selectedCardIdx]
			if r[0] < originY {
				originY = r[0]
			} else if r[1] > originY+viewHeight {
				originY = r[1] - viewHeight
			}
		}
	}
	ns.ScrollOffset = originY
	v.SetOrigin(0, originY)
}

// skipFrontmatter returns the 0-indexed line of the first content line in fileLines.
// If no frontmatter is present, returns 0.
func skipFrontmatter(fileLines []string) int {
	if len(fileLines) == 0 || !strings.HasPrefix(fileLines[0], "---") {
		return 0
	}
	for i := 1; i < len(fileLines); i++ {
		if strings.TrimSpace(fileLines[i]) == "---" {
			return i + 1
		}
	}
	return 0 // unclosed frontmatter — treat whole file as content
}

// loadNoteFrontmatter returns the raw YAML frontmatter block (without the --- delimiters).
func (gui *Gui) loadNoteFrontmatter(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	fileLines := strings.Split(string(data), "\n")
	contentStart := skipFrontmatter(fileLines)
	if contentStart <= 1 {
		return "", nil // no frontmatter
	}
	// frontmatter is lines 1..contentStart-2 (between the --- delimiters)
	fm := strings.Join(fileLines[1:contentStart-1], "\n")
	return fm, nil
}

func (gui *Gui) loadNoteContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	fileLines := strings.Split(string(data), "\n")
	contentStart := skipFrontmatter(fileLines)
	content := strings.Join(fileLines[contentStart:], "\n")
	content = strings.TrimLeft(content, "\n")
	return content, nil
}
