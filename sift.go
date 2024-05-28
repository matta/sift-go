package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ghodss/yaml"
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

type persistedModel struct {
	Items    []todo
	Cursor   int
	Selected map[int]struct{}
}

type model struct {
	keys      keyMap
	help      help.Model
	persisted persistedModel
	textInput textinput.Model
}

// keyMap holds a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Toggle key.Binding
	Add    key.Binding
	Help   key.Binding
	Quit   key.Binding
	Cancel key.Binding
	Accept key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Toggle, k.Add}, // first column
		{k.Help, k.Quit},                // second column
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Toggle: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "toggle item"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add item"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Cancel: key.NewBinding(key.WithKeys("esc")),
	Accept: key.NewBinding(
		key.WithKeys("enter")),
}

// Init implements tea.Model.
func (m model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle inconvenience of textinput.Update not taking a pointer receiver.
	updateTextInput := func(input *textinput.Model, msg tea.Msg) tea.Cmd {
		temp, cmd := input.Update(msg)
		*input = temp
		return cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// If we set a width on our sub-models so they can respond as needed.
		m.help.Width = msg.Width
		m.textInput.Width = msg.Width
	case tea.KeyMsg:
		if m.textInput.Focused() {
			switch {
			case key.Matches(msg, m.keys.Accept):
				if m.textInput.Value() != "" {
					m.persisted.Items = append(m.persisted.Items, todo{Title: m.textInput.Value()})
				}
				fallthrough

			case key.Matches(msg, m.keys.Cancel):
				m.textInput.Reset()
				m.textInput.Blur()

			default:
				cmd := updateTextInput(&m.textInput, msg)
				return m, cmd
			}
		} else {
			switch {
			case key.Matches(msg, m.keys.Help):
				m.help.ShowAll = !m.help.ShowAll
			case key.Matches(msg, m.keys.Up):
				if m.persisted.Cursor > 0 {
					m.persisted.Cursor--
				}
			case key.Matches(msg, m.keys.Down):
				if m.persisted.Cursor < len(m.persisted.Items)-1 {
					m.persisted.Cursor++
				}

			case key.Matches(msg, m.keys.Toggle):
				m.persisted.Items[m.persisted.Cursor].Done = !m.persisted.Items[m.persisted.Cursor].Done

			case key.Matches(msg, m.keys.Add):
				cmd := m.textInput.Focus()
				return m, cmd

			case key.Matches(msg, m.keys.Quit):
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

// Update implements tea.Model.
func (m model) View() string {
	s := ""
	for i, item := range m.persisted.Items {
		cursor := " "
		if i == m.persisted.Cursor {
			cursor = ">"
		}

		done := " "
		if m.persisted.Items[i].Done {
			done = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, done, item.Title)
	}

	if m.textInput.Focused() {
		s += m.textInput.View()
	}

	helpView := m.help.View(m.keys)
	return s + "\n" + helpView
}

func (m *model) Save() error {
	b, err := yaml.Marshal(m.persisted)
	if err != nil {
		return fmt.Errorf("failed to marshal model: %w", err)
	}
	err = os.WriteFile(UserDataFile(), b, 0600)
	if err != nil {
		return fmt.Errorf("failed to save model: %w", err)
	}
	return nil
}

func UserHomeDir() string {
	usr, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return usr
}

func UserDataFile() string {
	return filepath.Join(UserHomeDir(), ".sift.yaml")
}

func InitialModel() model {
	return model{
		keys: keys,
		help: help.New(),
		persisted: persistedModel{
			Items:    samples(),
			Selected: make(map[int]struct{}),
		},
		textInput: textinput.New(),
	}
}

func LoadModel() model {
	m := InitialModel()

	b, err := os.ReadFile(UserDataFile())
	if err != nil {
		log.Printf("Failed to read model file: %v", err)
		return m
	}

	var p persistedModel
	err = yaml.Unmarshal(b, &p)
	if err != nil {
		log.Printf("Failed to unmarshal model file: %v", err)
		return m
	}

	m.persisted = p
	return m
}

func main() {
	p := tea.NewProgram(LoadModel(), tea.WithAltScreen())
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
