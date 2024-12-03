package cmd

import (
	"github.com/Ferlab-Ste-Justine/etcd-backup/config"

	"github.com/spf13/cobra"
)

func generatePruneCmd(confPath *string) *cobra.Command {
	var pruneCmd = &cobra.Command{
		Use:   "prune",
		Short: "Prune older backups in S3",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := config.GetConfig(*confPath)
			AbortOnErr(err)

		},
	}

	return pruneCmd
}
