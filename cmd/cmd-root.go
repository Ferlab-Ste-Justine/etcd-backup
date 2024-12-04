package cmd

import (
	"github.com/spf13/cobra"
)

func generateRootCmd() *cobra.Command {
	var confPath string

	var rootCmd = &cobra.Command{
		Use:   "etcd-backup",
		Short: "Manages etcd backups and restore in an s3 store",
	}

	rootCmd.PersistentFlags().StringVarP(&confPath, "config", "c", "config.yml", "Path to a yaml configuration file")
	rootCmd.MarkPersistentFlagFilename("config")

	rootCmd.AddCommand(generateBackupCmd(&confPath))
	rootCmd.AddCommand(generateRestoreCmd(&confPath))
	rootCmd.AddCommand(generatePruneCmd(&confPath))
	rootCmd.AddCommand(generateRotateKeyCmd(&confPath))

	return rootCmd
}

func Execute() error {
	return generateRootCmd().Execute()
}
