package menu

import (
	"github.com/charmbracelet/bubbles/key"
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
	description = `Quantum CLI (qcli) is a developer's companion providing an interactive terminal interface 
for various development tools, with AI capabilities powered by Ollama.

Currently available:
• AI Chat with Chain of Thought reasoning for more detailed responses

Coming soon:
• OCR capabilities
• Additional developer tools and AI features
`
	titleStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Bold(true)
	descriptionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99")).MarginTop(1)
	listStyle        = lipgloss.NewStyle().MarginTop(7)
)

type Model struct {
	list     list.Model
	choice   string
	quitting bool
	width    int
	height   int
}

type item struct {
	title, description string
}

func (listItem item) Title() string       { return listItem.title }
func (listItem item) Description() string { return listItem.description }
func (listItem item) FilterValue() string { return listItem.title }

func New() *Model {
	items := []list.Item{
		item{title: "AI chat", description: "chat with AI"},
		item{title: "AI OCR", description: "COMING SOON: extract text from images"},
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
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c", "esc"))):
			menuModel.quitting = true
			return menuModel, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			selectedItem, ok := menuModel.list.SelectedItem().(item)
			if ok {
				menuModel.choice = selectedItem.title
				return menuModel, tea.Quit
			}
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

func (menuModel Model) Choice() string {
	return menuModel.choice
}

func (menuModel Model) Quitting() bool {
	return menuModel.quitting
}
