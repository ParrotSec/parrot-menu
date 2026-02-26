package dpkg

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const statusPath = "/var/lib/dpkg/status"

func QueryInstalled() (map[string]struct{}, error) {
	// Instead of launching `dpkg -l`, let's open and read `/var/lib/dpkg/status`.
	// If we find "install ok installed" for each package, we add it to a map[string]struct{} for O(1) lookups.
	installed := make(map[string]struct{})

	file, err := os.Open(statusPath)
	if err != nil {
		return nil, fmt.Errorf("could not open dpkg status file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			slog.Error("failed to close file", "statusPath", statusPath, "err", err)
		}
	}(file)

	scanner := bufio.NewScanner(file)

	const maxCapacity = 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	var pkgName string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Package: ") {
			pkgName = strings.TrimPrefix(line, "Package: ")
		} else if line == "Status: install ok installed" && pkgName != "" {
			installed[pkgName] = struct{}{}
			pkgName = ""
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading dpkg status file: %w", err)
	}
	return installed, nil
}
