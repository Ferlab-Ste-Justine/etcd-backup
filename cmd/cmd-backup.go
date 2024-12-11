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
			conf, confErr := config.GetConfig(*confPath)
			AbortOnErr("Error getting configurations: %s", confErr)

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
			AbortOnErr("Error connecting to etcd: %s", cliErr)

			duration, _ := time.ParseDuration("1h")
			snapshotErr := cli.Snapshot(true, conf.SnapshotPath, duration)
			AbortOnErr("Error generating a snapshot file from etcd: %s", snapshotErr)

			backupFileHandle, openErr := os.Open(conf.SnapshotPath)
			AbortOnErr("Error opening the generated snapshot file: %s", openErr)

			if conf.EncryptionKeyPath != "" {
				masterKeyHex, readErr := os.ReadFile(conf.EncryptionKeyPath)
				AbortOnErr("Error opening master key file: %s", readErr)
	
				masterKey := make([]byte, hex.DecodedLen(len(masterKeyHex)))
				_, convErr := hex.Decode(masterKey, masterKeyHex)
				AbortOnErr("Error decoding master key hex format: %s", convErr)
	
				encrStream, encStreamErr := encryption.NewEncryptStream(masterKey, backupFileHandle, 1024*1024)
				AbortOnErr("Error generating an encryption stream from master key and snapshot file: %s", encStreamErr)
	
				encCiph, encCiphErr := encrStream.GetEncryptedCipherKey()
				AbortOnErr("Error generating an encryption key cypher: %s", encCiphErr)

				backupErr := s3.Backup(encrStream, conf.S3Client, encCiph)
				AbortOnErr("Error storing encrypted snapshot in s3: %s", backupErr)

				delErr := os.Remove(conf.SnapshotPath)
				AbortOnErr("Error deleting the transient snapshot file: %s", delErr)

				return
			}

			backupErr := s3.Backup(backupFileHandle, conf.S3Client, []byte{})
			AbortOnErr("Error storing snapshot in s3: %s", backupErr)

			delErr := os.Remove(conf.SnapshotPath)
			AbortOnErr("Error deleting the transient snapshot file: %s", delErr)
		},
	}

	return backupCmd
}
