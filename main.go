package main

import (
	"flag"
	"fmt"
	"os"

	"kvnd/lazyruin/pkg/app"
)

func main() {
	vaultPath := flag.String("vault", "", "Path to the ruin vault")
	ruinBin := flag.String("ruin", "", "Path to the ruin binary")
	newNote := flag.Bool("new", false, "Open directly into new note capture, exit on save")
	flag.Parse()

	a, err := app.NewApp(*vaultPath, *ruinBin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	a.QuickCapture = *newNote

	if err := a.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
