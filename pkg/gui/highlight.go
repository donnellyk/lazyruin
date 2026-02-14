package gui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/jesseduffield/gocui"
	"github.com/muesli/reflow/wordwrap"
)

// highlightMarkdown applies chroma syntax highlighting to markdown content
// without wrapping. Returns the original content unchanged on any error.
func (gui *Gui) highlightMarkdown(content string) string {
	lexer := lexers.Get("markdown")
	if lexer == nil {
		return content
	}
	lexer = chroma.Coalesce(lexer)

	styleName := gui.config.ChromaTheme
	if styleName == "" {
		styleName = "catppuccin-mocha"
		if !gui.darkBackground {
			styleName = "catppuccin-latte"
		}
	}
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		return content
	}

	iter, err := lexer.Tokenise(nil, content)
	if err != nil {
		return content
	}

	tokens := highlightWikilinks(iter.Tokens())

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, chroma.Literator(tokens...)); err != nil {
		return content
	}

	return strings.TrimRight(buf.String(), "\n")
}

func (gui *Gui) renderMarkdown(content string, width int) string {
	highlighted := gui.highlightMarkdown(content)
	wrapped := wordwrap.String(highlighted, width)
	return strings.TrimRight(wrapped, "\n")
}

// highlightWikilinks post-processes chroma tokens to style [[wikilinks]] as links.
func highlightWikilinks(tokens []chroma.Token) []chroma.Token {
	var result []chroma.Token
	for _, tok := range tokens {
		if !strings.Contains(tok.Value, "[[") {
			result = append(result, tok)
			continue
		}
		result = append(result, splitWikilinks(tok)...)
	}
	return result
}

// renderCaptureTextArea replaces v.RenderTextArea() for the capture view,
// applying chroma syntax highlighting to the content.
func (gui *Gui) renderCaptureTextArea(v *gocui.View) {
	v.Clear()
	content := v.TextArea.GetContent()
	fmt.Fprint(v, gui.highlightMarkdown(content))

	cursorX, cursorY := v.TextArea.GetCursorXY()
	prevOriginX, prevOriginY := v.Origin()
	width, height := v.InnerWidth(), v.InnerHeight()

	newViewCursorX, newOriginX := captureUpdatedCursorAndOrigin(prevOriginX, width, cursorX)
	newViewCursorY, newOriginY := captureUpdatedCursorAndOrigin(prevOriginY, height, cursorY)

	v.SetCursor(newViewCursorX, newViewCursorY)
	v.SetOrigin(newOriginX, newOriginY)
}

// captureUpdatedCursorAndOrigin computes new view cursor and origin positions.
// Replicates the unexported gocui updatedCursorAndOrigin function.
func captureUpdatedCursorAndOrigin(prevOrigin int, size int, cursor int) (int, int) {
	var newViewCursor int
	newOrigin := prevOrigin
	usableSize := size - 1

	if cursor > prevOrigin+usableSize {
		newOrigin = cursor - usableSize
		newViewCursor = usableSize
	} else if cursor < prevOrigin {
		newOrigin = cursor
		newViewCursor = 0
	} else {
		newViewCursor = cursor - prevOrigin
	}

	return newViewCursor, newOrigin
}

// splitWikilinks splits a single token around [[...]] patterns,
// re-emitting the wikilink portions as NameTag (same as markdown links).
func splitWikilinks(tok chroma.Token) []chroma.Token {
	var result []chroma.Token
	val := tok.Value
	for {
		start := strings.Index(val, "[[")
		if start == -1 {
			break
		}
		end := strings.Index(val[start:], "]]")
		if end == -1 {
			break
		}
		end += start + 2
		if start > 0 {
			result = append(result, chroma.Token{Type: tok.Type, Value: val[:start]})
		}
		result = append(result, chroma.Token{Type: chroma.NameTag, Value: val[start:end]})
		val = val[end:]
	}
	if val != "" {
		result = append(result, chroma.Token{Type: tok.Type, Value: val})
	}
	return result
}
