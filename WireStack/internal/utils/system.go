package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// RunCommand executes the named program with arguments and returns trimmed stdout.
func RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command %s failed: %w (%s)", name, err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

// RunCommandWithInput runs the named program with stdin populated and returns trimmed stdout.
func RunCommandWithInput(input string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = bytes.NewBufferString(input)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command %s failed: %w (%s)", name, err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}
