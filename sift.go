package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
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
	persisted persistedModel
	quit      bool
}

func (m *model) Update(screen tcell.Screen, event tcell.Event) {
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
			m.persisted.Items = append(m.persisted.Items, todo{Title: ""})
		}
	}
}

func drawText(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	for _, r := range text {
		s.SetContent(col, row, r, nil, style)
		col++
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
	}
}

func (m *model) Draw(s tcell.Screen) {
	style := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
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
		drawText(s, 0, i, 20, i, style, line)
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

func (m model) Save() error {
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

func InitialModel() model {
	return model{
		persisted: persistedModel{
			Items:    samples(),
			Selected: make(map[int]struct{}),
		},
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
	model := LoadModel()

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

	wantSync := false
	for !model.quit {
		// Update screen
		model.Draw(s)
		if wantSync {
			s.Sync()
			wantSync = false
		} else {
			s.Show()
		}

		// Poll event
		ev := s.PollEvent()

		// Process event
		switch ev.(type) {
		case *tcell.EventResize:
			wantSync = true
		}
		model.Update(s, ev)
	}
	s.Fini()
	if err := model.Save(); err != nil {
		panic(err)
	}
	os.Exit(0)
}
