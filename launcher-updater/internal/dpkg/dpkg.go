package dpkg

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const queryFormat = `${Package}\t${db:Status-Status}\n`

func QueryInstalled() (map[string]struct{}, error) {
	cmd := exec.Command(
		"dpkg-query",
		"--show",
		"--showformat="+queryFormat,
	)
	cmd.Env = append(os.Environ(), "LC_ALL=C", "LANG=C")

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf(
				"dpkg-query failed: %w: %s",
				err,
				strings.TrimSpace(string(exitErr.Stderr)),
			)
		}
		return nil, fmt.Errorf("run dpkg-query: %w", err)
	}

	return parseInstalled(output)
}

func parseInstalled(output []byte) (map[string]struct{}, error) {
	installed := make(map[string]struct{})
	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		pkgName, status, found := strings.Cut(line, "\t")
		if !found || pkgName == "" || status == "" {
			return nil, fmt.Errorf("unexpected dpkg-query output %q", line)
		}
		if status == "installed" {
			installed[pkgName] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read dpkg-query output: %w", err)
	}
	return installed, nil
}
