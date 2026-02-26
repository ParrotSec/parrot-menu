package desktop

import (
	"bufio"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const DirLauncherDest = "/usr/share/applications/"

func GetXPackageName(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			slog.Error("failed to close file", "path", path, "err", err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "X-Parrot-") {
			if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
		if err := scanner.Err(); err != nil {
			slog.Error("failed to read desktop file", "path", path, "err", err)
			return "", err
		}
	}
	return "", nil
}

func FixOldLaunchers(fileName string) {
	// If a new launcher (e.g., "serv-tool.desktop") is installed, this function
	// ensures that the older version (e.g., "parrot-toolname.desktop") is removed
	// to avoid duplicates in the application menu.

	newNamePrefixes := []string{"serv-"}
	for _, checkName := range newNamePrefixes {
		if suffix, found := strings.CutPrefix(fileName, checkName); found {
			oldFileName := "parrot-" + suffix
			destPath := filepath.Join(DirLauncherDest, oldFileName)
			if _, err := os.Stat(destPath); err == nil {
				if err := os.Remove(destPath); err != nil {
					slog.Error("could not remove duplicate launcher", "destPath", destPath, "err", err)
				}
			}
			break
		}
	}
}

func CopyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(source *os.File) {
		err := source.Close()
		if err != nil {
			slog.Error("failed to close file", "src", src, "err", err)
		}
	}(source)

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(destination *os.File) {
		err := destination.Close()
		if err != nil {
			slog.Error("failed to close file", "dst", dst, "err", err)
		}
	}(destination)

	scanner := bufio.NewScanner(source)
	writer := bufio.NewWriter(destination)

	for scanner.Scan() {
		line := scanner.Text()

		// Desktop entries usually prefer icon names without extensions.
		// If the Icon field explicitly specifies .png, remove it.
		if strings.HasPrefix(line, "Icon=") {
			if newLine, found := strings.CutSuffix(line, ".png"); found {
				line = newLine
			}
		}

		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		slog.Error("error reading file", "src", src, "err", err)
		return err
	}

	return writer.Flush()
}
