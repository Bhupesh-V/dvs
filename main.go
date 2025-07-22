package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	// TODO: change to moby when the migration is complete

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

func usageAndExit() {
	fmt.Println(`usage: dvs (create|restore) volume destination
  create         create snapshot file from docker volume
  restore        restore snapshot file to docker volume
  volume         volume name
  destination    destination path ending with .tar or .tar.gz

Examples:
  dvs create xyz_volume xyz_volume.tar.gz
  dvs restore xyz_volume.tar.gz xyz_volume`)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 4 {
		usageAndExit()
	}

	mode := os.Args[1]
	source := os.Args[2]
	dest := os.Args[3]

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Unable to create Docker client: %v", err)
	}

	switch mode {
	case "create":
		runCreate(ctx, cli, source, dest)
	case "restore":
		runRestore(ctx, cli, source, dest)
	default:
		usageAndExit()
	}
}

func runCreate(ctx context.Context, cli *client.Client, volume string, outputFile string) {
	outputDir := resolveDir(outputFile)
	ensureDir(outputDir)

	filename := filepath.Base(outputFile)

	containerConfig := &container.Config{
		Image: "busybox",
		Cmd:   []string{"tar", "cvaf", "/dest/" + filename, "-C", "/source", "."},
	}
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: volume,
				Target: "/source",
			},
			{
				Type:   mount.TypeBind,
				Source: outputDir,
				Target: "/dest",
			},
		},
		AutoRemove: true,
	}

	runContainer(ctx, cli, containerConfig, hostConfig)
	fmt.Printf("Snapshot created at %s\n", outputFile)
}

func runRestore(ctx context.Context, cli *client.Client, snapshotPath string, volume string) {
	inputDir := resolveDir(snapshotPath)
	filename := filepath.Base(snapshotPath)

	containerConfig := &container.Config{
		Image: "busybox",
		Cmd:   []string{"tar", "xvf", "/source/" + filename, "-C", "/dest"},
	}
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: inputDir,
				Target: "/source",
			},
			{
				Type:   mount.TypeVolume,
				Source: volume,
				Target: "/dest",
			},
		},
		AutoRemove: true,
	}

	runContainer(ctx, cli, containerConfig, hostConfig)
	fmt.Println("Snapshot restored to volume:", volume)
}

func runContainer(ctx context.Context, cli *client.Client, config *container.Config, hostConfig *container.HostConfig) {
	_, _, err := cli.ImageInspectWithRaw(ctx, config.Image)
	if err != nil {
		out, pullErr := cli.ImagePull(ctx, config.Image, image.PullOptions{})
		if pullErr != nil {
			log.Fatalf("Failed to pull image: %v", pullErr)
		}
		defer out.Close()
	}

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		log.Fatalf("Container creation failed: %v", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		log.Fatalf("Failed to start container: %v", err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionRemoved)
	select {
	case err := <-errCh:
		if err != nil {
			log.Fatalf("Container error: %v", err)
		}
	case <-statusCh:
	}
}

func resolveDir(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("Unable to resolve path: %v", err)
	}
	return filepath.Dir(absPath)
}

func ensureDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("Creating directory:", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory: %v", err)
		}
	}
}
