package analyzer

import (
	"fmt"
	"os/exec"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
)

// LayerAnalysis represents the analysis of a single layer
type LayerAnalysis struct {
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
	Command   string `json:"command"`
	CreatedBy string `json:"created_by"`
	DiffID    string `json:"diff_id"`
}

// ImageAnalysis represents the full analysis of an image
type ImageAnalysis struct {
	ImageName string          `json:"image_name"`
	TotalSize int64           `json:"total_size"`
	Layers    []LayerAnalysis `json:"layers"`
	Config    *v1.ConfigFile  `json:"config"`
}

// AnalyzeImage pulls and inspects the given image
func AnalyzeImage(imageName string, allowRemote bool) (*ImageAnalysis, error) {
	var img v1.Image
	var err error
	var source string

	// Try local daemon first
	if CheckDaemon() {
		fmt.Println("Checking local daemon for image...")
		ref, err := name.ParseReference(imageName)
		if err == nil {
			img, err = daemon.Image(ref)
			if err == nil {
				source = "local"
			} else {
				fmt.Printf("Image not found locally: %v\n", err)
			}
		} else {
			fmt.Printf("Invalid image name: %v\n", err)
		}
	} else {
		fmt.Println("Docker daemon not running. Skipping local check.")
	}

	// Fallback to remote if not found locally
	if img == nil {
		if !allowRemote {
			return nil, fmt.Errorf("image not found locally and remote pull disabled (use --remote to enable)")
		}
		fmt.Println("Pulling from remote...")
		img, err = crane.Pull(imageName)
		if err != nil {
			return nil, fmt.Errorf("failed to pull image %s: %w", imageName, err)
		}
		source = "remote"
	}

	fmt.Printf("Analyzing %s image: %s\n", source, imageName)

	manifest, err := img.Manifest()
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest: %w", err)
	}
	_ = manifest // Keep manifest for potential future use, or remove if strictly unused.

	configFile, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("failed to get config file: %w", err)
	}

	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("failed to get layers: %w", err)
	}

	var layerAnalyses []LayerAnalysis
	history := configFile.History

	// Map layers to history
	// Note: History entries correspond to layers, but empty layers (like ENV vars) are also in history.
	// We need to align them carefully.
	layerIdx := 0
	for _, h := range history {
		if h.EmptyLayer {
			continue
		}
		if layerIdx >= len(layers) {
			break
		}

		l := layers[layerIdx]
		size, _ := l.Size()
		digest, _ := l.Digest()
		diffID, _ := l.DiffID()

		layerAnalyses = append(layerAnalyses, LayerAnalysis{
			Digest:    digest.String(),
			Size:      size,
			Command:   h.CreatedBy,
			CreatedBy: h.CreatedBy,
			DiffID:    diffID.String(),
		})
		layerIdx++
	}

	// Calculate total size
	var totalSize int64
	for _, l := range layers {
		s, _ := l.Size()
		totalSize += s
	}

	return &ImageAnalysis{
		ImageName: imageName,
		TotalSize: totalSize,
		Layers:    layerAnalyses,
		Config:    configFile,
	}, nil
}

// CheckDaemon checks if the Docker daemon is running
func CheckDaemon() bool {
	// Simple check: try to list images or just ping
	// We can use the crane/daemon package to check availability implicitly
	// or check for the docker socket/pipe.
	// For simplicity, we'll assume if we can't connect, it's not running.
	// However, go-containerregistry's daemon package doesn't have a direct "Ping".
	// We will try to get a known image or just rely on the error from daemon.Image.
	// But to be more robust as requested:

	// A quick way is to run "docker info" via command execution,
	// but we want to stay in Go if possible.
	// Let's stick to the "try and fail" approach in AnalyzeImage for now,
	// but we can add a helper here if needed.

	// Actually, let's try to execute "docker info" to be sure.
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		// Try podman
		cmd = exec.Command("podman", "info")
		if err := cmd.Run(); err != nil {
			return false
		}
	}
	return true
}
