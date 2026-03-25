package main

import (
	"fmt"

	"github.com/logsport/logfindr/internal/db"
	"github.com/logsport/logfindr/internal/ingest"
	"github.com/spf13/cobra"
)

var serveAddr string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the log ingest server",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open db: %w", err)
		}
		defer database.Close()

		srv := ingest.New(database, serveAddr)
		return srv.Start()
	},
}

func init() {
	serveCmd.Flags().StringVar(&serveAddr, "addr", ":8080", "listen address")
	rootCmd.AddCommand(serveCmd)
}
