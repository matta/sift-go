package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type todo struct {
	Title string
	Done  bool
}

func samples() []todo {
	return []todo{
		{Title: "todo 1", Done: true},
		{Title: "todo 2", Done: false},
		{Title: "todo 3", Done: true},
		{Title: "todo 4", Done: false},
	}
}

type addItemMessage struct {
	Title string
}

type model struct {
	items    []todo
	cursor   int
	selected map[int]struct{}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "p", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "n", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.items[m.cursor].Done = !m.items[m.cursor].Done
		}
	case addItemMessage:
		m.items = append(m.items, todo{Title: msg.Title})
	}
	return m, nil
}

func (m model) View() string {
	s := ""
	for i, item := range m.items {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		done := " "
		if m.items[i].Done {
			done = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, done, item.Title)
	}

	s += "\nPress q to quit.\n"
	return s
}

func main() {
	initialModel := model{
		items:    samples(),
		selected: make(map[int]struct{}),
	}

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
