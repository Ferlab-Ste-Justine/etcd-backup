package s3

import (
	"context"
	"time"

	"github.com/Ferlab-Ste-Justine/etcd-backup/config"

	"github.com/minio/minio-go/v7"
)

func PruneBackupEntry(cli *minio.Client, bucket string, namingConv NamingConvention, entry BackupEntry) error {
	backupName, backupKeyName := namingConv.GetObjectNames(entry.Timestamp)

	if entry.DumpFound {
		delErr := cli.RemoveObject(context.Background(), bucket, backupName, minio.RemoveObjectOptions{})
		if delErr != nil {
			return delErr
		}
	}

	if entry.Encrypted {
		delErr := cli.RemoveObject(context.Background(), bucket, backupKeyName, minio.RemoveObjectOptions{})
		if delErr != nil {
			return delErr
		}
	}

	return nil
}

func Prune(s3Conf config.S3ClientConfig, expiry time.Duration, minCount int64) error {
	cli, cliErr := connect(s3Conf)
	if cliErr != nil {
		return cliErr
	}

	namingConv := NewNamingConvention(s3Conf.ObjectsPrefix)

	entries, listErr := ListBackups(cli, s3Conf.Bucket, namingConv)
	if listErr != nil {
		return listErr
	}

	deletables := entries.GetDeletable(time.Now().Add(-expiry), minCount)

	for _, entry := range deletables {
		delErr := PruneBackupEntry(cli, s3Conf.Bucket, namingConv, entry)
		if delErr != nil {
			return delErr
		}
	}

	return nil
}