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
		defer close(userInputChan)
		ollamaOutputChan := make(chan string)
		defer close(ollamaOutputChan)

		p := tea.NewProgram(
			chat.New(userInputChan, ollamaOutputChan),
			tea.WithAltScreen(),
		)

		go func() {
			ollamaClient := llama.NewClient(ollamaUrl, ollamaModel)
			for userInput := range userInputChan {
				err := ollamaClient.Chat(userInput, 300, ollamaOutputChan)
				if err != nil {
					fmt.Println("Error chatting with Ollama:", err)
					continue
				}
			}
		}()

		if _, err := p.Run(); err != nil {
			fmt.Println("Error running program:", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
}
