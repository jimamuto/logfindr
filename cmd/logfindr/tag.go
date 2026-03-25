package main

import (
	"fmt"

	"github.com/logsport/logfindr/internal/db"
	"github.com/spf13/cobra"
)

var tagTaskID string

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Set the active task ID for incoming logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		if tagTaskID == "" {
			return fmt.Errorf("--task is required")
		}

		database, err := db.Open(dbPath)
		if err != nil {
			return err
		}
		defer database.Close()

		if err := database.SetActiveTask(tagTaskID); err != nil {
			return err
		}

		fmt.Printf("Active task set to: %s\n", tagTaskID)
		return nil
	},
}

func init() {
	tagCmd.Flags().StringVar(&tagTaskID, "task", "", "task ID to tag incoming logs with")
	rootCmd.AddCommand(tagCmd)
}
