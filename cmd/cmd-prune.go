package cmd

import (
	"time"

	"github.com/Ferlab-Ste-Justine/etcd-backup/config"
	"github.com/Ferlab-Ste-Justine/etcd-backup/s3"

	"github.com/spf13/cobra"
)

func generatePruneCmd(confPath *string) *cobra.Command {
	var maxAge string
	var minCount int64

	var pruneCmd = &cobra.Command{
		Use:   "prune",
		Short: "Prune older backups in S3",
		Run: func(cmd *cobra.Command, args []string) {
			conf, confErr := config.GetConfig(*confPath)
			AbortOnErr("Error getting configurations: %s", confErr)

			expiry, expiryErr := time.ParseDuration(maxAge)
			AbortOnErr("Error parsing max-age argument: %s", expiryErr)

			pruneErr := s3.Prune(conf.S3Client, expiry, minCount)
			AbortOnErr("Error pruning backups: %s", pruneErr)
		},
	}

	pruneCmd.Flags().StringVarP(&maxAge, "max-age", "a", "15d", "Max age after which backups should be deleted")
	pruneCmd.Flags().Int64VarP(&minCount, "min-count", "i", 20, "Minimum number of backups to keep, regardless of the maximum age")

	return pruneCmd
}
