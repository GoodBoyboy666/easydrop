package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pterm/pterm"
)

func TestPrintBuildInfoBanner(t *testing.T) {
	originalName := appDisplayName
	originalVersion := appVersion
	originalBuildTime := buildTime
	originalCommit := gitCommit
	t.Cleanup(func() {
		appDisplayName = originalName
		appVersion = originalVersion
		buildTime = originalBuildTime
		gitCommit = originalCommit
	})

	appDisplayName = "EasyDrop"
	appVersion = "v1.2.3"
	buildTime = "2026-04-04T12:34:56Z"
	gitCommit = "abcdef1234567890"

	var buf bytes.Buffer
	printBuildInfoBanner(&buf)

	output := pterm.RemoveColorFromString(buf.String())
	expected := []string{
		"Program    : EasyDrop",
		"Version    : v1.2.3",
		"Build Time : 2026-04-04T12:34:56Z",
		"Commit     : abcdef1234567890",
		"EasyDrop Runtime",
	}

	for _, item := range expected {
		if !strings.Contains(output, item) {
			t.Fatalf("expected banner to contain %q, got %q", item, output)
		}
	}

	if !strings.Contains(output, "┌") || !strings.Contains(output, "┘") {
		t.Fatalf("expected pterm box border in output, got %q", output)
	}
}
