package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// downloadClip executes yt-dlp safely using context and argument separation
func downloadClip(ctx context.Context, url, start, end, quality string) error {
	// 1. Sanitization
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, "http") {
		return fmt.Errorf("invalid URL: must start with http/https")
	}

	// 2. Resolve Quality
	format := "bestvideo+bestaudio/best"
	if quality != "best" {
		format = fmt.Sprintf("bestvideo[height<=%s]+bestaudio/best[height<=%s]/best", quality, quality)
	}

	// 3. Section Syntax (yt-dlp format: *start-end)
	section := fmt.Sprintf("*%s-%s", strings.TrimSpace(start), strings.TrimSpace(end))

	// 4. Arguments with "--" to prevent flag injection
	args := []string{
		"-f", format,
		"--download-sections", section,
		"--force-keyframes-at-cuts",
		"--merge-output-format", "mp4",
		"-o", "clip_%(title)s.%(ext)s",
		"--", // Critical security: ensures URL is not parsed as a flag
		url,
	}

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	
	// Direct output so user sees yt-dlp's native progress bar
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// checkDeps ensures the user has required CLI tools installed
func checkDeps() error {
	deps := []string{"yt-dlp", "ffmpeg"}
	for _, bin := range deps {
		if _, err := exec.LookPath(bin); err != nil {
			return fmt.Errorf("dependency missing: '%s' not found in PATH", bin)
		}
	}
	return nil
}