package dpkg

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const dpkgStatusPath = "/var/lib/dpkg/status"

func QueryInstalled() (map[string]struct{}, error) {
	installed := make(map[string]struct{})
	file, err := os.Open(dpkgStatusPath)
	if err != nil {
		return nil, fmt.Errorf("could not open dpkg status file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("Error closing file %s: %v", dpkgStatusPath, err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
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
