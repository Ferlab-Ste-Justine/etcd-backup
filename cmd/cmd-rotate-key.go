package cmd

import (
	"encoding/hex"
	"os"

	"github.com/Ferlab-Ste-Justine/etcd-backup/config"
	"github.com/Ferlab-Ste-Justine/etcd-backup/encryption"
	"github.com/Ferlab-Ste-Justine/etcd-backup/s3"

	"github.com/spf13/cobra"
)

func generateRotateKeyCmd(confPath *string) *cobra.Command {
	var prevKeyPath string

	var rotateKeyCmd = &cobra.Command{
		Use:   "rotate-key",
		Short: "Rotate the master key used to encrypt the encryption keys in S3",
		Run: func(cmd *cobra.Command, args []string) {
			conf, err := config.GetConfig(*confPath)
			AbortOnErr(err)

			masterKeyHex, readErr := os.ReadFile(conf.EncryptionKeyPath)
			AbortOnErr(readErr)

			masterKey := make([]byte, hex.DecodedLen(len(masterKeyHex)))
			_, convErr := hex.Decode(masterKey, masterKeyHex)
			AbortOnErr(convErr)

			prevMasterKeyHex, prevReadErr := os.ReadFile(prevKeyPath)
			AbortOnErr(prevReadErr)

			prevMasterKey := make([]byte, hex.DecodedLen(len(prevMasterKeyHex)))
			_, prevConvErr := hex.Decode(prevMasterKey, prevMasterKeyHex)
			AbortOnErr(prevConvErr)

			rotateErr := s3.RotateKey(conf.S3Client, func(keyCypher []byte) ([]byte, error) {
				keyPlaintext, decErr := encryption.DecryptBytes(keyCypher, prevMasterKey)
				if decErr != nil {
					//Try with new master key in case it was already switched
					_, decNewKeyErr := encryption.DecryptBytes(keyCypher, masterKey)
					if decNewKeyErr != nil {
						return []byte{}, decErr
					}

					//Key was already switched, probably in a previous rotation that didn't complete 
					return keyCypher, nil
				}

				return encryption.EncryptBytes(keyPlaintext, masterKey)
			})
			AbortOnErr(rotateErr)
		},
	}

	rotateKeyCmd.Flags().StringVarP(&prevKeyPath, "previous-key", "p", "", "Path to the previous master key currently encrypting the backup keys")

	return rotateKeyCmd
}
