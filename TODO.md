# Todo

1. Add warn logs by checking linux container emulation is enabled or not. Read the `OSType` in response to following request (should be `linux`).
   1. Mac
      ```
      curl --unix-socket /var/run/docker.sock http://localhost/info
      ```
   2. Windows
      ```
      curl --noproxy '*' http://localhost:2375/info
      ```
2. Use `--use-compress-program` with parallel compression tools (for big volumes)
   ```
   tar -cv --use-compress-program=pigz -f /dest/filename.tar.gz -C /source .
   # or
   tar -cv --use-compress-program='zstd -T0' -f /dest/filename.tar.zst -C /source .
   ```
3. Figure out encrption/decryption using both symmetric and asymmetric approaches.
4. Add check for when the volume is already mounted with any other container (only for restore case).
   ```bash
   docker ps -q | while read cid; do
      docker inspect "$cid" \
         --format='{{.Name}}: {{range .Mounts}}{{if eq .Name "your_volume_name"}}USED{{end}}{{end}}'
   done
   ```