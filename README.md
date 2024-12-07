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

The Configuration file is a yaml file which takes the following keys hiearchy:

- **etcd_client**: Parameters for etcd communication.
  - **endpoints**: List of etcd endpoints (**ip:port** format)
  - **connection_timeout**: Etcd connection timeout as a duration (ex: 1m)
  - **request_timeout**: Etcd request timeout as a duration (ex: 1m)
  - **retries**: Number of retries when an etcd operaiton fails 
  - **auth**: Etcd authentication parameters
    - **ca_cert**: Path to a CA certificate file the client will use to validate the server certificates of the etcd cluster.
    - **password_auth**: Path to a yaml containing a **username** and **password** key to be used if password client authentication is employed for the etcd cluster.
    - **client_cert**: Client certificate file, to be used if certificate client authentication is employed for the etcd cluster.
    - **client_key**: Client private key file, to be used if certificate client authentication is employed for the etcd cluster.
- **snapshot_path**: Path where to temporarily store the transient snapshot file for the **backup** and **restore** commands. Note that this file is temporary and will always be deleted by the time the command completes.
- **encryption_key_path**: Path to the file containg the master key for encrypting and decryption backups in the **backup** and **restore** commands. You can omit it if you do not wish to encrypt your backups. Also used to specify the file that contains the new master key with the **rotate-key** command. 
- **s3_client**: Parameters for s3 communication.
  - **objects_prefix**: Prefix to put on all s3 objects. Backups will be stored in objects named `<object_prefix>-<timestamp>.dump` and encrypted encryption keys will be stored in objects named `<object_prefix>-<timestamp>.key`. The default value is **backup** if omited.
  - **endpoint**: Endpoint of the s3 store. Takes the format **ip:port**.
  - **bucket**: Bucket in the s3 store where the backups are managed.
  - **auth**: S3 Authentication parameters.
    - **ca_cert**: Path to a CA file used to authentify the S3 store server certificate. Can be omited if the S3 store server certificate has been signed by a well established CA.
    - **key_auth**: Path to a yaml authentication file containing two keys: **access_key** and **secret_key**. These are the credentials the client will present to the S3 store.
  - **region**: Region to use in the s3 store.
  - **connection_timeout**: S3 connection timeout as a duration (ex: 1m)
  - **request_timeout**: S3 request timeout as a duration (ex: 1m)

