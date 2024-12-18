/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/andreivisan/quantum_cli/pkg/ai"
	"github.com/andreivisan/quantum_cli/pkg/chat"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an AI chat session",
	Long: `Start an interactive chat session with the AI assistant.

The chat interface provides:
• A clean terminal UI
• Real-time streaming responses
• Support for multi-line input
• Clear separation between user and AI messages

Usage:
  qcli chat

Press Ctrl+C to exit the chat session.`,
	Run: func(cmd *cobra.Command, args []string) {
		userInputChan := make(chan string)
		aiOutputChan := make(chan string)

		// Start goroutine to handle communication with Python server
		go func() {
			defer close(aiOutputChan)
			client := ai.NewClient("http://localhost:8000")

			for message := range userInputChan {
				if err := client.Chat(message, aiOutputChan); err != nil {
					fmt.Printf("Error communicating with AI server: %v\n", err)
					continue
				}
			}
		}()

		p := tea.NewProgram(
			chat.New(userInputChan, aiOutputChan),
			tea.WithAltScreen(),
		)

		if _, err := p.Run(); err != nil {
			fmt.Println("Error running program:", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
}
