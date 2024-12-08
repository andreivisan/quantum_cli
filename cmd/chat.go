/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/andreivisan/quantum_cli/pkg/chat"
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

		p := tea.NewProgram(chat.New(),
			tea.WithAltScreen(),       // use alternate screen buffer
			tea.WithMouseCellMotion(), // enable mouse support
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
