package main

import (
	"fmt"
	"os"

	"kvnd/lazyruin/pkg/app"
)

func main() {
	a, err := app.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := a.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
