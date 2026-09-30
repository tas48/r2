package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/tas48/r2/internal/ui"
)

func main() {
	program := tea.NewProgram(ui.New())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "r2: %v\n", err)
		os.Exit(1)
	}
}
