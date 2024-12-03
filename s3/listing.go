package s3

import (
	"context"
	"errors"
    "slices"
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

func (entries *BackupEntries) countValidBackups() int64 {
    count := 0

    for _, entry := range entries.Entries {
        if entry.DumpFound {
            count += 1
        }
    }

    return int64(count)
}

func (entries *BackupEntries) GetDeletable(cutoff time.Time, minCount int64) []BackupEntry {
    count := entries.countValidBackups()
    if count <= minCount {
        return []BackupEntry{}
    }

    toDeleteInc := []BackupEntry{}
    toDelete := []BackupEntry{}

	for _, entry := range entries.Entries {
        if entry.Timestamp.Equal(cutoff) || entry.Timestamp.Before(cutoff) {
            if !entry.DumpFound {
                toDeleteInc = append(toDeleteInc, entry)
                continue
            }

            toDelete = append(toDelete, entry)
        }
	}

    if (count - int64(len(toDelete))) < minCount {
        slices.SortFunc(toDelete, func(a, b BackupEntry) int {
            if a.Timestamp.Equal(b.Timestamp) {
                return 0
            }

            if a.Timestamp.After(b.Timestamp) {
                return 1
            }

            return -1
        })

        toRecup := int64(len(toDelete)) - (count - minCount)
        toDelete = toDelete[:int64(len(toDelete))-toRecup]
    }

    return append(toDelete, toDeleteInc...)
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