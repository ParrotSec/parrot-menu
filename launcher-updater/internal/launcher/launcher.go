package launcher

import (
	"launcher-updater/internal"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func RemoveOldLaunchers() {
	err := filepath.WalkDir(internal.DirLauncherDest, func(path string, d os.DirEntry, err error) error {
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
			srcToCheck := filepath.Join(internal.DirLauncherSource, currentLauncher)
			if _, err := os.Stat(srcToCheck); os.IsNotExist(err) {
				if err := os.Remove(path); err != nil {
					log.Printf("Failed to remove %s: %v", path, err)
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("Error walking directory %s: %v", internal.DirLauncherDest, err)
	}
}
