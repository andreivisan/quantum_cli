/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/andreivisan/quantum_cli/pkg/menu"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "qcli",
	Short: "Quantum CLI - Local AI Assistant using Ollama supercharged with thinking",
	Long: `Quantum CLI (qcli) provides an interactive terminal interface for chatting with 
a Chain of Thought Large Language Model powered by Ollama.

This CLI tool allows you to:
• Have natural conversations with a local LLM
• Leverage Chain of Thought prompting for more reasoned responses
• Use different Ollama models through configuration
• Enjoy a clean, terminal-based UI for your AI interactions

Configure your experience using environment variables or config files:
- OLLAMA_URL: URL of your Ollama instance
- OLLAMA_MODEL: The model you want to use (e.g., llama2, mistral)`,
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
					os.Exit(0)
				}
				if menuModel.Choice() == "chat" {
					chatCmd.Run(cmd, args)
				}
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.quantum_cli.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".quantum_cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".quantum_cli")
	}

	// Add this to read .env file from the current directory
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("QCLI")

	// Additionally read in .env file
	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintln(os.Stderr, "Error reading .env file:", err)
		}
	}
}
