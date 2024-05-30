package main

import (
	"example/user/sift/internal/replicatedtodo"
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

type modelWrapper struct {
	wrapped *model
}

type model struct {
	keys      keyMap
	help      help.Model
	persisted replicatedtodo.Model
	ids       []string
	cursor    int
	selected  map[int]struct{}
	textInput textinput.Model
}

// Init implements tea.Model.
func (outer modelWrapper) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (outer modelWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return outer, outer.wrapped.update(msg)
}

// View implements tea.Model.
func (outer modelWrapper) View() string {
	return outer.wrapped.view()
}

func newModel() *model {
	return &model{
		keys:      keys,
		help:      help.New(),
		persisted: replicatedtodo.New(),
		textInput: textinput.New(),
	}
}

//goland:noinspection GoMixedReceiverTypes
func (m *model) addSampleItems() {
	m.newTodo("todo 1")
	m.newTodo("todo 2")
	m.newTodo("todo 3")
	m.newTodo("todo 4")
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

func (m *model) newTodo(title string) {
	id := m.persisted.NewTodo(title)
	m.ids = append(m.ids, id)
}

func (m *model) save() error {
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

func (m *model) update(msg tea.Msg) tea.Cmd {
	// Handle inconvenience of textinput.Update not taking a pointer receiver.
	updateTextInput := func(input *textinput.Model, msg tea.Msg) tea.Cmd {
		temp, cmd := input.Update(msg)
		*input = temp
		return cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// If we set a width on our sub-models, so they can respond as needed.
		m.help.Width = msg.Width
		m.textInput.Width = msg.Width
	case tea.KeyMsg:
		switch {
		case m.textInput.Focused():
			switch {
			case key.Matches(msg, m.keys.Cancel):
				m.textInput.Reset()
				m.textInput.Blur()

			case key.Matches(msg, m.keys.Accept):
				title := m.textInput.Value()
				if title != "" {
					m.newTodo(title)
				}
				m.textInput.Reset()
				m.textInput.Blur()

			default:
				return updateTextInput(&m.textInput, msg)
			}

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll

		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.persisted.Items)-1 {
				m.cursor++
			}

		case key.Matches(msg, m.keys.Toggle):
			m.persisted.ToggleDone(m.ids[m.cursor])

		case key.Matches(msg, m.keys.Add):
			return m.textInput.Focus()

		case key.Matches(msg, m.keys.Quit):
			return tea.Quit
		}
	}
	return nil
}

func (m *model) view() string {
	s := ""
	for i, item := range m.persisted.Items {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		done := " "
		if m.persisted.GetState(m.ids[i]) == "checked" {
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

func loadModel() modelWrapper {
	m := loadPersistedModel()

	for _, item := range m.persisted.Items {
		if m.persisted.GetState(item.Id) != "removed" {
			m.ids = append(m.ids, item.Id)
		}
	}

	return modelWrapper{m}
}

func loadPersistedModel() *model {
	m := newModel()

	b, err := os.ReadFile(UserDataFile())
	if err != nil {
		log.Printf("Failed to read model file: %v", err)
		m.addSampleItems()
		return m
	}

	var p replicatedtodo.Model
	err = yaml.Unmarshal(b, &p)
	if err != nil {
		log.Printf("Failed to unmarshal model file: %v", err)
		m.addSampleItems()
		return m
	}

	m.persisted = p
	return m
}

func main() {
	p := tea.NewProgram(loadModel())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
	finalModel := m.(modelWrapper)
	if err := finalModel.wrapped.save(); err != nil {
		panic(err)
	}
}
