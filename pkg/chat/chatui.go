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

type OutputDoneMsg struct{}

type Message struct {
    Role    string
    Content string
}

type Styles struct {
	BorderColor lipgloss.Color
	InputStyle  lipgloss.Style
	PromptStyle lipgloss.Style
}

func DefaultStyles() *Styles {
	styles := new(Styles)
	styles.BorderColor = lipgloss.Color("240")
	styles.InputStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("36")).
		Padding(1).
        MarginTop(0).
        Height(6)
	styles.PromptStyle = lipgloss.NewStyle().
        PaddingLeft(1).
        PaddingRight(1).
		Foreground(lipgloss.Color("99"))
	return styles
}

type model struct {
	viewport         viewport.Model
	textarea         textarea.Model
	userInputChan    chan<- string
	ollamaOutputChan <-chan string
	messages         []Message
	err              error
	styles           *Styles
	ready            bool
	width            int
	height           int
}

func New(userInputChan chan<- string, ollamaOutputChan <-chan string) *model {
	textarea := textarea.New()
	textarea.Placeholder = "Send a message..."
	textarea.Focus()
	textarea.CharLimit = 0
	textarea.ShowLineNumbers = true
	textarea.SetHeight(4)
	textarea.MaxHeight = 100

	// Configure Shift+Enter for new lines
	newlineKey := textarea.KeyMap.InsertNewline
	newlineKey.SetKeys("alt+enter")
	textarea.KeyMap.InsertNewline.SetEnabled(true)
	textarea.KeyMap.InsertNewline = newlineKey

	viewport := viewport.New(0, 0)
	viewport.SetContent(`Type a message and press Enter to send.`)
	viewport.MouseWheelEnabled = false
	viewport.YPosition = 0

	styles := DefaultStyles()

	return &model{
		textarea:         textarea,
		viewport:         viewport,
		userInputChan:    userInputChan,
		ollamaOutputChan: ollamaOutputChan,
		messages:         []Message{},
		err:              nil,
		styles:           styles,
		ready:            true,
		width:            0,
		height:           0,
	}
}

func (model model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, listenForOllamaOutput(model.ollamaOutputChan))
}

func listenForOllamaOutput(outputChannel <-chan string) tea.Cmd {
	return func() tea.Msg {
		llamaMessage, ok := <-outputChannel
		if !ok {
			return OutputMsg(llamaMessage)
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
            headerHeight := 2
            textInputHeight := 8
            viewportHeight := msg.Height - headerHeight - textInputHeight
            model.viewport = viewport.New(msg.Width, viewportHeight)
			model.viewport.HighPerformanceRendering = true
			model.viewport.SetContent(`Type a message and press Enter to send.`)
			model.ready = true
		}
        model.width = msg.Width
        model.height = msg.Height
        model.viewport.Width = msg.Width
        model.viewport.Height = msg.Height - 10
        model.textarea.SetWidth(msg.Width - 4)
        model.textarea.SetHeight(4)
        model.styles.InputStyle = model.styles.InputStyle.Width(model.width - 2)
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
            newMsg := Message{
                Role:    "You",
                Content: userInput,
            }
			model.messages = append(model.messages, newMsg)
			//model.viewport.SetContent(strings.Join(model.messages, "\n"))
			model.rebuildViewport()
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
		//if len(model.messages) == 0 {
		//	model.messages = append(model.messages, model.styles.PromptStyle.Render("Llama: ")+chunk)
		//} else {
		//	model.messages[len(model.messages)-1] += "" + chunk
		//}
		//model.viewport.SetContent(strings.Join(model.messages, "\n"))
		if len(model.messages) == 0 || model.messages[len(model.messages)-1].Role != "AI" {
            newMsg := Message{
                Role:    "AI",
                Content: chunk,
            }
            model.messages = append(model.messages, newMsg)
        } else {
            model.messages[len(model.messages)-1].Content += chunk
        }
        model.rebuildViewport()
        model.viewport.GotoBottom()
		// Re-issue the listen command to wait for the next message
		return model, tea.Batch(textareaCmd, viewportCmd, listenForOllamaOutput(model.ollamaOutputChan))
	case OutputDoneMsg:
		// Llama finished responding, add a new line
		//model.messages = append(model.messages, "")
		//model.viewport.SetContent(strings.Join(model.messages, "\n"))
		model.viewport.GotoBottom()
		// No need to re-listen here since response ended
		return model, tea.Batch(textareaCmd, viewportCmd)
	}
	model.viewport, viewportCmd = model.viewport.Update(msg)
	return model, tea.Batch(textareaCmd, viewportCmd)
}

func (model *model) View() string {
	if !model.ready {
		return "\n Initializing..."
	}

	//return lipgloss.PlaceHorizontal(
	//	model.width,
	//	lipgloss.Center,
	//	lipgloss.JoinVertical(
	//		lipgloss.Left,
	//		model.viewport.View(),
	//		model.styles.InputStyle.Render(model.textarea.View()),
	//	),
	//)

    return lipgloss.JoinVertical(
        lipgloss.Left,
        lipgloss.NewStyle().
            BorderStyle(lipgloss.RoundedBorder()).
            BorderForeground(model.styles.BorderColor).
            Width(model.width - 2).
            Render(model.viewport.View()),
        model.styles.InputStyle.Render(model.textarea.View()),
    )
}

func (chatModel *model) formatMessage(msg Message) string {
    return fmt.Sprintf("\n%s\n%s\n",
        chatModel.styles.PromptStyle.Render(msg.Role+":"),
        chatModel.styles.InputStyle.Copy().
            UnsetBorderStyle().
            PaddingLeft(2).
            Render(msg.Content))
}

func (chatModel *model) rebuildViewport() {
    var strBuilder strings.Builder
    for _, msg := range chatModel.messages {
        strBuilder.WriteString(chatModel.formatMessage(msg))
    }
    chatModel.viewport.SetContent(strBuilder.String())
}
