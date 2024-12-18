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
	Short: "Quantum CLI - AI Assistant",
	Long: `Quantum CLI (qcli) provides an interactive terminal interface for chatting with 
an AI assistant.

This CLI tool allows you to:
• Have natural conversations with an AI
• Enjoy a clean, terminal-based UI for your AI interactions`,
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
				if menuModel.Choice() == "AI chat" {
					chatCmd.Run(cmd, args)
				}
			}
		}
	},
}

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

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".quantum_cli")
	}

	viper.AutomaticEnv()

	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintln(os.Stderr, "Error reading config file:", err)
		}
	}
}
