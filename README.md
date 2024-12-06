# About

This is an utility to backup the state of an etcd cluster in an s3 store and restore the etcd state from a snapshot.

Backups can be done remotely as it relies on the etcd snapshot api. Restores, however, should be performed from the etcd node restoring the data and additionally require the **etcdutl** binary to be present on the etcd node to unpack the downloaded snapshot into etcd's data directory.

The utility optionally supports encryption of the backups (and decryption during restores) using the official XChaCha20-Poly1305 encryption implementation from the golang crypto package. It takes a master key as an argument for the encryption/decryption and will generate an encryption key unique to each backup which it will encrypt with the master key. It is recommended to use a different master key to encrypt the backup of different etcd clusters.

The utility also provides a command to prune aging backups as needed.

# Usage

## Commands

All the commands support a **-c**/**--config** argument to specify the path of the configuration file which defaults to **config.yml** (present in the execution directory).

The utility has the following commands:
  - **backup**: Command to backup a snapshot of the s3 store
  - **restore**: Command to restore a snapshot on the etcd node. It takes the following arguments:
    - **-t**/**--backup-timestamp**: Timestamp of the backup to restore in RFC3339 format (ex: **2024-12-06T21:22:25Z**). If omited, the lastest backup will be restored.
    - **-d**/**--data-dir**: Path of the etcd data directory on the node where the snapshot will be unpacked. This is a mandatory argument.
    - **-e**/**--etcdutl-path**: Path of the **etcdutl** binary which will be used to unpack the snapshot on the filesystem. Can be omited if **etcdutl** is already in the system's **PATH**.
  - **rotate-key**: Command to rotate the master key that is encrypting the backups. It takes the following arguments:
    - **-p**/**--previous-key**: Path to a file containing the previous key that was used to encrypt the backup encryption keys currently in s3. This is a mandatory argument. The file containing the key used to re-encrypt the encryption keys in the s3 store is specified in the configuration file.
  - **prune**: Command to prune aging backups. It takes the following arguments:
    - **-a**/**--max-age**: Maximum age of the backups that should be kept, as a duration (ex: "15d", "10w", "1y"). Backups that are older will be deleted. Defaults to **15d** (15 days).
    - **-i**/**--min-count**: Absolute minimum number of backups that should remain after pruning, regardless of the **max-age** argument. If a prune operation would cause fewer backups to remain, newer backups scheduled for deletion will not be deleted. Defaults to **20**.

## Configuration

...

