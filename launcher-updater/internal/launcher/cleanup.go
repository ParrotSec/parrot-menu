package launcher

import (
	"launcher-updater/internal/desktop"
	"log/slog"
	"os"
	"path/filepath"
)

func RemoveOldLaunchers() {
	err := filepath.WalkDir(
		desktop.DirLauncherDest,
		func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !isManaged(d.Name()) || !desktop.IsManaged(path) {
				return nil
			}

			srcToCheck := filepath.Join(dirLauncherSource, d.Name())
			if _, err := os.Stat(srcToCheck); os.IsNotExist(err) {
				if err := os.Remove(path); err != nil {
					slog.Error("failed to remove", "path", path, "err", err)
				}
			}
			return nil
		},
	)

	if err != nil {
		slog.Error(
			"failed to walk source directory",
			"DirLauncherDest",
			desktop.DirLauncherDest,
			"err",
			err,
		)
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
			slog.Error(
				"failed to remove blacklisted launcher",
				"path",
				origPath,
				"err",
				err,
			)
		}
	}
}
