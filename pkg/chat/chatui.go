package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
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
	ready    bool
	viewport viewport.Model
	messages []string
	textarea textarea.Model
	width    int
	height   int
	err      error
	styles   *Styles
}

func New() *model {
	textarea := textarea.New()
	textarea.Placeholder = "Send a message..."
	textarea.Focus()
	textarea.CharLimit = 280
	// Remove cursor line styling
	textarea.FocusedStyle.CursorLine = lipgloss.NewStyle()
	viewport := viewport.New(0, 0)
	viewport.SetContent(`Type a message and press Enter to send.`)
	textarea.KeyMap.InsertNewline.SetEnabled(false)
	styles := DefaultStyles()
	return &model{
		ready:    true,
		viewport: viewport,
		messages: []string{},
		textarea: textarea,
		width:    0,
		height:   0,
		err:      nil,
		styles:   styles,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (model *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		textareaCmd tea.Cmd
		viewportCmd tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !model.ready {
			model.viewport = viewport.New(msg.Width, msg.Height)
			model.viewport.HighPerformanceRendering = false
			model.viewport.SetContent(`Type a message and press Enter to send.`)
			model.ready = true
		} else {
			model.width = msg.Width
			model.height = msg.Height
			textareaHeight := 10
			model.viewport.Width = model.width
			model.viewport.Height = model.height - textareaHeight
			model.textarea.SetWidth(model.width - 4)
			model.styles.InputStyle = model.styles.InputStyle.Width(model.width - 2)
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			// Quit.
			fmt.Println(model.textarea.Value())
			return model, tea.Quit
		case "enter":
			userInput := model.textarea.Value()
			if userInput == "" {
				// Don't send empty messages.
				return model, nil
			}
			// Simulate sending a message. In your application you'll want to
			// also return a custom command to send the message off to
			// a server.
			model.messages = append(model.messages, userInput)
			model.viewport.SetContent(strings.Join(model.messages, "\n"))
			model.textarea.Reset()
		default:
			model.textarea, textareaCmd = model.textarea.Update(msg)
		}
	case error:
		model.err = msg
	case cursor.BlinkMsg:
		// Forward cursor blink messages to the textarea as well.
		model.textarea, textareaCmd = model.textarea.Update(msg)
	}
	model.viewport, viewportCmd = model.viewport.Update(msg)
	return model, tea.Batch(textareaCmd, viewportCmd)
}

func (model *model) View() string {
	if !model.ready {
		return "\n Initializing..."
	}
	return lipgloss.PlaceHorizontal(
		model.width,
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Left,
			model.viewport.View(),
			model.styles.InputStyle.Render(model.textarea.View()),
		),
	)
}
