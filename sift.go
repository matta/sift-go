package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/ghodss/yaml"
)

type position struct {
	col int
	row int
}

type extent struct {
	width  int
	height int
}

type bounds struct {
	position
	extent
}

func drawText(s tcell.Screen, b bounds, style tcell.Style, text string) position {
	p := b.position
	for _, r := range text {
		if p.row >= b.row+b.height {
			break
		}
		// TODO: handle word wrapping and wide chars properly.
		s.SetContent(p.col, p.row, r, nil, style)
		p.col++
		if p.col >= b.col+b.width {
			p = position{row: p.row + 1, col: b.col}
		}
	}
	return p
}

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

type model interface {
	Update(screen tcell.Screen, event tcell.Event) model
	Draw(s tcell.Screen)
}

type persistedModel struct {
	Items    []todo
	Cursor   int
	Selected map[int]struct{}
}

type listModel struct {
	persisted persistedModel
	quit      bool
}

func (m *listModel) Update(screen tcell.Screen, event tcell.Event) model {
	switch event := event.(type) {
	case *tcell.EventKey:
		switch {
		case event.Key() == tcell.KeyEscape ||
			event.Key() == tcell.KeyCtrlC ||
			(event.Key() == tcell.KeyRune && event.Rune() == 'q'):
			m.quit = true
		case event.Key() == tcell.KeyRune && event.Rune() == 'k':
			if m.persisted.Cursor > 0 {
				m.persisted.Cursor--
			}
		case event.Key() == tcell.KeyRune && event.Rune() == 'j':
			if m.persisted.Cursor < len(m.persisted.Items)-1 {
				m.persisted.Cursor++
			}
		case event.Key() == tcell.KeyRune && event.Rune() == 'x':
			m.persisted.Items[m.persisted.Cursor].Done = !m.persisted.Items[m.persisted.Cursor].Done
		case event.Key() == tcell.KeyRune && event.Rune() == 'a':
			return &addModel{
				list: m,
			}
		}
	}
	return m
}

func (m *listModel) Draw(s tcell.Screen) {
	style := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	screenExtent := ScreenExtent(s)
	for i, item := range m.persisted.Items {
		cursor := " "
		if i == m.persisted.Cursor {
			cursor = ">"
		}

		done := " "
		if m.persisted.Items[i].Done {
			done = "x"
		}

		line := fmt.Sprintf("%s [%s] %s", cursor, done, item.Title)
		drawText(s, bounds{position{col: 0, row: i}, extent{width: screenExtent.width, height: 1}}, style, line)
	}
}

func (m *listModel) Save() error {
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

type addModel struct {
	list   *listModel
	title  string
	events []tcell.Event
}

func (m *addModel) Update(screen tcell.Screen, event tcell.Event) model {
	switch event := event.(type) {
	case *tcell.EventKey:
		m.events = append(m.events, event)
		// If m.events has more than 5 elements remove the first one.
		// TODO: Why? Looks like a hack to prevent unbounded type ahead?
		for len(m.events) > 5 {
			m.events = m.events[1:]
		}
		switch {
		case event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyCtrlC:
			return m.list
		case event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2:
			// Remove the last rune from m.title
			if len(m.title) > 0 {
				m.title = m.title[:len(m.title)-1]
			}
		case event.Key() == tcell.KeyRune:
			m.title += string(event.Rune())
		case event.Key() == tcell.KeyEnter:
			m.list.persisted.Items = append(m.list.persisted.Items, todo{Title: m.title, Done: false})
			return m.list
		}
	}
	return m
}

func ScreenExtent(s tcell.Screen) extent {
	width, height := s.Size()
	return extent{width: width, height: height}
}

func (m *addModel) Draw(s tcell.Screen) {
	screenSize := ScreenExtent(s)
	line := fmt.Sprintf("Add new todo with title: %s", m.title)
	p := drawText(s, bounds{position{0, 0}, screenSize}, tcell.StyleDefault, line)
	s.ShowCursor(p.col, p.row)

	for _, e := range m.events {
		if p.col != 0 {
			p.col = 0
			p.row++
		}
		extent := screenSize
		extent.height -= p.row
		end := drawText(s, bounds{p, extent}, tcell.StyleDefault, fmt.Sprintf("%+v", e))
		p.row = end.row
	}
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

func InitialModel() listModel {
	return listModel{
		persisted: persistedModel{
			Items:    samples(),
			Selected: make(map[int]struct{}),
		},
	}
}

func LoadModel() listModel {
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
	listModel := LoadModel()
	var model model = &listModel

	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Init(); err != nil {
		log.Fatal(err)
	}

	// Set default text style
	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	s.SetStyle(defStyle)

	s.Clear()

	wasResize := false
	for !listModel.quit {
		// Update screen
		s.Clear()
		model.Draw(s)
		if wasResize {
			s.Sync()
			wasResize = false
		} else {
			s.Show()
		}

		// Poll event
		ev := s.PollEvent()

		// Process event
		switch ev.(type) {
		case *tcell.EventResize:
			wasResize = true
		}
		model = model.Update(s, ev)
	}
	s.Fini()
	if err := listModel.Save(); err != nil {
		panic(err)
	}
	os.Exit(0)
}
