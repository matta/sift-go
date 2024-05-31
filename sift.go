package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/matta/sift/internal/replicatedtodo"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ghodss/yaml"
)

type teaModel struct {
	wrapped *model
}

type model struct {
	keys      keyMap
	help      help.Model
	persisted replicatedtodo.Model
	ids       []string
	cursor    int
	textInput textinput.Model
}

// Init implements tea.Model.
func (outer teaModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (outer teaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return outer, outer.wrapped.update(msg)
}

// View implements tea.Model.
func (outer teaModel) View() string {
	return outer.wrapped.view()
}

func newModel() *model {
	return &model{
		keys:      keys,
		help:      help.New(),
		persisted: replicatedtodo.New(),
		textInput: textinput.New(),
		ids:       make([]string, 0),
		cursor:    0,
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
	bytes, err := yaml.Marshal(m.persisted)
	if err != nil {
		return fmt.Errorf("failed to marshal model: %w", err)
	}

	var PERM = 0600
	err = os.WriteFile(UserDataFile(), bytes, os.FileMode(PERM))

	if err != nil {
		return fmt.Errorf("failed to save model: %w", err)
	}

	return nil
}

func updateTextInput(input *textinput.Model, msg tea.Msg) tea.Cmd {
	temp, cmd := input.Update(msg)
	*input = temp

	return cmd
}

func (m *model) update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// If we set a width on our sub-models, so they can respond as needed.
		m.help.Width = msg.Width
		m.textInput.Width = msg.Width
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	return nil
}

func (m *model) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch {
	case m.textInput.Focused():
		return m.handleFocusedTextInput(msg)

	case key.Matches(msg, m.keys.Help):
		m.help.ShowAll = !m.help.ShowAll

	case key.Matches(msg, m.keys.Up):
		m.cursorUp()

	case key.Matches(msg, m.keys.Down):
		m.cursorDown()

	case key.Matches(msg, m.keys.Toggle):
		m.persisted.ToggleDone(m.ids[m.cursor])

	case key.Matches(msg, m.keys.Add):
		return m.textInput.Focus()

	case key.Matches(msg, m.keys.Quit):
		return tea.Quit
	}

	return nil
}

func (m *model) cursorDown() {
	if m.cursor < len(m.persisted.Items)-1 {
		m.cursor++
	}
}

func (m *model) cursorUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

func (m *model) handleFocusedTextInput(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.resetTextInput()

	case key.Matches(msg, m.keys.Accept):
		m.accept()

	default:
		return updateTextInput(&m.textInput, msg)
	}

	return nil
}

func (m *model) accept() {
	title := m.textInput.Value()
	if title != "" {
		m.newTodo(title)
	}

	m.resetTextInput()
}

func (m *model) resetTextInput() {
	m.textInput.Reset()
	m.textInput.Blur()
}

func (m *model) view() string {
	out := ""

	for itemIndex, item := range m.persisted.Items {
		cursor := " "
		if itemIndex == m.cursor {
			cursor = ">"
		}

		done := " "
		if m.persisted.GetState(m.ids[itemIndex]) == "checked" {
			done = "x"
		}

		out += fmt.Sprintf("%s [%s] %s\n", cursor, done, item.Title)
	}

	if m.textInput.Focused() {
		out += m.textInput.View()
	} else {
		out += m.help.View(m.keys)
	}

	return out
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

func loadModel() teaModel {
	model := loadPersistedModel()

	for _, item := range model.persisted.Items {
		if model.persisted.GetState(item.ID) != "removed" {
			model.ids = append(model.ids, item.ID)
		}
	}

	return teaModel{model}
}

func loadPersistedModel() *model {
	model := newModel()

	bytes, err := os.ReadFile(UserDataFile())
	if err != nil {
		log.Printf("Failed to read model file: %v", err)
		model.addSampleItems()

		return model
	}

	var replicatedModel replicatedtodo.Model
	if err = yaml.Unmarshal(bytes, &replicatedModel); err != nil {
		log.Printf("Failed to unmarshal model file: %v", err)
		model.addSampleItems()

		return model
	}

	model.persisted = replicatedModel

	return model
}

func main() {
	teaModel := loadModel()
	program := tea.NewProgram(loadModel())
	_, err := program.Run()

	if err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	if err := teaModel.wrapped.save(); err != nil {
		panic(err)
	}
}
