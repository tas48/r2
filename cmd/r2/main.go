package main

import (
	"fmt"
	"os"

	"github.com/tas48/r2/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "r2: %v\n", err)
		os.Exit(1)
	}
}
