package browser

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

// Open launches the system's default web browser to the specified URL after a brief delay.
func Open(url string) error {
	// Give server a moment to start listening
	time.Sleep(100 * time.Millisecond)

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
