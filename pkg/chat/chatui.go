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

type OutputMsg string

type Styles struct {
	BorderColor lipgloss.Color
	InputStyle  lipgloss.Style
	PromptStyle lipgloss.Style
}

func DefaultStyles() *Styles {
	styles := new(Styles)
	styles.BorderColor = lipgloss.Color("36")
	styles.InputStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.BorderColor).
		Padding(1)
	styles.PromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("99"))
	return styles
}

type model struct {
	ready            bool
	viewport         viewport.Model
	messages         []string
	textarea         textarea.Model
	width            int
	height           int
	err              error
	styles           *Styles
	userInputChan    chan<- string
	ollamaOutputChan <-chan string
}

func New(userInputChan chan<- string, ollamaOutputChan <-chan string) *model {
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
		ready:            true,
		viewport:         viewport,
		messages:         []string{},
		textarea:         textarea,
		width:            0,
		height:           0,
		err:              nil,
		styles:           styles,
		userInputChan:    userInputChan,
		ollamaOutputChan: ollamaOutputChan,
	}
}

func (model model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, listenForOllamaOutput(model.ollamaOutputChan))
}

func listenForOllamaOutput(outputChannel <-chan string) tea.Cmd {
	return func() tea.Msg {
		llamaMessage, ok := <-outputChannel
		if !ok {
			return nil
		}
		return OutputMsg(llamaMessage)
	}
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
				return model, nil
			}
			model.userInputChan <- userInput
			model.messages = append(model.messages, model.styles.PromptStyle.Render("You: ")+userInput)
			model.viewport.SetContent(strings.Join(model.messages, "\n"))
			model.textarea.Reset()
			model.viewport.GotoBottom()
		default:
			model.textarea, textareaCmd = model.textarea.Update(msg)
		}
	case error:
		model.err = msg
	case cursor.BlinkMsg:
		// Forward cursor blink messages to the textarea as well.
		model.textarea, textareaCmd = model.textarea.Update(msg)
	case OutputMsg:
		chunk := string(msg)
		if len(model.messages) == 0 {
			model.messages = append(model.messages, model.styles.PromptStyle.Render("Llama: ")+chunk)
		} else {
			model.messages[len(model.messages)-1] = model.messages[len(model.messages)-1] + " " + chunk
		}
		model.viewport.SetContent(strings.Join(model.messages, " "))
		model.textarea.Reset()
		model.viewport.GotoBottom()
		// Re-issue the listen command to wait for the next message
		return model, tea.Batch(textareaCmd, viewportCmd, listenForOllamaOutput(model.ollamaOutputChan))
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
