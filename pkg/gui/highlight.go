package gui

import (
	"bytes"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/muesli/reflow/wordwrap"
)

func (gui *Gui) renderMarkdown(content string, width int) string {
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

	wrapped := wordwrap.String(buf.String(), width)
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
