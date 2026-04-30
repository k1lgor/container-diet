package mcp

import (
	"strings"
	"testing"

	"github.com/k1lgor/container-diet/internal/analyzer"
)

func TestFormatImageSummary(t *testing.T) {
	analysis := &analyzer.ImageAnalysis{
		ImageName: "nginx:latest",
		TotalSize: 142 * 1024 * 1024, // 142 MB
		Layers: []analyzer.LayerAnalysis{
			{
				Size:    71 * 1024 * 1024,
				Command: "/bin/sh -c apt-get update && apt-get install -y curl wget vim",
			},
			{
				Size:    71 * 1024 * 1024,
				Command: "COPY nginx.conf /etc/nginx/nginx.conf",
			},
		},
	}

	summary := formatImageSummary(analysis)

	// Check key content is present
	for _, want := range []string{
		"nginx:latest",
		"MB",
		"Layers: 2",
		"Layer Breakdown",
		"apt-get update",
		"COPY nginx.conf",
	} {
		if !strings.Contains(summary, want) {
			t.Errorf("summary missing %q\nGot:\n%s", want, summary)
		}
	}

	// Long commands should be truncated
	if strings.Contains(summary, "vim") {
		t.Error("expected long command to be truncated, but saw full command")
	}
}

func TestFormatImageSummary_ShortCommands(t *testing.T) {
	analysis := &analyzer.ImageAnalysis{
		ImageName: "alpine:3.19",
		TotalSize: 7 * 1024 * 1024,
		Layers: []analyzer.LayerAnalysis{
			{Size: 7 * 1024 * 1024, Command: "ADD file"},
		},
	}

	summary := formatImageSummary(analysis)

	if !strings.Contains(summary, "alpine:3.19") {
		t.Error("summary missing image name")
	}
	if !strings.Contains(summary, "ADD file") {
		t.Error("summary missing short command")
	}
	if !strings.Contains(summary, "Layers: 1") {
		t.Error("summary missing layer count")
	}
}

func TestFormatImageSummary_EmptyLayers(t *testing.T) {
	analysis := &analyzer.ImageAnalysis{
		ImageName: "scratch",
		TotalSize: 0,
		Layers:    []analyzer.LayerAnalysis{},
	}

	summary := formatImageSummary(analysis)

	if !strings.Contains(summary, "scratch") {
		t.Error("summary missing image name")
	}
	if !strings.Contains(summary, "Layers: 0") {
		t.Error("summary should show 0 layers")
	}
}
