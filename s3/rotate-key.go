package s3

import (
	"bytes"
	"context"
	"io/ioutil"

	"github.com/Ferlab-Ste-Justine/etcd-backup/config"

	"github.com/minio/minio-go/v7"
)

type ConvertKeyFn func([]byte) ([]byte, error)

func RotateKey(s3Conf config.S3ClientConfig, conv ConvertKeyFn) error {
	cli, cliErr := connect(s3Conf)
	if cliErr != nil {
		return cliErr
	}

	namingConv := NewNamingConvention(s3Conf.ObjectsPrefix)

	entries, listErr := ListBackups(cli, s3Conf.Bucket, namingConv)
	if listErr != nil {
		return listErr
	}

	for _, entry := range entries.Entries {
		if !entry.Encrypted {
			continue
		}

		_, backupKeyName := namingConv.GetObjectNames(entry.Timestamp)

		keyObj, keyObjErr := cli.GetObject(context.Background(), s3Conf.Bucket, backupKeyName, minio.GetObjectOptions{})
		if keyObjErr != nil {
			return keyObjErr
		}

		keyCypher, keyReadErr := ioutil.ReadAll(keyObj)
		if keyReadErr != nil {
			return keyReadErr
		}

		newKeyCypher, newKeyErr := conv(keyCypher)
		if newKeyErr != nil {
			return newKeyErr
		}

		_, keyPutErr := cli.PutObject(
			context.Background(), 
			s3Conf.Bucket,
			backupKeyName,
			bytes.NewBuffer(newKeyCypher),
			int64(len(newKeyCypher)),
			minio.PutObjectOptions{},
		)

		if keyPutErr != nil {
			return keyPutErr
		}
	}

	return nil
}