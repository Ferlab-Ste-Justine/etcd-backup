package cmd

import (
	"context"
	"encoding/hex"
	"os"
	"time"

	"github.com/Ferlab-Ste-Justine/etcd-backup/config"
	"github.com/Ferlab-Ste-Justine/etcd-backup/encryption"
	"github.com/Ferlab-Ste-Justine/etcd-backup/s3"

	"github.com/Ferlab-Ste-Justine/etcd-sdk/client"
	"github.com/spf13/cobra"
)

func generateBackupCmd(confPath *string) *cobra.Command {
	var backupCmd = &cobra.Command{
		Use:   "backup",
		Short: "Create a snapshot in s3",
		Run: func(cmd *cobra.Command, args []string) {
			conf, err := config.GetConfig(*confPath)
			AbortOnErr(err)

			cli, cliErr := client.Connect(context.Background(), client.EtcdClientOptions{
				ClientCertPath:    conf.EtcdClient.Auth.ClientCert,
				ClientKeyPath:     conf.EtcdClient.Auth.ClientKey,
				CaCertPath:        conf.EtcdClient.Auth.CaCert,
				Username:          conf.EtcdClient.Auth.Username,
				Password:          conf.EtcdClient.Auth.Password,
				EtcdEndpoints:     conf.EtcdClient.Endpoints,
				ConnectionTimeout: conf.EtcdClient.ConnectionTimeout,
				RequestTimeout:    conf.EtcdClient.RequestTimeout,
				Retries:           conf.EtcdClient.Retries,
			})
			AbortOnErr(cliErr)

			duration, _ := time.ParseDuration("1h")
			snapshotErr := cli.Snapshot(true, conf.SnapshotPath, duration)
			AbortOnErr(snapshotErr)

			backupFileHandle, openErr := os.Open(conf.SnapshotPath)
			AbortOnErr(openErr)

			if conf.EncryptionKeyPath != "" {
				masterKeyHex, readErr := os.ReadFile(conf.EncryptionKeyPath)
				AbortOnErr(readErr)
	
				masterKey := make([]byte, hex.DecodedLen(len(masterKeyHex)))
				_, convErr := hex.Decode(masterKey, masterKeyHex)
				AbortOnErr(convErr)
	
				encrStream, encStreamErr := encryption.NewEncryptStream(masterKey, backupFileHandle, 1024*1024)
				AbortOnErr(encStreamErr)
	
				encCiph, encCiphErr := encrStream.GetEncryptedCipherKey()
				AbortOnErr(encCiphErr)

				backupErr := s3.Backup(encrStream, conf.S3Client, encCiph)
				AbortOnErr(backupErr)

				AbortOnErr(os.Remove(conf.SnapshotPath))

				return
			}

			backupErr := s3.Backup(backupFileHandle, conf.S3Client, []byte{})
			AbortOnErr(backupErr)

			AbortOnErr(os.Remove(conf.SnapshotPath))
		},
	}

	return backupCmd
}
