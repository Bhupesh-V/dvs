# Author Notes


## Testing snapshot creating of big volumes

```
docker run --rm -it -v big-volume:/data alpine sh
```

Create 10 files, each 2GB (change `count` to control size)

```
for i in $(seq 1 10); do dd if=/dev/zero of=/data/file_$i bs=1M count=2024; done
```

```
docker run --rm -v big-volume:/data alpine ls -lh /data
```

Creating a snapshot of a volume sized 20GB, dvs took 2-3 mins (although it can be highly wrong)
```
real    2m8.993s
user    0m0.009s
sys     0m0.016s
```

Restoring the same to a fresh volume took 1-2 mins

```
real    1m55.040s
user    0m0.010s
sys     0m0.018s
```