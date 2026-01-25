package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sergekust/polygo/internal/app"
)

func main() {
	p := tea.NewProgram(app.NewModel())

	if _, err := p.Run(); err != nil {
		fmt.Printf("There's been an error: %v", err)
		os.Exit(1)
	}
}
