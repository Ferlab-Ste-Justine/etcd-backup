package cmd

import (
	"encoding/hex"
	"io"
	"os"
	"os/exec"

	"github.com/Ferlab-Ste-Justine/etcd-backup/config"
	"github.com/Ferlab-Ste-Justine/etcd-backup/encryption"
	"github.com/Ferlab-Ste-Justine/etcd-backup/s3"

	"github.com/spf13/cobra"
)

func generateRestoreCmd(confPath *string) *cobra.Command {
	var backupTimestamp string
	var dataDir string
	var etcdutlPath string
	var etcdutlInitialClusterToken string
	var etcdutlInitinalCluster string
	var etcdutlInitialAdvertisePeerUrls string
	var etcdutlName string
	var UseEtcdutl bool

	var restoreCmd = &cobra.Command{
		Use:   "restore",
		Short: "Restore a snapshot in s3 to the local filesystem",
		Run: func(cmd *cobra.Command, args []string) {
			conf, confErr := config.GetConfig(*confPath)
			AbortOnErr("Error getting configurations: %s", confErr)

			func() {
				if conf.EncryptionKeyPath != "" {
					masterKeyHex, readErr := os.ReadFile(conf.EncryptionKeyPath)
					AbortOnErr("Error opening master key file: %s", readErr)
		
					masterKey := make([]byte, hex.DecodedLen(len(masterKeyHex)))
					_, convErr := hex.Decode(masterKey, masterKeyHex)
					AbortOnErr("Error decoding master hex format: %s", convErr)
				
					reader, keyCypher, restoreErr := s3.Restore(conf.S3Client, backupTimestamp)
					AbortOnErr("Error getting a snapshot download from s3: %s", restoreErr)
	
					decryptStr, decryptStrErr := encryption.NewDecryptStream(
						masterKey,
						keyCypher,
						reader,
						1024*1024,
					)
					AbortOnErr("Error generating a decryption stream from the s3 snapshot download: %s", decryptStrErr)
	
					file, fErr := os.OpenFile(conf.SnapshotPath, os.O_RDWR|os.O_CREATE, 0600)
					AbortOnErr("Error creating a snapshot file: %s", fErr)
	
					defer file.Close()
					_, cpyErr := io.Copy(file, decryptStr)
					AbortOnErr("Error copying the descryption stream into the snapshot file: %s", cpyErr)
	
					return
				}
	
				reader, _, restoreErr := s3.Restore(conf.S3Client, backupTimestamp)
				AbortOnErr("Error getting a snapshot download from s3: %s", restoreErr)
	
				file, fErr := os.OpenFile(conf.SnapshotPath, os.O_RDWR|os.O_CREATE, 0600)
				AbortOnErr("Error creating a snapshot file: %s", fErr)
	
				defer file.Close()
				_, cpyErr := io.Copy(file, reader)
				AbortOnErr("Error copying the snaphot download into the snapshot file: %s", cpyErr)
			}()

			if UseEtcdutl {
				defer func() {
					delErr := os.Remove(conf.SnapshotPath)
					AbortOnErr("Error deleting the transient snapshot file: %s", delErr)
				}()

				restoreCmd := exec.Command(
					etcdutlPath,
					"snapshot",
					"restore",
					conf.SnapshotPath,
					"--data-dir", dataDir,
					"--name", etcdutlName,
					"--initial-cluster", etcdutlInitinalCluster,
					"--initial-cluster-token", etcdutlInitialClusterToken,
					"--initial-advertise-peer-urls", etcdutlInitialAdvertisePeerUrls,
				)
				restoreCmd.Stdout = os.Stdout
				restoreCmd.Stderr = os.Stderr
				cmdErr := restoreCmd.Run()
				AbortOnErr("Error running command to unpack snapshot with etcdutl: %s", cmdErr)
			}
		},
	}

	restoreCmd.Flags().StringVarP(&backupTimestamp, "backup-timestamp", "t", "", "Timestamp part of the backup to restore. If empty, the latest backup will be restored")
	restoreCmd.Flags().StringVarP(&dataDir, "data-dir", "d", "", "Etcd data directory where the snapshot should be unpacked when unpacking the snapshot with etcdutl")
	restoreCmd.Flags().StringVarP(&etcdutlPath, "etcdutl-path", "e", "etcdutl", "Path to the etcdutl binary which will unpack the downloaded snapshot in the data directory.")
	restoreCmd.Flags().StringVarP(&etcdutlInitialClusterToken, "initial-cluster-token", "o", "etcd-cluster", "Value of the '--initial-cluster-token' argument passed when unpacking the snapshot with etcdutl")
	restoreCmd.Flags().StringVarP(&etcdutlInitinalCluster, "initial-cluster", "l", "default=http://localhost:2380", "Value of the '--initial-cluster' argument passed when unpacking the snapshot with etcdutl")
	restoreCmd.Flags().StringVarP(&etcdutlInitialAdvertisePeerUrls, "initial-advertise-peer-urls", "a", "http://localhost:2380", "Value of the '--initial-advertise-peer-urls' argument passed when unpacking the snapshot with etcdutl")
	restoreCmd.Flags().StringVarP(&etcdutlName, "name", "n", "default", "Value of the '--name' argument passed when unpacking the snapshot with etcdutl")
	restoreCmd.Flags().BoolVarP(&UseEtcdutl, "use-etcdutl", "u", true, "Whether to use etcdutl to unpack the snapshot in the directory specified by the '--data-dir' argument. If true, the snapshot will be deleted after unpacking.")

	return restoreCmd
}
