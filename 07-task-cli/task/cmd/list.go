/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/djcurill/task/tasks"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all todos",
	Long:  `list all todos in task database`,
	Run: func(cmd *cobra.Command, args []string) {
		tasks, err := tasks.GetTasks()
		if err != nil {
			fmt.Printf("error occurred retrieving tasks: %s", err)
		}
		if len(tasks) == 0 {
			fmt.Println("You currently have no active todos!")
			return
		}
		fmt.Printf("%-8s %-20s\n", "Task No.", "Description")
		fmt.Printf("%-8s %-20s\n", "---", "-------------------")
		for _, t := range tasks {
			fmt.Printf("%-8d %-20s\n", t.Key, t.Value)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
