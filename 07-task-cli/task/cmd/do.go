/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strconv"

	"github.com/djcurill/task/tasks"
	"github.com/spf13/cobra"
)

// doCmd represents the do command
var doCmd = &cobra.Command{
	Use:   "do",
	Short: "complete a task",
	Long: `completing a task using the 'do' command will delete the task
	from the db`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ids := []int{}
		for _, arg := range args {
			i, err := strconv.Atoi(arg)
			if err != nil {
				return fmt.Errorf("invalid integer %s", arg)
			}
			ids = append(ids, i)

		}
		for _, id := range ids {
			err := tasks.DoTask(id)
			if err != nil {
				return fmt.Errorf("error occurred deleting task: %s", err)
			}
			fmt.Printf("marked task #%d complete!", id)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// doCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// doCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
