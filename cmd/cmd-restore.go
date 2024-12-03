package cmd

import (
	"encoding/hex"
	"io"
	"os"

	"github.com/Ferlab-Ste-Justine/etcd-backup/config"
	"github.com/Ferlab-Ste-Justine/etcd-backup/encryption"
	"github.com/Ferlab-Ste-Justine/etcd-backup/s3"

	"github.com/spf13/cobra"
)

func generateRestoreCmd(confPath *string) *cobra.Command {
	var backupTimestamp string

	var restoreCmd = &cobra.Command{
		Use:   "restore",
		Short: "Restore a snapshot in s3 to the local filesystem",
		Run: func(cmd *cobra.Command, args []string) {
			conf, err := config.GetConfig(*confPath)
			AbortOnErr(err)

			if conf.EncryptionKeyPath != "" {
				masterKeyHex, readErr := os.ReadFile(conf.EncryptionKeyPath)
				AbortOnErr(readErr)
	
				masterKey := make([]byte, hex.DecodedLen(len(masterKeyHex)))
				_, convErr := hex.Decode(masterKey, masterKeyHex)
				AbortOnErr(convErr)
			
				reader, key, restoreErr := s3.Restore(conf.S3Client, backupTimestamp)
				AbortOnErr(restoreErr)

				decryptStr, decryptStrErr := encryption.NewDecryptStream(
					masterKey,
					key,
					reader,
					1024*1024,
				)
				AbortOnErr(decryptStrErr)

				file, fErr := os.OpenFile(conf.SnapshotPath, os.O_RDWR|os.O_CREATE, 0600)
				AbortOnErr(fErr)

				defer file.Close()
				_, cpyErr := io.Copy(file, decryptStr)
				AbortOnErr(cpyErr)

				return
			}

			reader, _, restoreErr := s3.Restore(conf.S3Client, backupTimestamp)
			AbortOnErr(restoreErr)

			file, fErr := os.OpenFile(conf.SnapshotPath, os.O_RDWR|os.O_CREATE, 0600)
			AbortOnErr(fErr)

			defer file.Close()
			_, cpyErr := io.Copy(file, reader)
			AbortOnErr(cpyErr)
		},
	}

	restoreCmd.Flags().StringVarP(&backupTimestamp, "backup-timestamp", "t", "", "Timestamp part of the backup to restore. If empty, the latest backup will be restored")

	return restoreCmd
}
