package s3

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"time"

	"github.com/Ferlab-Ste-Justine/etcd-backup/config"

	"github.com/minio/minio-go/v7"
)

func Restore(s3Conf config.S3ClientConfig, timestamp string) (io.Reader, []byte, error) {
	cli, cliErr := connect(s3Conf)
	if cliErr != nil {
		return nil, []byte{}, cliErr
	}

	namingconv := NewNamingConvention(s3Conf.ObjectsPrefix)

	entries, listErr := ListBackups(cli, s3Conf.Bucket, namingconv)
	if listErr != nil {
		return nil, []byte{}, listErr
	}

	key := []byte{}
	var entry BackupEntry

	if timestamp == "" {
		if entries.LastEntry == nil || (!entries.LastEntry.DumpFound) {
			return nil, key, errors.New("No valid backups to restore")
		}

		entry = *entries.LastEntry
	} else {
		timestampTime, parseErr := time.Parse(time.RFC3339, timestamp)
		if parseErr != nil {
			return nil, key, parseErr
		}

		var ok bool
		entry, ok = entries.Entries[timestampTime]
		if !ok {
			return nil, key, errors.New("No valid with given timestamp to restore")
		}
	}

	dumpKey, keyKey := namingconv.GetObjectNames(entry.Timestamp)
	
	if entry.Encrypted {
		keyObj, keyObjErr := cli.GetObject(context.Background(), s3Conf.Bucket, keyKey, minio.GetObjectOptions{})
		if keyObjErr != nil {
			return nil, key, keyObjErr
		}
		
		var keyErr error
		key, keyErr = ioutil.ReadAll(keyObj)
		if keyErr != nil {
			return nil, key, keyErr
		}
	}

	dumpObj, dumpErr := cli.GetObject(context.Background(), s3Conf.Bucket, dumpKey, minio.GetObjectOptions{})
	if dumpErr != nil {
		return nil, key, dumpErr
	}

	return dumpObj, key, nil
} 