package cmd

import (
	"context"
	"encoding/hex"
	"io"
	"os"
	"time"

	"github.com/Ferlab-Ste-Justine/etcd-backup/config"
	"github.com/Ferlab-Ste-Justine/etcd-backup/encryption"

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

			masterKeyHex, readErr := os.ReadFile(conf.EncryptionKeyPath)
			AbortOnErr(readErr)

			masterKey := make([]byte, hex.DecodedLen(len(masterKeyHex)))
			_, convErr := hex.Decode(masterKey, masterKeyHex)
			AbortOnErr(convErr)

			backupFileHandle, openErr := os.Open(conf.SnapshotPath)
			AbortOnErr(openErr)

			encrStream, encStreamErr := encryption.NewEncryptStream(masterKey, backupFileHandle, 1024*1024)
			AbortOnErr(encStreamErr)

			encCiph, encCiphErr := encrStream.GetEncryptedCipherKey()
			AbortOnErr(encCiphErr)

			//Should be to minio, but writing to file for now to validate
			dest, fileCrErr := os.Create("snapshots/backup.enc")
			AbortOnErr(fileCrErr)

			_, copyErr := io.Copy(dest, encrStream)
			AbortOnErr(copyErr)

			//Just to validating decryption
			encBackupFileHandle, encOpenErr := os.Open("snapshots/backup.enc")
			AbortOnErr(encOpenErr)

			decrStream, decrStreamErr := encryption.NewDecryptStream(masterKey, encCiph, encBackupFileHandle, 1024*1024)
			AbortOnErr(decrStreamErr)

			dest, fileCrErr = os.Create("snapshots/backup.dec")
			AbortOnErr(fileCrErr)

			_, copyErr = io.Copy(dest, decrStream)
			AbortOnErr(copyErr)
		},
	}

	return backupCmd
}
