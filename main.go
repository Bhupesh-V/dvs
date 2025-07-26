// Based on Juned's work in https://github.com/junedkhatri31/docker-volume-snapshot
package main

import (
	"bytes"
	"context"
	"dvs/images"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	// TODO: change to moby when the migration is complete

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

var rootCmd = &cobra.Command{
	Use:   "dvs",
	Short: "Docker Volume Snapshot (dvs)",
	Long:  "A tool to create and restore snapshots of Docker volumes.",
}

var createCmd = &cobra.Command{
	Use:     "create <source_volume> <destination_file>",
	Short:   "Create snapshot file from docker volume",
	Example: "dvs create my_volume my_volume.tar.gz",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fatal(fmt.Sprintf("Unable to create Docker client: %v", err))
		}
		runCreate(ctx, cli, args[0], args[1])
	},
}

var restoreCmd = &cobra.Command{
	Use:     "restore <snapshot_file> <destination_volume>",
	Short:   "Restore snapshot file to docker volume",
	Example: "dvs restore my_volume.tar.gz my_volume",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fatal(fmt.Sprintf("Unable to create Docker client: %v", err))
		}
		runRestore(ctx, cli, args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(restoreCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func runCreate(ctx context.Context, cli *client.Client, volume string, outputFile string) {
	outputDir := resolveDir(outputFile)
	ensureDir(outputDir)

	filename := filepath.Base(outputFile)

	vol := volumeHealthCheck(ctx, cli, volume)

	fmt.Println("Creating snapshot of volume:", vol)

	containerConfig := &container.Config{
		Cmd: []string{"tar", "cvaf", "/dest/" + filename, "-C", "/source", "."},
	}
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: vol,
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

	vol := volumeHealthCheck(ctx, cli, volume)

	fmt.Println("Restoring snapshot from:", snapshotPath)

	containerConfig := &container.Config{
		Cmd: []string{"tar", "xvf", "/source/" + filename, "-C", "/dest"},
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
				Source: vol,
				Target: "/dest",
			},
		},
		AutoRemove: true,
	}

	runContainer(ctx, cli, containerConfig, hostConfig)
	fmt.Println("Snapshot restored to volume:", vol)
}

func runContainer(ctx context.Context, cli *client.Client, config *container.Config, hostConfig *container.HostConfig) {
	arch := getArch()
	if arch == "" {
		fatal("Unsupported architecture: " + arch)
	}

	// TODO: add custom dvs tag to not mess with user's images
	config.Image = fmt.Sprintf("busybox:%s", arch)

	_, err := cli.ImageInspect(ctx, config.Image)
	if err != nil {
		if client.IsErrConnectionFailed(err) {
			fatal("Ensure the Docker daemon is up and running.")
		}
		if errdefs.IsNotFound(err) {
			rdr := bytes.NewReader(images.Busybox)
			_, loadErr := cli.ImageLoad(ctx, rdr)
			if loadErr != nil {
				out, pullErr := cli.ImagePull(ctx, config.Image, image.PullOptions{})
				if pullErr != nil {
					fatal("Failed to pull busybox image")
				}
				defer out.Close()

				// Wait for pull to complete by reading the response
				_, err = io.Copy(io.Discard, out)
				if err != nil {
					fatal("Failed to read image pull response")
				}
			}
		}
	}

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		fatal("Container creation failed")
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		fatal("Failed to start container")
	}

	// Change from WaitConditionRemoved to WaitConditionNotRunning
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			fatal(fmt.Sprintf("Container execution error: %v", err))
		}
	case status := <-statusCh:
		// Check if the container exited with an error
		if status.StatusCode != 0 {
			fatal(fmt.Sprintf("Container exited with status code: %d", status.StatusCode))
		}
	}
}

func resolveDir(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		fatal(fmt.Sprintf("Unable to resolve path: %v", path))
	}
	return filepath.Dir(absPath)
}

func ensureDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fatal(fmt.Sprintf("Failed to create directory: %v", err))
		}
	}
}

// Checks if the volume exists and returns its name.
func volumeHealthCheck(ctx context.Context, cli *client.Client, volume string) string {
	vol, err := cli.VolumeInspect(ctx, volume)
	if err != nil {
		if errdefs.IsNotFound(err) {
			fatal(fmt.Sprintf("Volume '%s' does not exist.", volume))
		} else {
			fatal(fmt.Sprintf("Failed to inspect volume '%s'", volume))
		}
	}

	return vol.Name
}

func getArch() string {
	return runtime.GOARCH
}
