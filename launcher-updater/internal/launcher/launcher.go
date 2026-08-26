package launcher

import (
	"fmt"
	"launcher-updater/internal/desktop"
	"os"
	"path/filepath"
	"strings"
)

const dirLauncherSource = "/usr/share/parrot-menu/applications/"

func SyncLaunchers(
	installed map[string]struct{},
	showInstallable bool,
) (int, int, error) {
	return syncLaunchers(
		dirLauncherSource,
		desktop.DirLauncherDest,
		installed,
		showInstallable,
	)
}

func syncLaunchers(
	sourceDir string,
	destDir string,
	installed map[string]struct{},
	showInstallable bool,
) (int, int, error) {
	var total, notInstalled int

	err := filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		if !isManaged(d.Name()) {
			return nil
		}

		total++
		isInstalled, err := syncSingleLauncher(
			path,
			d.Name(),
			destDir,
			installed,
			showInstallable,
		)
		if err != nil {
			return err
		}
		if !isInstalled {
			notInstalled++
		}
		return nil
	})

	if err != nil {
		return total, notInstalled, fmt.Errorf(
			"sync launchers from %s: %w", sourceDir, err)
	}

	return total, notInstalled, nil
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

func syncSingleLauncher(
	srcPath string,
	fileName string,
	destDir string,
	installed map[string]struct{},
	showInstallable bool,
) (bool, error) {
	pkgName, err := desktop.GetXPackageName(srcPath)
	if err != nil {
		return false, fmt.Errorf("read package metadata from %s: %w", srcPath, err)
	}

	destPath := filepath.Join(destDir, fileName)
	if pkgName == "" {
		if err := ensureLauncherUpdated(srcPath, destPath); err != nil {
			return false, err
		}
		desktop.FixOldLaunchers(fileName)
		return true, nil
	}

	if _, ok := installed[pkgName]; ok {
		if err := ensureLauncherUpdated(srcPath, destPath); err != nil {
			return false, err
		}
		desktop.FixOldLaunchers(fileName)
		return true, nil
	}

	if showInstallable {
		err = ensureLauncherTemplate(srcPath, destPath, pkgName)
	} else {
		err = removeManagedLauncher(destPath)
	}
	if err != nil {
		return false, err
	}

	desktop.FixOldLaunchers(fileName)
	return false, nil
}

func ensureLauncherUpdated(srcPath, destPath string) error {
	if err := desktop.CopyFile(srcPath, destPath); err != nil {
		return fmt.Errorf("copy launcher %s to %s: %w", srcPath, destPath, err)
	}
	return nil
}

func ensureLauncherTemplate(srcPath, destPath, pkgName string) error {
	if err := desktop.CopyTemplateLauncher(srcPath, destPath, pkgName); err != nil {
		return fmt.Errorf("create installable launcher %s: %w", destPath, err)
	}
	return nil
}

func removeManagedLauncher(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("inspect launcher %s: %w", path, err)
	}

	if !desktop.IsManaged(path) {
		return nil
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove hidden launcher %s: %w", path, err)
	}
	return nil
}
