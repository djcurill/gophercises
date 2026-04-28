/*
Copyright © 2026 Daniel Curilla curilladaniel@gmail.com
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/djcurill/task/tasks"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "task",
	Short: "CLI tool for managing a task list",
	Long: `The task cli is a simple CLI tool for managing todos.

To add a todo:
	task add <todo>

To list todos:
	task list
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path := filepath.Join(homeDir, "tasks.db")
		err = tasks.InitDb(path)
		return err
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		err := tasks.CloseDb()
		if err != nil {
			fmt.Printf("an error occurred closing the db: %s\n", err)
		}
		return err
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	fmt.Println("executing cmd")
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.task.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
