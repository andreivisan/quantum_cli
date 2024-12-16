package menu

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	asciiArt = `
   ____                   _                    ____ _     ___ 
  / __ \                 | |                  / ___| |   |_ _|
 | |  | |_   _  __ _ _ __| |_ _   _ _ __ ___ | |   | |    | | 
 | |  | | | | |/ _' | '__| __| | | | '_ ' _ \| |___| |___ | | 
 | |__| | |_| | (_| | |  | |_| |_| | | | | | |\____|_____|___|
  \___\_\\__,_|\__,_|_|   \__|\__,_|_| |_| |_|                

`
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Bold(true)
)

type Model struct {
	width  int
	height int
}

func New() *Model {
	return &Model{}
}

func (menuModel Model) Init() tea.Cmd {
	// No init command needed, just display ASCII art
	return nil
}

func (menuModel Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		menuModel.width, menuModel.height = msg.Width, msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return menuModel, tea.Quit
		}
	}
	return menuModel, nil
}

func (menuModel Model) View() string {
	// We’ll place the ASCII art in the center of the screen
	// both horizontally and vertically using lipgloss.Place.
	return lipgloss.Place(
		menuModel.width,
		menuModel.height,
		lipgloss.Center, // Horizontal center
		lipgloss.Top,    // Vertical center
		titleStyle.Render(asciiArt),
	)
}
