package analyzer

import (
	"fmt"

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
func AnalyzeImage(imageName string, allowRemote bool, pullMissing bool) (*ImageAnalysis, error) {
	var img v1.Image
	var err error
	var source string

	if allowRemote {
		fmt.Println("Pulling from remote...")
		img, err = crane.Pull(imageName)
		if err != nil {
			return nil, fmt.Errorf("failed to pull image %s: %w", imageName, err)
		}
		source = "remote"
	} else {
		fmt.Println("Checking local daemon for image...")
		ref, err := name.ParseReference(imageName)
		if err != nil {
			return nil, fmt.Errorf("invalid image name: %w", err)
		}
		img, err = daemon.Image(ref)
		if err != nil {
			if pullMissing {
				fmt.Println("Image not found locally. Pulling missing image from remote...")
				img, err = crane.Pull(imageName)
				if err != nil {
					return nil, fmt.Errorf("failed to pull missing image %s: %w", imageName, err)
				}
				source = "remote (missing image)"
			} else {
				return nil, fmt.Errorf("image not found locally and remote pull disabled (use --remote to enable): %w", err)
			}
		}
		if source == "" {
			source = "local"
		}
	}

	fmt.Printf("Analyzing %s image: %s\n", source, imageName)

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
	layerAnalyses = make([]LayerAnalysis, 0, len(layers))

	type layerMeta struct {
		size   int64
		digest string
		diffID string
	}
	layerMetadata := make([]layerMeta, 0, len(layers))
	var totalSize int64
	for _, l := range layers {
		size, err := l.Size()
		if err != nil {
			return nil, fmt.Errorf("failed to get layer size: %w", err)
		}
		digest, err := l.Digest()
		if err != nil {
			return nil, fmt.Errorf("failed to get layer digest: %w", err)
		}
		diffID, err := l.DiffID()
		if err != nil {
			return nil, fmt.Errorf("failed to get layer diff id: %w", err)
		}

		totalSize += size
		layerMetadata = append(layerMetadata, layerMeta{
			size:   size,
			digest: digest.String(),
			diffID: diffID.String(),
		})
	}

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

		meta := layerMetadata[layerIdx]

		layerAnalyses = append(layerAnalyses, LayerAnalysis{
			Digest:    meta.digest,
			Size:      meta.size,
			Command:   h.CreatedBy,
			CreatedBy: h.CreatedBy,
			DiffID:    meta.diffID,
		})
		layerIdx++
	}

	return &ImageAnalysis{
		ImageName: imageName,
		TotalSize: totalSize,
		Layers:    layerAnalyses,
		Config:    configFile,
	}, nil
}
