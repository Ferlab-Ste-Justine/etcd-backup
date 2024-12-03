package s3

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/Ferlab-Ste-Justine/etcd-backup/config"

	"github.com/minio/minio-go/v7"
)

func Backup(source io.Reader, s3Conf config.S3ClientConfig, cypherKey []byte) error {
	cli, cliErr := connect(s3Conf)
	if cliErr != nil {
		return cliErr
	}

	namingConv := NewNamingConvention(s3Conf.ObjectsPrefix)
	backupName, backupKeyName := namingConv.GetObjectNames(time.Now())

	if len(cypherKey) > 0 {
		_, keyErr := cli.PutObject(
			context.Background(), 
			s3Conf.Bucket,
			backupKeyName,
			bytes.NewBuffer(cypherKey),
			int64(len(cypherKey)),
			minio.PutObjectOptions{},
		)

		if keyErr != nil {
			return keyErr
		}
	}

	_, backErr := cli.PutObject(
		context.Background(), 
		s3Conf.Bucket,
		backupName,
		source,
		-1,
		minio.PutObjectOptions{},
	)

	return backErr
} 