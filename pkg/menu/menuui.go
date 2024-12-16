package menu

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	asciiArt = `
   ____                   _                    ____ _     ___ 
  / __ \                 | |                  / ___| |   |_ _|
 | |  | |_   _  __ _ _ __| |_ _   _ _ __ ___ | |   | |    | | 
 | |  | | | | |/ _' | '__| __| | | | '_ ' _ \| |   | |    | | 
 | |__| | |_| | (_| | |  | |_| |_| | | | | | | |___| |___ | | 
  \___\_\\__,_|\__,_|_|   \__|\__,_|_| |_| |_|\____|_____|___|                

`
	description = `Quantum CLI (qcli) provides an interactive terminal interface for chatting with 
a Chain of Thought Large Language Model powered by Ollama.

This CLI tool allows you to:
• Have natural conversations with a local LLM
• Leverage Chain of Thought prompting for more reasoned responses
• Use different Ollama models through configuration
• Enjoy a clean, terminal-based UI for your AI interactions

Configure your experience using environment variables or config files:
- OLLAMA_URL: URL of your Ollama instance
- OLLAMA_MODEL: The model you want to use (e.g., llama2, mistral)
`
	titleStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Bold(true)
	descriptionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99")).MarginTop(1)
	listStyle        = lipgloss.NewStyle().MarginTop(7)
)

type Model struct {
	list   list.Model
	width  int
	height int
}

type item struct {
	title, description string
}

func (listItem item) Title() string       { return listItem.title }
func (listItem item) Description() string { return listItem.description }
func (listItem item) FilterValue() string { return listItem.title }

func New() *Model {
	items := []list.Item{
		item{title: "chat", description: "chat with AI"},
		item{title: "prettyJson", description: "pretty print JSON"},
	}
	delegate := list.NewDefaultDelegate()
	cyberpunkYellow := lipgloss.Color("226")

	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(cyberpunkYellow).
		BorderLeftForeground(cyberpunkYellow).
		Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(cyberpunkYellow).
		BorderLeftForeground(cyberpunkYellow)
	menuModel := &Model{
		list: list.New(items, delegate, 0, 0),
	}
	menuModel.list.SetShowTitle(false)
	menuModel.list.SetShowStatusBar(false)
	return menuModel
}

func (menuModel Model) Init() tea.Cmd {
	return nil
}

func (menuModel Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		menuModel.width, menuModel.height = msg.Width, msg.Height
		headerContent := lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render(asciiArt),
			descriptionStyle.Render(description),
		)
		headerContentHeight := lipgloss.Height(headerContent)
		listHeight := menuModel.height - headerContentHeight
		horizontalMargin, verticalMargin := listStyle.GetFrameSize()
		menuModel.list.SetSize(
			menuModel.width-horizontalMargin,
			listHeight-verticalMargin,
		)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return menuModel, tea.Quit
		}
	}
	var cmd tea.Cmd
	menuModel.list, cmd = menuModel.list.Update(msg)
	return menuModel, cmd
}

func (menuModel Model) View() string {
	header := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(asciiArt),
		descriptionStyle.Render(description),
	)
	listView := listStyle.Render(menuModel.list.View())
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		listView,
	)
	return lipgloss.Place(
		menuModel.width,
		menuModel.height,
		lipgloss.Center, // Horizontal center
		lipgloss.Top,    // Vertical center
		content,
	)
}
