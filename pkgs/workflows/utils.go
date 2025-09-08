package workflows

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func runBinaryWithPipe(binary, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	cmd := exec.Command(binary)
	cmd.Stdin = file

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("binary failed: %v, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// runBinaryWithString pipes the given input string into the binary's stdin
// and returns the binary's stdout output.
func runBinaryWithString(binary string, input Outputres) (string, error) {
	// Prepare the command
	cmd := exec.Command(binary)

	// Pipe string into stdin
	cmd.Stdin = strings.NewReader(string(input))

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("binary failed: %v, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}
