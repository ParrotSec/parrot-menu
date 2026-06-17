package launcher

import (
	"launcher-updater/internal/desktop"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type RemovedTool struct {
	Name    string
	Package string
}

const dirLauncherSource = "/usr/share/parrot-menu/applications/"

func RemoveOldLaunchers() {
	err := filepath.WalkDir(desktop.DirLauncherDest, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		currentLauncher := d.Name()

		if isManaged(currentLauncher) {
			// Build the path to the corresponding file in the source directory.
			srcToCheck := filepath.Join(dirLauncherSource, currentLauncher)
			if _, err := os.Stat(srcToCheck); os.IsNotExist(err) {
				if err := os.Remove(path); err != nil {
					slog.Error("failed to remove", "path", path, "err", err)
				}
			}
		}
		return nil
	})

	if err != nil {
		slog.Error("failed to walk source directory", "DirLauncherDest", desktop.DirLauncherDest, "err", err)
	}
}

func SyncLaunchers(installed map[string]struct{}) []RemovedTool {
	var removed []RemovedTool

	err := filepath.WalkDir(dirLauncherSource, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		if !isManaged(d.Name()) {
			return nil
		}

		if rt := syncSingleLauncher(path, d, installed); rt != nil {
			removed = append(removed, *rt)
		}
		return nil
	})

	if err != nil {
		slog.Error("failed to walk source directory",
			"dirLauncherSource", dirLauncherSource, "err", err)
	}

	return removed
}

var managedPrefixes = []string{"parrot-", "serv-"}

func isManaged(name string) bool {
	if !strings.HasSuffix(name, ".desktop") {
		return false
	}

	for _, prefix := range managedPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func syncSingleLauncher(srcPath string, d os.DirEntry, installed map[string]struct{}) *RemovedTool {
	pkgName, err := desktop.GetXPackageName(srcPath)
	if err != nil || pkgName == "" {
		return nil
	}

	fileName := d.Name()
	destPath := filepath.Join(desktop.DirLauncherDest, fileName)

	if _, ok := installed[pkgName]; ok {
		ensureLauncherUpdated(srcPath, destPath, d)
		desktop.FixOldLaunchers(fileName)
		return nil
	}

	ensureLauncherTemplate(srcPath, destPath, pkgName, d)
	desktop.FixOldLaunchers(fileName)
	return &RemovedTool{Name: fileName, Package: pkgName}
}

func ensureLauncherUpdated(srcPath, destPath string, d os.DirEntry) {
	srcInfo, err := d.Info()
	if err != nil {
		return
	}

	if err := desktop.CopyFile(srcPath, destPath); err != nil {
		slog.Error("failed to copy source path to destination path",
			"srcPath", srcPath, "destPath", destPath, "err", err)
	} else {
		_ = os.Chtimes(destPath, srcInfo.ModTime(), srcInfo.ModTime())
	}
}

func ensureLauncherTemplate(srcPath, destPath, pkgName string, d os.DirEntry) {
	srcInfo, err := d.Info()
	if err != nil {
		return
	}

	if err := desktop.CopyTemplateLauncher(srcPath, destPath, pkgName); err != nil {
		slog.Error("failed to create template launcher", "destPath", destPath, "err", err)
	} else {
		_ = os.Chtimes(destPath, srcInfo.ModTime(), srcInfo.ModTime())
	}
}
