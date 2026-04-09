package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/donnellyk/lazyruin/pkg/app"
)

// version is injected at build time via ldflags (see .goreleaser.yml).
// Dev builds show "dev".
var version = "dev"

// linkFlag is a hybrid bool/string flag.
//   - "--link"        → set=true, url=""    (open the link input popup)
//   - "--link=<url>"  → set=true, url=<url> (skip popup, resolve immediately)
//
// IsBoolFlag() lets the Go flag package accept "--link" without a value.
type linkFlag struct {
	set bool
	url string
}

func (l *linkFlag) String() string   { return l.url }
func (l *linkFlag) IsBoolFlag() bool { return true }
func (l *linkFlag) Set(s string) error {
	l.set = true
	if s != "true" {
		l.url = s
	}
	return nil
}

func main() {
	vaultPath := flag.String("vault", "", "Path to the ruin vault")
	ruinBin := flag.String("ruin", "", "Path to the ruin binary")
	newNote := flag.Bool("new", false, "Open directly into new note capture, exit on save")
	var link linkFlag
	flag.Var(&link, "link", "Open directly into new link capture, exit on save.\n  --link             open the link input popup\n  --link=<url>       skip the popup and resolve <url> immediately")
	debugBindings := flag.Bool("debug-bindings", false, "Print all registered keybindings and exit")
	openRef := flag.String("open", "", "Open a specific note (path/title) or parent bookmark on launch")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("lazyruin %s\n", version)
		return
	}

	a, err := app.NewApp(*vaultPath, *ruinBin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	a.QuickCapture = *newNote
	a.QuickLink = link.set
	a.QuickLinkURL = link.url
	a.DebugBindings = *debugBindings
	a.OpenRef = *openRef

	if err := a.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
