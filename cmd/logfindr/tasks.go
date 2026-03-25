package main

import (
	"fmt"

	"github.com/logsport/logfindr/internal/db"
	"github.com/spf13/cobra"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List all task IDs with log counts",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.Open(dbPath)
		if err != nil {
			return err
		}
		defer database.Close()

		tasks, err := database.ListTasks()
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks found.")
			return nil
		}

		fmt.Printf("%-30s %8s  %-20s  %-20s\n", "TASK ID", "LOGS", "FIRST SEEN", "LAST SEEN")
		fmt.Println("------------------------------------------------------------------------------------")
		for _, t := range tasks {
			fmt.Printf("%-30s %8d  %-20s  %-20s\n", t.TaskID, t.Count, t.FirstSeen, t.LastSeen)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tasksCmd)
}
