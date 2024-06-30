package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/ghodss/yaml"
	"github.com/google/uuid"
	"github.com/matta/sift/internal/loghelp"
	"github.com/matta/sift/internal/replicatedtodo"
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

type model interface {
	Update(screen tcell.Screen, event tcell.Event) model
	Draw(s tcell.Screen)
}

type listModel struct {
	selected map[string]struct{}
	items    replicatedtodo.ItemList
	cursor   *uuid.UUID
	quit     bool
}

func (m *listModel) addSampleItems() {
	m.newTodo("todo 1")
	m.newTodo("todo 2")
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
			// if m.persisted.Cursor > 0 {
			// 	m.persisted.Cursor--
			// }
			panic("write me")
		case event.Key() == tcell.KeyRune && event.Rune() == 'j':
			// if m.persisted.Cursor < len(m.persisted.Items)-1 {
			// 	m.persisted.Cursor++
			// }
			panic("write me")
		case event.Key() == tcell.KeyRune && event.Rune() == 'x':
			panic("write me")
			// m.persisted.Items[m.persisted.Cursor].Done = !m.persisted.Items[m.persisted.Cursor].Done
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

	// If the cursor isn't valid take a random item from the item set.
	if m.cursor == nil {
		for _, item := range m.items.Items() {
			m.cursor = &item.ID
		}
	}

	row := 0
	for _, item := range m.items.Items() {
		cursor := " "
		if item.ID == *m.cursor {
			cursor = ">"
		}

		done := " "
		// TODO: use a constant for this state value
		if item.State == "completed" {
			done = "x"
		}

		line := fmt.Sprintf("%s [%s] %s", cursor, done, item.Title)
		drawText(s, bounds{position{col: 0, row: row}, extent{width: screenExtent.width, height: 1}}, style, line)
		row += 1
	}
}

func (m *listModel) newTodo(title string) {
	var previous uuid.UUID
	if m.cursor != nil {
		previous = *m.cursor
	}
	m.items.NewTodo(title, previous)
}

func (m *listModel) Save() error {
	bytes, err := yaml.Marshal(&m.items)
	if err != nil {
		return fmt.Errorf("failed to marshal model: %w", err)
	}
	err = os.WriteFile(UserDataFile(), bytes, os.FileMode(0600))
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
			panic("write me")
			// m.list.persisted.Items = append(m.list.persisted.Items, todo{Title: m.title, Done: false})
			// return m.list
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

func NewModel() listModel {
	return listModel{
		cursor:   nil,
		selected: map[string]struct{}{},
		items:    replicatedtodo.ItemList{},
		quit:     false,
	}
}

func LoadModel() listModel {
	bytes, err := os.ReadFile(UserDataFile())
	if err != nil {
		log.Printf("Failed to read model file: %v", err)
		model := NewModel()
		model.addSampleItems()
		return model
	}

	var items replicatedtodo.ItemList
	if err = yaml.Unmarshal(bytes, &items); err != nil {
		log.Printf("Failed to unmarshal model file: %v", err)
		model := NewModel()
		model.addSampleItems()
		return model
	}

	model := NewModel()
	model.items = items
	return model
}

func setUpLogging() *os.File {
	logfilePath := os.Getenv("SIFT_LOGFILE")
	if logfilePath != "" {
		file, err := loghelp.LogToFileWith(logfilePath, "sift", log.Default())
		if err != nil {
			fmt.Fprintf(os.Stderr, "error logging to file: %s\n", err)
			os.Exit(1)
		}

		log.Default().SetFlags(log.LstdFlags | log.Lmicroseconds | log.Llongfile)

		return file
	}

	return nil
}

func main() {
	logFile := setUpLogging()
	defer func() {
		if logFile != nil {
			_ = logFile.Close()
		}
	}()
	slog.Info("program started")

	listModel := LoadModel()
	slog.Info("Loaded model", slog.Any("model", listModel))
	var model model = &listModel

	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Init(); err != nil {
		log.Fatal(err)
	}
	defer s.Fini()

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
	if err := listModel.Save(); err != nil {
		slog.Error("Error saving", slog.Any("error", err))
	}

	slog.Debug("program exiting")
}
