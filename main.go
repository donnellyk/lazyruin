package main

import (
	"flag"
	"fmt"
	"os"

	"kvnd/lazyruin/pkg/app"
)

func main() {
	vaultPath := flag.String("vault", "", "Path to the ruin vault")
	flag.Parse()

	a, err := app.NewApp(*vaultPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := a.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
