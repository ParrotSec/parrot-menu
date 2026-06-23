package launcher

import (
	"launcher-updater/internal/desktop"
	"log/slog"
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

		if !isManaged(d.Name()) {
			return nil
		}

		if !desktop.IsManaged(path) {
			return nil
		}

		srcToCheck := filepath.Join(dirLauncherSource, d.Name())
		if _, err := os.Stat(srcToCheck); os.IsNotExist(err) {
			if err := os.Remove(path); err != nil {
				slog.Error("failed to remove", "path", path, "err", err)
			}
		}
		return nil
	})

	if err != nil {
		slog.Error("failed to walk source directory", "DirLauncherDest", desktop.DirLauncherDest, "err", err)
	}
}

func FixDebLaunchers() {
	blacklist := map[string]string{
		"wireshark.desktop":               "parrot-wireshark.desktop",
		"org.wireshark.Wireshark.desktop": "parrot-wireshark.desktop",
		"ettercap.desktop":                "parrot-ettercap-graphical.desktop",
		"chirp.desktop":                   "parrot-chirp.desktop",
		"driftnet.desktop":                "parrot-driftnet.desktop",
		"lynis.desktop":                   "parrot-lynis.desktop",
		"xsser.desktop":                   "parrot-xsser.desktop",
		"etherape.desktop":                "parrot-etherape.desktop",
		"ophcrack.desktop":                "parrot-ophcrack.desktop",
		"gqrx.desktop":                    "parrot-gqrx.desktop",
		"gpa.desktop":                     "parrot-gpa.desktop",
		"arduino.desktop":                 "parrot-arduino.desktop",
		"rtlsdr-scanner.desktop":          "parrot-rtlsdr-scanner.desktop",
		"org.radare.Cutter.desktop":       "parrot-rizin-cutter.desktop",
		"re.rizin.cutter.desktop":         "parrot-rizin-cutter.desktop",
	}

	for origName, wrapperName := range blacklist {
		origPath := filepath.Join(desktop.DirLauncherDest, origName)
		if _, err := os.Stat(origPath); os.IsNotExist(err) {
			continue
		}

		if desktop.IsManaged(origPath) {
			continue
		}

		wrapperPath := filepath.Join(desktop.DirLauncherDest, wrapperName)
		if _, err := os.Stat(wrapperPath); os.IsNotExist(err) {
			continue
		}

		if err := os.Remove(origPath); err != nil {
			slog.Error("failed to remove blacklisted launcher", "path", origPath, "err", err)
		}
	}
}

func SyncLaunchers(installed map[string]struct{}) (int, int) {
	var total, notInstalled int

	err := filepath.WalkDir(dirLauncherSource, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		if !isManaged(d.Name()) {
			return nil
		}

		total++
		if !syncSingleLauncher(path, d, installed) {
			notInstalled++
		}
		return nil
	})

	if err != nil {
		slog.Error("failed to walk source directory",
			"dirLauncherSource", dirLauncherSource, "err", err)
	}

	return total, notInstalled
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

func syncSingleLauncher(srcPath string, d os.DirEntry, installed map[string]struct{}) bool {
	pkgName, err := desktop.GetXPackageName(srcPath)
	if err != nil || pkgName == "" {
		fileName := d.Name()
		destPath := filepath.Join(desktop.DirLauncherDest, fileName)
		ensureLauncherUpdated(srcPath, destPath, d)
		desktop.FixOldLaunchers(fileName)
		return true
	}

	fileName := d.Name()
	destPath := filepath.Join(desktop.DirLauncherDest, fileName)

	if _, ok := installed[pkgName]; ok {
		ensureLauncherUpdated(srcPath, destPath, d)
		desktop.FixOldLaunchers(fileName)
		return true
	}

	ensureLauncherTemplate(srcPath, destPath, pkgName, d)
	desktop.FixOldLaunchers(fileName)
	return false
}

func ensureLauncherUpdated(srcPath, destPath string, d os.DirEntry) {
	if err := desktop.CopyFile(srcPath, destPath); err != nil {
		slog.Error("failed to copy source path to destination path",
			"srcPath", srcPath, "destPath", destPath, "err", err)
	}
}

func ensureLauncherTemplate(srcPath, destPath, pkgName string, d os.DirEntry) {
	if err := desktop.CopyTemplateLauncher(srcPath, destPath, pkgName); err != nil {
		slog.Error("failed to create template launcher", "destPath", destPath, "err", err)
	}
}
