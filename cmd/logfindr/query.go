package main

import (
	"fmt"
	"time"

	"github.com/logsport/logfindr/internal/db"
	"github.com/spf13/cobra"
)

var (
	queryTask      string
	queryContainer string
	querySeverity  string
	querySince     string
	queryLimit     int
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query logs by task, container, severity, or time range",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.Open(dbPath)
		if err != nil {
			return err
		}
		defer database.Close()

		var since time.Duration
		if querySince != "" {
			since, err = time.ParseDuration(querySince)
			if err != nil {
				return fmt.Errorf("invalid --since duration: %w", err)
			}
		}

		entries, err := database.Query(db.QueryFilter{
			TaskID:    queryTask,
			Container: queryContainer,
			Severity:  querySeverity,
			Since:     since,
			Limit:     queryLimit,
		})
		if err != nil {
			return err
		}

		if len(entries) == 0 {
			fmt.Println("No logs found.")
			return nil
		}

		for _, e := range entries {
			fmt.Printf("[%s] %s | %s | task=%s | %s\n",
				e.Timestamp.Format("2006-01-02 15:04:05"),
				e.ContainerName,
				e.Severity,
				e.TaskID,
				string(e.Message),
			)
		}
		fmt.Printf("\n--- %d log(s) returned ---\n", len(entries))
		return nil
	},
}

func init() {
	queryCmd.Flags().StringVar(&queryTask, "task", "", "filter by task ID")
	queryCmd.Flags().StringVar(&queryContainer, "container", "", "filter by container name")
	queryCmd.Flags().StringVar(&querySeverity, "severity", "", "filter by severity (info, warn, error)")
	queryCmd.Flags().StringVar(&querySince, "since", "", "time range (e.g. 1h, 30m, 24h)")
	queryCmd.Flags().IntVar(&queryLimit, "limit", 100, "max results")
	rootCmd.AddCommand(queryCmd)
}
