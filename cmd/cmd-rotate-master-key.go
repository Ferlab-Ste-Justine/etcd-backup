package cmd

import (
	"github.com/Ferlab-Ste-Justine/etcd-backup/config"

	"github.com/spf13/cobra"
)

func generateRotateMasterKeyCmd(confPath *string) *cobra.Command {
	var rotateMasterKeyCmd = &cobra.Command{
		Use:   "rotate-master-key",
		Short: "Rotate the master key used to encrypt the encryption keys in S3",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := config.GetConfig(*confPath)
			AbortOnErr(err)

		},
	}

	return rotateMasterKeyCmd
}
