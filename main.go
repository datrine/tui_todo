package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	prog := tea.NewProgram(initModel())
	if _, err := prog.Run(); err != nil {
		log.Fatal(err)
	}
}
