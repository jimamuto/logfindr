package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var dbPath string

var rootCmd = &cobra.Command{
	Use:   "logfindr",
	Short: "AI-queryable log store for Docker containers",
	Long:  "Logfindr — a lightweight, task-indexed log aggregator designed for coding agents and developer workflows.",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "/data/logfindr.db", "path to SQLite database")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
