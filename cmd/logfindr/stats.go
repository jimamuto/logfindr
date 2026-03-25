package main

import (
	"fmt"

	"github.com/logsport/logfindr/internal/db"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show database statistics and compression ratio",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.Open(dbPath)
		if err != nil {
			return err
		}
		defer database.Close()

		s, err := database.Stats(dbPath)
		if err != nil {
			return err
		}

		fmt.Println("Logfindr Statistics")
		fmt.Println("====================")
		fmt.Printf("Total logs:        %d\n", s.TotalLogs)
		fmt.Printf("Total tasks:       %d\n", s.TotalTasks)
		fmt.Printf("DB file size:      %s\n", humanBytes(s.DBSizeBytes))
		fmt.Printf("Raw log data:      %s\n", humanBytes(s.TotalRawBytes))
		fmt.Printf("Stored (Zstd):     %s\n", humanBytes(s.TotalStoredBytes))
		fmt.Printf("Compression ratio: %.1fx\n", s.CompressionRatio)
		return nil
	},
}

func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
