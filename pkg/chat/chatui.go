package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
}

func New() *model {
	// Initialize the chat's text area
	textArea := textarea.New()
	textArea.Placeholder = "Ask a question..."
	textArea.Focus()
	textArea.SetWidth(70)
	textArea.SetHeight(10)
	textArea.ShowLineNumbers = false
	textArea.KeyMap.InsertNewline.SetEnabled(false)
	textArea.KeyMap.InsertNewline.SetKeys("ctrl+enter", "cmd+enter")
	// Initialize the chat's viewport
	viewport := viewport.New(70, 10)
	viewport.SetContent(`Welcome to Quantum CLI Chat! Type a message and press Enter to send.`)
	// Return the initialized model
	return &model{
		viewport:    viewport,
		messages:    []string{},
		textarea:    textArea,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
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
	model.textarea, textareaCmd = model.textarea.Update(msg)
	model.viewport, viewportCmd = model.viewport.Update(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return model, tea.Quit
		case tea.KeyEnter:
			if str := strings.TrimSpace(model.textarea.Value()); str != "" {
				model.messages = append(model.messages, str)
				model.viewport.SetContent(strings.Join(model.messages, "\n"))
				model.textarea.Reset()
				model.viewport.GotoBottom()
			}
		}
	case error:
		model.err = msg
	}
	return model, tea.Batch(textareaCmd, viewportCmd)
}

func (model *model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		model.viewport.View(),
		model.textarea.View(),
	)
}
