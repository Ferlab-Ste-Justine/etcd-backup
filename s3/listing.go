package s3

import (
	"context"
	"errors"
	"time"

	minio "github.com/minio/minio-go/v7"
)

type BackupEntry struct {
	Timestamp time.Time
	Encrypted bool
	DumpFound bool
}

type BackupEntries struct {
	Entries map[time.Time]BackupEntry
	LastEntry *BackupEntry
}

func (entries *BackupEntries) findEntry(timestamp time.Time) (BackupEntry, error) {
	for _, entry := range entries.Entries {
		if entry.Timestamp == timestamp && entry.DumpFound {
			return entry, nil
		}
	}

	return BackupEntry{}, errors.New("Entry not found for given timestamp")
}

func (entries *BackupEntries) getLastEntry() (BackupEntry, error) {
	if entries.LastEntry != nil {
		return *entries.LastEntry, nil
	}

	return BackupEntry{}, errors.New("Not dump entry found")
}

func ListBackups(cli *minio.Client, bucket string, nameConv NamingConvention) (BackupEntries, error) {
	entries := BackupEntries{
		Entries: map[time.Time]BackupEntry{},
		LastEntry: nil,
	}

	objCh := cli.ListObjects(context.Background(), bucket, minio.ListObjectsOptions{})
	for object := range objCh {
		if object.Err != nil {
			return entries, object.Err
		}

		info, infoErr := nameConv.GetObjectInfo(object.Key)
		if infoErr != nil {
			continue
		}

		entry := BackupEntry{
			Timestamp: info.Timestamp,
			Encrypted: false,
			DumpFound: false,
		}

		if val, ok := entries.Entries[info.Timestamp]; ok {
			entry = val
		}

		if info.Type == OBJ_TYPE_DUMP {
			entry.DumpFound = true
		} else {
			entry.Encrypted = true
		}

		entries.Entries[info.Timestamp] = entry

		if entry.DumpFound {
			if entries.LastEntry == nil || entry.Timestamp.Equal(entries.LastEntry.Timestamp) || entry.Timestamp.After(entries.LastEntry.Timestamp) {
				entries.LastEntry = &entry
			}
		} 
	}

	return entries, nil
}