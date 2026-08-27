package desktop

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const (
	DirLauncherDest         = "/usr/share/applications/"
	installableNamePrefix   = "[not installed] "
	maxInstallableNameRunes = 40
)

var installableNameSeparators = []string{" — ", " - "}

func GetXPackageName(path string) (string, error) {
	return getDesktopValue(path, "X-Parrot-Package")
}

func getDesktopValue(path, key string) (string, error) {
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
	prefix := key + "="
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix)), nil
		}
	}

	if err := scanner.Err(); err != nil {
		slog.Error("failed to read desktop file", "path", path, "err", err)
		return "", err
	}
	return "", nil
}

func IsManaged(path string) bool {
	pkg, err := GetXPackageName(path)
	if err != nil {
		return false
	}
	if pkg != "" {
		return true
	}

	managed, err := getDesktopValue(path, "X-Parrot-Managed")
	return err == nil && strings.EqualFold(managed, "true")
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
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return writeFileAtomic(dst, data)
}

func CopyTemplateLauncher(src, dst, pkgName string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	generated, err := renderTemplateLauncher(data, pkgName)
	if err != nil {
		return fmt.Errorf("render template from %s: %w", src, err)
	}
	return writeFileAtomic(dst, generated)
}

func renderTemplateLauncher(data []byte, pkgName string) ([]byte, error) {
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	output := make([]string, 0, len(lines)+1)
	inDesktopEntry := false
	foundDesktopEntry := false
	hasTerminal := false

	for _, originalLine := range lines {
		line := originalLine
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if inDesktopEntry && !hasTerminal {
				output = append(output, "Terminal=true")
			}
			inDesktopEntry = line == "[Desktop Entry]"
			if inDesktopEntry {
				foundDesktopEntry = true
				hasTerminal = false
			}
			output = append(output, line)
			continue
		}

		if inDesktopEntry {
			key, value, found := strings.Cut(line, "=")
			if found {
				switch {
				case key == "Name" || strings.HasPrefix(key, "Name["):
					line = key + "=" + compactInstallableName(value)
				case key == "Icon":
					line = "Icon=software-manager"
				case key == "Exec":
					line = "Exec=parrot-exec --install " + pkgName
				case key == "Terminal":
					line = "Terminal=true"
					hasTerminal = true
				}
			}
		}
		output = append(output, line)
	}

	if !foundDesktopEntry {
		return nil, fmt.Errorf("missing [Desktop Entry] group")
	}
	if inDesktopEntry && !hasTerminal {
		output = append(output, "Terminal=true")
	}

	return []byte(strings.Join(output, "\n") + "\n"), nil
}

func writeFileAtomic(path string, data []byte) error {
	current, err := os.ReadFile(path)
	if err == nil && bytes.Equal(current, data) {
		return nil
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read existing launcher %s: %w", path, err)
	}

	dir := filepath.Dir(path)
	temporary, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-")
	if err != nil {
		return fmt.Errorf("create temporary launcher for %s: %w", path, err)
	}
	temporaryPath := temporary.Name()
	defer func() {
		_ = os.Remove(temporaryPath)
	}()

	if err := temporary.Chmod(0o644); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("set permissions on %s: %w", temporaryPath, err)
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write temporary launcher %s: %w", temporaryPath, err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync temporary launcher %s: %w", temporaryPath, err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary launcher %s: %w", temporaryPath, err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace launcher %s: %w", path, err)
	}

	directory, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("open launcher directory %s: %w", dir, err)
	}
	if err := directory.Sync(); err != nil {
		_ = directory.Close()
		return fmt.Errorf("sync launcher directory %s: %w", dir, err)
	}
	if err := directory.Close(); err != nil {
		return fmt.Errorf("close launcher directory %s: %w", dir, err)
	}
	return nil
}

func compactInstallableName(name string) string {
	name = strings.TrimSpace(name)
	for _, separator := range installableNameSeparators {
		if shortName, _, found := strings.Cut(name, separator); found {
			name = shortName
			break
		}
	}

	maxNameRunes := maxInstallableNameRunes - len([]rune(installableNamePrefix))
	nameRunes := []rune(name)
	if len(nameRunes) > maxNameRunes {
		const suffix = "..."
		nameRunes = append(
			nameRunes[:maxNameRunes-len([]rune(suffix))],
			[]rune(suffix)...,
		)
	}

	return installableNamePrefix + string(nameRunes)
}
