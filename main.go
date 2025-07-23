package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	// TODO: change to moby when the migration is complete

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
	Use:   "create <source_volume> <destination_file>",
	Short: "Create snapshot file from docker volume",
	Args:  cobra.ExactArgs(2),
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
	Use:   "restore <snapshot_file> <destination_volume>",
	Short: "Restore snapshot file to docker volume",
	Args:  cobra.ExactArgs(2),
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
	_, err := cli.ImageInspect(ctx, config.Image)
	if err != nil {
		if client.IsErrConnectionFailed(err) {
			fatal("Please check your Docker daemon connection.")
		}

		out, pullErr := cli.ImagePull(ctx, config.Image, image.PullOptions{})
		if pullErr != nil {
			fatal(fmt.Sprintf("Failed to pull image: %v", pullErr))
		}
		defer out.Close()

		// Wait for pull to complete by reading the response
		_, err = io.Copy(io.Discard, out)
		if err != nil {
			fatal(fmt.Sprintf("Failed to read image pull response: %v", err))
		}
	}

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		fatal(fmt.Sprintf("Container creation failed: %v", err))
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		fatal(fmt.Sprintf("Failed to start container: %v", err))
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionRemoved)
	select {
	case err := <-errCh:
		if err != nil {
			fatal(fmt.Sprintf("Container error: %v", err))
		}
	case <-statusCh:
	}
}

func resolveDir(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		fatal(fmt.Sprintf("Unable to resolve path: %v", err))
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
