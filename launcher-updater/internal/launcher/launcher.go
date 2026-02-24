package launcher

import (
	"launcher-updater/internal/desktop"
	"log"
	"os"
	"path/filepath"
	"strings"
)

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

		if (strings.HasPrefix(currentLauncher, "parrot-") ||
			strings.HasPrefix(currentLauncher, "serv-")) &&
			strings.HasSuffix(currentLauncher, ".desktop") {
			// Build the path to the corresponding file in the source directory.
			srcToCheck := filepath.Join(dirLauncherSource, currentLauncher)
			if _, err := os.Stat(srcToCheck); os.IsNotExist(err) {
				if err := os.Remove(path); err != nil {
					log.Printf("Failed to remove %s: %v", path, err)
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("Error walking directory %s: %v", desktop.DirLauncherDest, err)
	}
}

func SyncLaunchers(installed map[string]struct{}) {
	err := filepath.WalkDir(dirLauncherSource, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		if !isManaged(d.Name()) {
			return nil
		}

		syncSingleLauncher(path, d, installed)
		return nil
	})

	if err != nil {
		log.Printf("Error walking source directory %s: %v", dirLauncherSource, err)
	}
}

func isManaged(name string) bool {
	return (strings.HasPrefix(name, "parrot-") ||
		strings.HasPrefix(name, "serv-")) &&
		strings.HasSuffix(name, ".desktop")
}

func syncSingleLauncher(srcPath string, d os.DirEntry, installed map[string]struct{}) {
	pkgName, err := desktop.GetXPackageName(srcPath)
	if err != nil || pkgName == "" {
		return
	}

	fileName := d.Name()
	destPath := filepath.Join(desktop.DirLauncherDest, fileName)

	if _, ok := installed[pkgName]; ok {
		ensureLauncherUpdated(srcPath, destPath, d)
	} else {
		ensureLauncherRemoved(destPath)
	}

	desktop.FixOldLaunchers(fileName)
}

func ensureLauncherUpdated(srcPath, destPath string, d os.DirEntry) {
	srcInfo, err := d.Info()
	if err != nil {
		return
	}

	destInfo, err := os.Stat(destPath)
	// Update if it doesn't exist or metadata differs
	if err != nil || srcInfo.Size() != destInfo.Size() || srcInfo.ModTime() != destInfo.ModTime() {
		if err := desktop.CopyFile(srcPath, destPath); err != nil {
			log.Printf("Error copying %s -> %s: %v", srcPath, destPath, err)
		}
	}
}

func ensureLauncherRemoved(destPath string) {
	if _, err := os.Stat(destPath); err == nil {
		if err := os.Remove(destPath); err != nil {
			log.Printf("Error removing old launcher %s: %v", destPath, err)
		}
	}
}
