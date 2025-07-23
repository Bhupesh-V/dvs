# dvs

Create & Restore docker volumes snapshots

> `dvs` is the cross platform port for [docker-volume-snapshot](https://github.com/junedkhatri31/docker-volume-snapshot) originally created by Juned Khatri.

## Why do I need this?

- Exchange persistent data across team members to reduce project onboarding time.
- Back up data when switching workstations.
- If you manage storage solutions (e.g., MongoDB, OpenSearch, Postgres) using Docker volumes, backing up and restoring data is easy.

## Installation

### Quick Download

The binary is present inside the archive.

| Platform | Architecture | Download |
|----------|--------------|----------|
| Linux | amd64 | [dvs-linux-amd64.tar.gz](https://github.com/Bhupesh-V/dvs/releases/latest/download/dvs-linux-amd64.tar.gz) |
|  | arm64 | [dvs-linux-arm64.tar.gz](https://github.com/Bhupesh-V/dvs/releases/latest/download/dvs-linux-arm64.tar.gz) |
| macOS | amd64 (Intel) | [dvs-darwin-amd64.tar.gz](https://github.com/Bhupesh-V/dvs/releases/latest/download/dvs-darwin-amd64.tar.gz) |
|  | arm64 (Apple Silicon) | [dvs-darwin-arm64.tar.gz](https://github.com/Bhupesh-V/dvs/releases/latest/download/dvs-darwin-arm64.tar.gz) |
| Windows | amd64 | [dvs-windows-amd64.zip](https://github.com/Bhupesh-V/dvs/releases/latest/download/dvs-windows-amd64.zip) |
|  | arm64 | [dvs-windows-arm64.zip](https://github.com/Bhupesh-V/dvs/releases/latest/download/dvs-windows-arm64.zip) |

Extract the `dvs` executable and run it to verify that you see the following output.

```
A tool to create and restore snapshots of Docker volumes.

Usage:
  dvs [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  create      Create snapshot file from docker volume
  help        Help about any command
  restore     Restore snapshot file to docker volume

Flags:
  -h, --help   help for dvs

Use "dvs [command] --help" for more information about a command.
```

## Usage

### Create a snapshot

```
dvs create <source_volume> <destination_file>
```

### Restore from snapshot

Make sure the `destination_volume` exists before restoring data to it.

```
dvs restore <snapshot_file> <destination_volume>
```