package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	BorderColor lipgloss.Color
	InputStyle  lipgloss.Style
}

func DefaultStyles() *Styles {
	styles := new(Styles)
	styles.BorderColor = lipgloss.Color("36")
	styles.InputStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.BorderColor).
		Padding(1)
	return styles
}

type model struct {
	viewport viewport.Model
	messages []string
	textarea textarea.Model
	err      error
	width    int
	height   int
	styles   *Styles
}

func New() *model {
	styles := DefaultStyles()
	// Initialize the chat's text area
	textArea := textarea.New()
	textArea.Placeholder = "Ask a question..."
	textArea.Focus()
	textArea.ShowLineNumbers = false
	// Customize key mappings
	textArea.KeyMap.InsertNewline.SetEnabled(false)
	textArea.KeyMap.InsertNewline.SetKeys("ctrl+enter", "cmd+enter")
	welcomeMessage := "Welcome to Quantum CLI Chat! Type a message and press Enter to send."
	viewport := viewport.New(80, 18)
	viewport.SetContent(welcomeMessage)
	viewport.YPosition = 0
	return &model{
		viewport: viewport,
		messages: []string{welcomeMessage},
		textarea: textArea,
		styles:   styles,
		err:      nil,
		width:    80, // Default width
		height:   24, // Default height
	}
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (model *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		// Commands for textarea updates
		textareaCmd tea.Cmd
		// Commands for viewport updates
		viewportCmd tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return model, tea.Quit
		case tea.KeyEnter:
			// Capture the submitted text
			str := strings.TrimSpace(model.textarea.Value())
			if str != "" {
				userMessage := fmt.Sprintf("You: %s", str)
				model.messages = append(model.messages, userMessage)
				model.viewport.SetContent(strings.Join(model.messages, "\n"))
				model.textarea.Reset()
				model.viewport.GotoBottom()
			}
		}
	case error:
		model.err = msg
	case tea.WindowSizeMsg:
		model.height = msg.Height
		model.width = msg.Width
		model.viewport.Width = msg.Width
		model.viewport.Height = msg.Height - 6 // Leave space for textarea
		model.styles.InputStyle = model.styles.InputStyle.Width(msg.Width - 2)
		model.viewport.SetContent(strings.Join(model.messages, "\n"))
	}

	model.textarea, textareaCmd = model.textarea.Update(msg)
	model.viewport, viewportCmd = model.viewport.Update(msg)

	return model, tea.Batch(textareaCmd, viewportCmd)
}

func (model *model) View() string {
	return lipgloss.Place(
		model.width,
		model.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Center,
			model.viewport.View(),
			model.styles.InputStyle.Render(model.textarea.View()),
		),
	)
}
