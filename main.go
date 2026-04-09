package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	inputURL state = iota
	inputStart
	inputEnd
	selectQuality
	downloading
)

var (
	choices = []string{"240", "360", "480", "720", "1080", "best"}
	cyan    = lipgloss.Color("39")
	gray    = lipgloss.Color("241")
	pink    = lipgloss.Color("205")

	asciiArt = `
  ______   __        ______  _______   _______ 
 /      \ /  |      /      |/       \ /       \
/$$$$$$  |$$ |      $$$$$$/ $$$$$$$  |$$$$$$$  |
$$ |  $$/ $$ |        $$ |  $$ |__$$ |$$ |__$$ |
$$ |      $$ |        $$ |  $$    $$/ $$    $$< 
$$ |   __ $$ |        $$ |  $$$$$$$/  $$$$$$$  |
$$ \__/  |$$ |_____  _$$ |_ $$ |      $$ |  $$ |
$$    $$/ $$       |/ $$   |$$ |      $$ |  $$ |
 $$$$$$/  $$$$$$$$/ $$$$$$/ $$/       $$/   $$/ 
`
)

type model struct {
	state      state
	urlInput   textinput.Model
	startInput textinput.Model
	endInput   textinput.Model
	cursor     int
	quality    string
	width      int
	height     int
}

func initialModel() model {
	style := lipgloss.NewStyle().Foreground(cyan)
	u := textinput.New()
	u.Placeholder = "Paste YouTube URL"
	u.TextStyle = style
	u.Focus()

	s := textinput.New()
	s.Placeholder = "00:00:10"
	s.TextStyle = style

	e := textinput.New()
	e.Placeholder = "00:00:20"
	e.TextStyle = style

	return model{state: inputURL, urlInput: u, startInput: s, endInput: e}
}

func (m model) Init() tea.Cmd { return textinput.Blink }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			switch m.state {
			case inputURL:
				if m.urlInput.Value() != "" {
					m.state = inputStart
					m.startInput.Focus()
				}
			case inputStart:
				m.state = inputEnd
				m.endInput.Focus()
			case inputEnd:
				m.state = selectQuality
			case selectQuality:
				m.quality = choices[m.cursor]
				m.state = downloading
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	switch m.state {
	case inputURL:
		m.urlInput, cmd = m.urlInput.Update(msg)
	case inputStart:
		m.startInput, cmd = m.startInput.Update(msg)
	case inputEnd:
		m.endInput, cmd = m.endInput.Update(msg)
	case selectQuality:
		if k, ok := msg.(tea.KeyMsg); ok {
			if k.String() == "up" && m.cursor > 0 { m.cursor-- }
			if k.String() == "down" && m.cursor < len(choices)-1 { m.cursor++ }
		}
	}
	return m, cmd
}

func (m model) View() string {
	if m.width == 0 { return "Initializing..." }
	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Foreground(cyan).Render(asciiArt))
	content.WriteString("\n\n")

	var active string
	switch m.state {
	case inputURL: active = fmt.Sprintf("VIDEO URL\n%s", m.urlInput.View())
	case inputStart: active = fmt.Sprintf("START TIME\n%s", m.startInput.View())
	case inputEnd: active = fmt.Sprintf("END TIME\n%s", m.endInput.View())
	case selectQuality:
		active = "QUALITY\n"
		for i, c := range choices {
			if m.cursor == i {
				active += fmt.Sprintf("➔ %s\n", lipgloss.NewStyle().Foreground(pink).Bold(true).Render(c))
			} else {
				active += fmt.Sprintf("  %s\n", c)
			}
		}
	}

	content.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Render(active))
	content.WriteString("\n\n" + lipgloss.NewStyle().Foreground(gray).Render("enter: next • q: quit"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content.String())
}

func main() {
	if err := checkDeps(); err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	m := finalModel.(model)
	if m.state == downloading {
		// Set a 10-minute timeout context for the download
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		fmt.Printf("\n🎬 %s INITIALIZING CLIP\n", lipgloss.NewStyle().Foreground(pink).Render("CLIPR:"))
		fmt.Printf("   Quality: %s | Range: %s - %s\n\n", m.quality, m.startInput.Value(), m.endInput.Value())

		err := downloadClip(ctx, m.urlInput.Value(), m.startInput.Value(), m.endInput.Value(), m.quality)
		if err != nil {
			fmt.Printf("\n❌ Download failed: %v\n", err)
		} else {
			fmt.Println("\n✅ Success! Check current folder for your clip.")
		}
	}
}
