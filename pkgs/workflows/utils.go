// Copyright (c) 2025 Alexsander Hamir Gomes Baptista
//
// This file is part of AtomOS and licensed under the Sustainable Use License (SUL).
// You may use, modify, and redistribute this software for personal or internal business use.
// Offering it as a commercial hosted service requires a separate license.
//
// Full license: see the LICENSE file in the root of this repository
// or contact alexsanderhamirgomesbaptista@gmail.com.

package workflows

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func runBinaryWithPipe(binary, entry, filePath string) (string, error) {
	file, err := os.Open(filePath)

	cmd := exec.Command(binary, entry)
	if err == nil {
		cmd.Stdin = file
	}
	defer file.Close()

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
func runBinaryWithString(binary, entry string, input Outputres) (string, error) {
	// Prepare the command
	cmd := exec.Command(binary, entry)

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
