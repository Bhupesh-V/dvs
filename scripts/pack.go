//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	"os"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

const (
	busybox     = "library/busybox"
	tag         = "latest"
	authURL     = "https://auth.docker.io/token?service=registry.docker.io&scope=repository:" + busybox + ":pull"
	manifestURL = "https://registry-1.docker.io/v2/" + busybox + "/manifests/" + tag
)

type TokenResponse struct {
	Token string `json:"token"`
}

type Platform struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
}

type Manifest struct {
	Digest    string   `json:"digest"`
	MediaType string   `json:"mediaType"`
	Platform  Platform `json:"platform"`
}

type ManifestList struct {
	Manifests []Manifest `json:"manifests"`
}

func main() {
	tokenResp, err := http.Get(authURL)
	if err != nil {
		panic(err)
	}
	defer tokenResp.Body.Close()

	body, _ := io.ReadAll(tokenResp.Body)
	var tokenData TokenResponse
	json.Unmarshal(body, &tokenData)

	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", manifestURL, nil)
	req.Header.Set("Authorization", "Bearer "+tokenData.Token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.list.v2+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)

	var manifestList ManifestList
	json.Unmarshal(body, &manifestList)

	targets := map[string]bool{
		"linux/amd64": true,
		"linux/arm64": true,
	}

	fmt.Println("Digests for busybox:latest:")
	digests := make(map[string]string)
	for _, m := range manifestList.Manifests {
		key := fmt.Sprintf("%s/%s", m.Platform.OS, m.Platform.Architecture)
		if targets[key] {
			fmt.Printf("  %s: %s\n", key, m.Digest)
			digests[key] = m.Digest
		}
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	fmt.Println("\nRemoving existing busybox images...")
	bimages, err := cli.ImageList(ctx, image.ListOptions{Filters: filters.NewArgs(filters.Arg("reference", "busybox"))})
	if err != nil {
		panic(err)
	}

	for _, img := range bimages {
		for _, tag := range img.RepoTags {
			if len(tag) >= 7 && tag[:7] == "busybox" {
				fmt.Printf("Removing image: %s\n", tag)
				_, err := cli.ImageRemove(ctx, img.ID, image.RemoveOptions{Force: true})
				if err != nil {
					fmt.Printf("Error removing image %s: %v\n", tag, err)
				}
				break
			}
		}
	}

	for platform, digest := range digests {
		arch := "amd64"
		if platform == "linux/arm64" {
			arch = "arm64"
		}

		imageRef := busybox + "@" + digest
		tagName := "busybox:" + arch
		fileName := fmt.Sprintf("images/busybox_%s.tar", arch)

		fmt.Printf("\nPulling %s\n", imageRef)
		pullResp, err := cli.ImagePull(ctx, imageRef, image.PullOptions{})
		if err != nil {
			panic(err)
		}

		defer pullResp.Close()

		// Wait for pull to complete by reading the response
		_, err = io.Copy(io.Discard, pullResp)
		if err != nil {
			panic(fmt.Sprintf("Failed to read image pull response: %v", err))
		}

		fmt.Printf("Tagging as %s\n", tagName)
		err = cli.ImageTag(ctx, imageRef, tagName)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Saving to %s\n", fileName)
		saveResp, err := cli.ImageSave(ctx, []string{tagName})
		if err != nil {
			panic(err)
		}

		outFile, err := os.Create(fileName)
		if err != nil {
			panic(err)
		}

		_, err = io.Copy(outFile, saveResp)
		if err != nil {
			panic(err)
		}

		outFile.Close()
		saveResp.Close()
		fmt.Printf("Successfully saved %s\n", fileName)
	}
}
