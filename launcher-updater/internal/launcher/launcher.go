package launcher

import (
	"fmt"
	"launcher-updater/internal/desktop"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const dirLauncherSource = "/usr/share/parrot-menu/applications/"

var validPackageName = regexp.MustCompile(`^[a-z0-9][a-z0-9+.-]*$`)

func CatalogPackageNames() ([]string, error) {
	packageSet := make(map[string]struct{})
	err := filepath.WalkDir(
		dirLauncherSource,
		func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !isManaged(d.Name()) {
				return nil
			}

			pkgName, err := desktop.GetXPackageName(path)
			if err != nil {
				return fmt.Errorf("read package metadata from %s: %w", path, err)
			}
			if pkgName == "" {
				if !desktop.IsManaged(path) {
					return fmt.Errorf(
						"%s must define X-Parrot-Package or X-Parrot-Managed=true",
						path,
					)
				}
				return nil
			}
			if !validPackageName.MatchString(pkgName) {
				return fmt.Errorf("invalid package name %q in %s", pkgName, path)
			}
			packageSet[pkgName] = struct{}{}
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("validate launcher catalog: %w", err)
	}

	packageNames := make([]string, 0, len(packageSet))
	for pkgName := range packageSet {
		packageNames = append(packageNames, pkgName)
	}
	sort.Strings(packageNames)
	return packageNames, nil
}

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
