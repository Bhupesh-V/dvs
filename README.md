# dvs

Create & Restore docker volumes snapshots

> dvs is the cross platform go based port for [docker-volume-snapshot](https://github.com/junedkhatri31/docker-volume-snapshot) originally created by Juned Khatri.

## Installation

Download the CLI by using the following URL

https://github.com/Bhupesh-V/dvs/releases/latest/download/dvs


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