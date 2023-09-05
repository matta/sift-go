package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	Items    []todo
	Cursor   int
	Selected map[int]struct{}
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
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "n", "j":
			if m.Cursor < len(m.Items)-1 {
				m.Cursor++
			}
		case "enter", " ":
			m.Items[m.Cursor].Done = !m.Items[m.Cursor].Done
		case "a":
			m.Items = append(m.Items, todo{Title: ""})
		}
	case addItemMessage:
		m.Items = append(m.Items, todo{Title: msg.Title})
	}
	return m, nil
}

func (m model) View() string {
	s := ""
	for i, item := range m.Items {
		cursor := " "
		if i == m.Cursor {
			cursor = ">"
		}

		done := " "
		if m.Items[i].Done {
			done = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, done, item.Title)
	}

	s += "\nPress q to quit.\n"
	return s
}

func UserHomeDir() string {
	usr, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return usr
}

func UserDataFile() string {
	return filepath.Join(UserHomeDir(), ".sift.json")
}

func (m model) Save() error {
	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal model: %w", err)
	}
	err = os.WriteFile(UserDataFile(), b, 0600)
	if err != nil {
		return fmt.Errorf("failed to save model: %w", err)
	}
	return nil
}

func InitialModel() model {
	return model{
		Items:    samples(),
		Selected: make(map[int]struct{}),
	}
}

func LoadModel() model {
	b, err := os.ReadFile(UserDataFile())
	if err != nil {
		log.Printf("Failed to read model file: %v", err)
		return InitialModel()
	}

	var m model
	err = json.Unmarshal(b, &m)
	if err != nil {
		log.Printf("Failed to unmarshal model file: %v", err)
		return InitialModel()
	}

	return m
}

func main() {
	p := tea.NewProgram(LoadModel())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
	finalModel := m.(model)
	if err := finalModel.Save(); err != nil {
		panic(err)
	}
}
