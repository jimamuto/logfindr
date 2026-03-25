package main

import (
	"fmt"

	"github.com/logsport/logfindr/internal/db"
	"github.com/spf13/cobra"
)

var (
	compareTaskA string
	compareTaskB string
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare logs between two tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		if compareTaskA == "" || compareTaskB == "" {
			return fmt.Errorf("both --task-a and --task-b are required")
		}

		database, err := db.Open(dbPath)
		if err != nil {
			return err
		}
		defer database.Close()

		logsA, err := database.Query(db.QueryFilter{TaskID: compareTaskA, Limit: 500})
		if err != nil {
			return err
		}
		logsB, err := database.Query(db.QueryFilter{TaskID: compareTaskB, Limit: 500})
		if err != nil {
			return err
		}

		errorsA := countBySeverity(logsA, "error")
		errorsB := countBySeverity(logsB, "error")
		warnsA := countBySeverity(logsA, "warn")
		warnsB := countBySeverity(logsB, "warn")

		fmt.Printf("Task Comparison: %s vs %s\n", compareTaskA, compareTaskB)
		fmt.Println("============================================")
		fmt.Printf("%-20s %10s %10s\n", "", compareTaskA, compareTaskB)
		fmt.Printf("%-20s %10d %10d\n", "Total logs", len(logsA), len(logsB))
		fmt.Printf("%-20s %10d %10d\n", "Errors", errorsA, errorsB)
		fmt.Printf("%-20s %10d %10d\n", "Warnings", warnsA, warnsB)

		if errorsA > 0 {
			fmt.Printf("\n--- Errors in %s ---\n", compareTaskA)
			for _, e := range logsA {
				if e.Severity == "error" {
					fmt.Printf("  [%s] %s: %s\n", e.Timestamp.Format("15:04:05"), e.ContainerName, string(e.Message))
				}
			}
		}
		if errorsB > 0 {
			fmt.Printf("\n--- Errors in %s ---\n", compareTaskB)
			for _, e := range logsB {
				if e.Severity == "error" {
					fmt.Printf("  [%s] %s: %s\n", e.Timestamp.Format("15:04:05"), e.ContainerName, string(e.Message))
				}
			}
		}
		return nil
	},
}

func countBySeverity(entries []db.LogEntry, severity string) int {
	count := 0
	for _, e := range entries {
		if e.Severity == severity {
			count++
		}
	}
	return count
}

func init() {
	compareCmd.Flags().StringVar(&compareTaskA, "task-a", "", "first task ID")
	compareCmd.Flags().StringVar(&compareTaskB, "task-b", "", "second task ID")
	rootCmd.AddCommand(compareCmd)
}
