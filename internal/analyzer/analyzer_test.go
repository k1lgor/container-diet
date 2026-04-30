package analyzer

import (
	"testing"
)

func TestAnalyzeImage_InvalidImageName(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		remote    bool
	}{
		{"empty name", "", false},
		{"invalid chars local", "INVALID UPPER", false},
		{"invalid chars remote", "INVALID UPPER", true},
		{"double colon remote", "image::tag", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AnalyzeImage(tt.imageName, tt.remote, false)
			if err == nil {
				t.Errorf("expected error for invalid image name %q", tt.imageName)
			}
		})
	}
}

func TestAnalyzeImage_NoDaemon(t *testing.T) {
	// When Docker daemon is not available and remote=false,
	// we should get a clear error
	_, err := AnalyzeImage("nonexistent-image:latest", false, false)
	if err == nil {
		t.Error("expected error when image not in local daemon and remote disabled")
	}
}

func TestLayerAnalysisStruct(t *testing.T) {
	la := LayerAnalysis{
		Digest:    "sha256:abc123",
		Size:      1024,
		Command:   "RUN apt-get update",
		CreatedBy: "/bin/sh -c apt-get update",
		DiffID:    "sha256:def456",
	}

	if la.Digest != "sha256:abc123" {
		t.Errorf("Digest mismatch")
	}
	if la.Size != 1024 {
		t.Errorf("Size mismatch")
	}
	if la.Command != "RUN apt-get update" {
		t.Errorf("Command mismatch")
	}
}

func TestImageAnalysisStruct(t *testing.T) {
	ia := ImageAnalysis{
		ImageName: "test:latest",
		TotalSize: 2048,
		Layers: []LayerAnalysis{
			{Size: 1024, Command: "layer1"},
			{Size: 1024, Command: "layer2"},
		},
	}

	if ia.ImageName != "test:latest" {
		t.Errorf("ImageName mismatch")
	}
	if ia.TotalSize != 2048 {
		t.Errorf("TotalSize mismatch")
	}
	if len(ia.Layers) != 2 {
		t.Errorf("expected 2 layers, got %d", len(ia.Layers))
	}
}
