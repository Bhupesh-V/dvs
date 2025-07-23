# dvs

Create & Restore docker volumes snapshots

> `dvs` is the cross platform port for [docker-volume-snapshot](https://github.com/junedkhatri31/docker-volume-snapshot) originally created by Juned Khatri.

## Why do I need this?

- Exchange persistent data across team members to reduce project onboarding time.
- Back up data when switching workstations.
- If you manage storage solutions (e.g., MongoDB, OpenSearch, Postgres) using Docker volumes, backing up and restoring data is easy.

## Installation

### Quick Download

| Platform | Architecture | Download |
|----------|--------------|----------|
| Linux | amd64 | [dvs-linux-amd64.tar.gz](https://github.com/Bhupesh-V/dvs/releases/latest/download/dvs-linux-amd64.tar.gz) |
| Linux | arm64 | [dvs-linux-arm64.tar.gz](https://github.com/Bhupesh-V/dvs/releases/latest/download/dvs-linux-arm64.tar.gz) |
| macOS | amd64 (Intel) | [dvs-darwin-amd64.tar.gz](https://github.com/Bhupesh-V/dvs/releases/latest/download/dvs-darwin-amd64.tar.gz) |
| macOS | arm64 (Apple Silicon) | [dvs-darwin-arm64.tar.gz](https://github.com/Bhupesh-V/dvs/releases/latest/download/dvs-darwin-arm64.tar.gz) |
| Windows | amd64 | [dvs-windows-amd64.zip](https://github.com/Bhupesh-V/dvs/releases/latest/download/dvs-windows-amd64.zip) |
| Windows | arm64 | [dvs-windows-arm64.zip](https://github.com/Bhupesh-V/dvs/releases/latest/download/dvs-windows-arm64.zip) |

## Usage

```
Docker Volume Snapshot (dvs)
A tool to create and restore snapshots of Docker volumes.

usage: dvs [create|restore] <source> <destination>
  create         create snapshot file from docker volume
  restore        restore snapshot file to docker volume
  source         source path to the volume or snapshot file
  destination    destination path for the snapshot file or volume name

Examples:
  dvs create my_volume my_volume.tar.gz
  dvs restore my_volume.tar.gz my_volume

```