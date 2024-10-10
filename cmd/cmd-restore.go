package cmd

import (
	"github.com/Ferlab-Ste-Justine/etcd-backup/config"

	"github.com/spf13/cobra"
)

func generateRestoreCmd(confPath *string) *cobra.Command {
	var restoreCmd = &cobra.Command{
		Use:   "restore",
		Short: "Restore a snapshot in s3 to the local filesystem",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := config.GetConfig(*confPath)
			AbortOnErr(err)

		},
	}

	return restoreCmd
}
