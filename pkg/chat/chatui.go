package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
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
	ChatStyle   lipgloss.Style
}

func DefaultStyles() *Styles {
	styles := new(Styles)
	styles.BorderColor = lipgloss.Color("240")
	styles.InputStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("36")).
		MarginTop(0).
		Height(6)
	styles.PromptStyle = lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1).
		PaddingTop(0).
		PaddingBottom(0).
		Foreground(lipgloss.Color("99"))
	styles.ChatStyle = lipgloss.NewStyle().
		Height(2).
		PaddingLeft(2).
		PaddingTop(0).
		PaddingBottom(0).
		MarginTop(0).
		MarginBottom(1)
	return styles
}

type Model struct {
	viewport         viewport.Model
	textarea         textarea.Model
	userInputChan    chan<- string
	ollamaOutputChan <-chan string
	messages         []Message
	err              error
	styles           *Styles
	waiting          bool
	mySpinner        spinner.Model
	ready            bool
	width            int
	height           int
	renderer         *glamour.TermRenderer
	quitting         bool
}

func New(userInputChan chan<- string, ollamaOutputChan <-chan string) *Model {
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
	viewport.YPosition = 0

	mySpinner := spinner.New()
	mySpinner.Spinner = spinner.Dot
	mySpinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("36"))

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(viewport.Width),
	)

	styles := DefaultStyles()

	return &Model{
		textarea:         textarea,
		viewport:         viewport,
		userInputChan:    userInputChan,
		ollamaOutputChan: ollamaOutputChan,
		messages:         []Message{},
		err:              nil,
		styles:           styles,
		waiting:          false,
		mySpinner:        mySpinner,
		ready:            true,
		width:            0,
		height:           0,
		renderer:         renderer,
		quitting:         false,
	}
}

func (model Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, listenForOllamaOutput(model.ollamaOutputChan))
}

func (myModel *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := 1
		inputTextHeight := 6 // textarea height + margins
		viewportHeight := msg.Height - headerHeight - inputTextHeight - 3
		if !myModel.ready {
			myModel.viewport = viewport.New(msg.Width, viewportHeight)
			myModel.viewport.HighPerformanceRendering = true
			myModel.viewport.MouseWheelEnabled = true
			myModel.viewport.SetContent(`Type a message and press Enter to send.`)
			myModel.ready = true
		}
		myModel.width = msg.Width
		myModel.height = msg.Height
		myModel.viewport.Width = myModel.width - 4
		myModel.viewport.Height = viewportHeight
		myModel.textarea.SetWidth(myModel.width - 2)
		myModel.textarea.SetHeight(4)

	case tea.KeyMsg:
		if myModel.waiting {
			// Ignore most key presses while waiting
			switch msg.String() {
			case "ctrl+c":
				myModel.quitting = true
				return myModel, tea.Quit
			default:
				return myModel, nil
			}
		}
		switch msg.String() {
		case "esc", "ctrl+c":
			myModel.quitting = true
			fmt.Println(myModel.textarea.Value())
			return myModel, tea.Quit
		case "enter":
			userInput := myModel.textarea.Value()
			if userInput == "" {
				return myModel, nil
			}
			myModel.userInputChan <- userInput
			newMsg := Message{
				Role:    "You",
				Content: userInput,
			}
			myModel.messages = append(myModel.messages, newMsg)
			myModel.rebuildViewport()
			myModel.textarea.Reset()
			myModel.viewport.GotoBottom()
			myModel.waiting = true
			myModel.textarea.Blur()
			cmds = append(cmds, myModel.mySpinner.Tick, listenForOllamaOutput(myModel.ollamaOutputChan))
			return myModel, tea.Batch(cmds...)
		}

	case error:
		myModel.err = msg

	case cursor.BlinkMsg:
		var cmd tea.Cmd
		myModel.textarea, cmd = myModel.textarea.Update(msg)
		cmds = append(cmds, cmd)

	case OutputMsg:
		chunk := string(msg)
		if chunk == "Thinking...\n" {
			if !myModel.waiting {
				myModel.waiting = true
				myModel.textarea.Blur()
				myModel.rebuildViewport()
				return myModel, myModel.mySpinner.Tick
			}
			return myModel, nil
		}
		myModel.waiting = false
		myModel.textarea.Focus()
		if len(myModel.messages) == 0 || myModel.messages[len(myModel.messages)-1].Role != "AI" {
			newMsg := Message{
				Role:    "AI",
				Content: chunk,
			}
			myModel.messages = append(myModel.messages, newMsg)
		} else {
			myModel.messages[len(myModel.messages)-1].Content += chunk
		}
		myModel.rebuildViewport()
		return myModel, listenForOllamaOutput(myModel.ollamaOutputChan)

	case spinner.TickMsg:
		var cmd tea.Cmd
		myModel.mySpinner, cmd = myModel.mySpinner.Update(msg)
		if myModel.waiting {
			myModel.rebuildViewport()
			cmds = append(cmds, cmd)
		}
	}

	if myModel.waiting {
		var cmd tea.Cmd
		myModel.mySpinner, cmd = myModel.mySpinner.Update(msg)
		myModel.rebuildViewport()
		cmds = append(cmds, cmd)
	} else {
		var cmd tea.Cmd
		myModel.textarea, cmd = myModel.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}

	var cmd tea.Cmd
	myModel.viewport, cmd = myModel.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return myModel, tea.Batch(cmds...)
}

func (myModel *Model) View() string {
	if !myModel.ready {
		return "\n Initializing..."
	}

	var textareaView string
	if myModel.waiting {
		textareaView = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Render(myModel.textarea.View())
	} else {
		textareaView = myModel.styles.InputStyle.Render(myModel.textarea.View())
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(myModel.styles.BorderColor).
			Width(myModel.width-2).
			Render(myModel.viewport.View()),
		textareaView,
	)
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

func (chatModel *Model) formatMessage(msg Message) string {
	// For AI messages, render with glamour
	if msg.Role == "AI" {
		renderedMessage, _ := chatModel.renderer.Render(msg.Content)
		return fmt.Sprintf("%s%s\n",
			chatModel.styles.PromptStyle.Render(msg.Role+":"),
			chatModel.styles.ChatStyle.Render(renderedMessage))
	}

	// For user messages, keep the original formatting
	return fmt.Sprintf("%s%s\n",
		chatModel.styles.PromptStyle.Render(msg.Role+":"),
		chatModel.styles.ChatStyle.Render(msg.Content))
}

func (chatModel *Model) rebuildViewport() {
	var strBuilder strings.Builder
	for _, msg := range chatModel.messages {
		strBuilder.WriteString(chatModel.formatMessage(msg))
	}
	if chatModel.waiting {
		strBuilder.WriteString(fmt.Sprintf("%s Thinking...", chatModel.mySpinner.View()))
	}
	chatModel.viewport.SetContent(strBuilder.String())
	chatModel.viewport.GotoBottom()
}

func (myModel *Model) Quitting() bool {
	return myModel.quitting
}
