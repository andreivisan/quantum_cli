/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/andreivisan/quantum_cli/pkg/chat"
	"github.com/andreivisan/quantum_cli/pkg/llama"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Chat with Ollama",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ollamaUrl := viper.GetString("OLLAMA_URL")
		ollamaModel := viper.GetString("OLLAMA_MODEL")
		fmt.Printf("Connecting to Ollama at %s using model %s\n", ollamaUrl, ollamaModel)
		userInputChan := make(chan string)
		ollamaOutputChan := make(chan string)

		p := tea.NewProgram(chat.New(userInputChan, ollamaOutputChan),
			tea.WithAltScreen(),       // use alternate screen buffer
			tea.WithMouseCellMotion(), // enable mouse support
		)

		go func() {
			for userInput := range userInputChan {
				ollamaClient := llama.NewClient(ollamaUrl, ollamaModel)
				ollamaResponse, err := ollamaClient.Chat(userInput, 1000)
				if err != nil {
					fmt.Println("Error chatting with Ollama:", err)
					continue
				}
				ollamaOutputChan <- ollamaResponse
			}
		}()

		if _, err := p.Run(); err != nil {
			fmt.Println("Error running program:", err)
			return
		}
		close(ollamaOutputChan)
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
}
