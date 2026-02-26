package blacklist

import (
	"launcher-updater/internal/desktop"
	"log/slog"
	"os"
	"path/filepath"
)

func FixDebLaunchers() {
	blacklistLauncherName := []string{
		"org.radare.Cutter.desktop", "gpa.desktop", "rtlsdr-scanner.desktop",
		"gnuradio-grc.desktop", "arduino.desktop", "gqrx.desktop",
		"zulucrypt-gui.desktop", "zulumount-gui.desktop", "ophcrack.desktop",
		"xsser.desktop", "io.github.mhogomchungu.sirikali.desktop", "etherape.desktop",
		"edb.desktop", "lynis.desktop", "wireshark.desktop",
		"org.wireshark.Wireshark.desktop", "ettercap.desktop", "chirp.desktop",
		"hydra-gtk.desktop", "driftnet.desktop", "gscriptor.desktop",
		"spectool_gtk.desktop", "gksu.desktop", "re.rizin.cutter.desktop",
		"openjdk-8-policytool.desktop", "org.keepassxc.KeePassXC.desktop",
	}

	for _, fileName := range blacklistLauncherName {
		finalPath := filepath.Join(desktop.DirLauncherDest, fileName)
		// If a file from the blacklist exists, remove it.
		if _, err := os.Stat(finalPath); err == nil {
			if err := os.Remove(finalPath); err != nil {
				slog.Error("error while removing", "path", finalPath, "err", err)
			}
		}
	}
}
