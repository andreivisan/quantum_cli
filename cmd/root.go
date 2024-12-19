/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/andreivisan/quantum_cli/pkg/menu"
	"github.com/andreivisan/quantum_cli/pkg/ollama"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var ollamaChecker *ollama.Checker

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "qcli",
	Short: "Quantum CLI - AI Assistant",
	Long: `Quantum CLI (qcli) provides an interactive terminal interface for chatting with 
an AI assistant.

This CLI tool allows you to:
• Have natural conversations with an AI
• Enjoy a clean, terminal-based UI for your AI interactions`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ollamaChecker = ollama.NewChecker("http://localhost:11434")

		// Check if Ollama is installed
		if !ollamaChecker.CheckInstallation() {
			fmt.Println("Ollama is not installed. Installing Ollama is required to use Quantum CLI.")
			if err := ollamaChecker.InstallOllama(); err != nil {
				fmt.Printf("Installation error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Please install Ollama and run it before starting Quantum CLI.")
			os.Exit(1)
		}

		// Check if server is running and offer to start it
		if !ollamaChecker.IsServerRunning() {
			fmt.Println("Ollama server is not running. Would you like to start it? (yes/no)")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response == "yes" || response == "y" {
				fmt.Println("Starting Ollama server...")
				if err := ollamaChecker.StartServer(); err != nil {
					fmt.Printf("Failed to start Ollama server: %v\n", err)
					os.Exit(1)
				}
				fmt.Println("Ollama server started successfully!")
			} else {
				fmt.Println("Ollama server is required to use Quantum CLI.")
				fmt.Println("You can start it manually by running 'ollama serve' in a separate terminal.")
				os.Exit(1)
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			p := tea.NewProgram(
				menu.New(),
				tea.WithAltScreen(),
			)

			finalModel, err := p.Run()
			if err != nil {
				fmt.Println("Error running program:", err)
				os.Exit(1)
			}

			if menuModel, ok := finalModel.(menu.Model); ok {
				if menuModel.Quitting() {
					cleanup()
					os.Exit(0)
				}
				if menuModel.Choice() == "AI chat" {
					chatCmd.Run(cmd, args)
				}
			}
		}
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Ollama server",
	Run: func(cmd *cobra.Command, args []string) {
		ollamaChecker = ollama.NewChecker("http://localhost:11434")
		if !ollamaChecker.IsServerRunning() {
			fmt.Println("Ollama server is not running.")
			return
		}

		fmt.Println("Stopping Ollama server...")
		ollamaChecker.ServerStartedByUs = true // Set to true to allow stopping
		err := ollamaChecker.StopServer()
		if err != nil {
			fmt.Printf("Failed to stop Ollama server: %v\n", err)
		} else {
			fmt.Println("Ollama server stopped successfully.")
		}
	},
}

func cleanup() {
	if ollamaChecker != nil && ollamaChecker.ServerStartedByUs {
		fmt.Println("Stopping Ollama server...")
		if err := ollamaChecker.StopServer(); err != nil {
			fmt.Printf("Error stopping server: %v\n", err)
		} else {
			fmt.Println("Ollama server stopped successfully.")
		}
	}
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
