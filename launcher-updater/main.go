package main

// launcher-updater is a maintenance tool for .desktop files on Parrot OS systems.

// It synchronizes launchers from a Parrot-specific source directory
// to the standard system directory (/usr/share/applications), handling the removal
// of obsolete files, fixing problematic launchers, and ensuring that only launchers
// for actually installed applications are present.

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const (
	dirLauncherSource = "/usr/share/parrot-menu/applications/"
	dirLauncherDest   = "/usr/share/applications/"
	dpkgStatusPath    = "/var/lib/dpkg/status"
)

func checkValidBinary(path string) {
	// Set a standard PATH environment variable to ensure the executable lookup happens in the **correct** directories.
	err := os.Setenv("PATH", "/usr/local/bin:/usr/bin:/bin:/usr/local/games:/usr/games:/usr/local/sbin:/usr/sbin:/sbin")
	if err != nil {
		return
	}

	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("Error closing file %s: %v", path, err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Exec=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) < 2 {
				continue
			}
			execLine := strings.TrimSpace(parts[1])
			execFile := strings.Split(execLine, " ")[0]
			execFile = strings.Trim(execFile, "\"'")

			if _, err := exec.LookPath(execFile); err != nil {
				fmt.Printf(" [-] Missing executable file %s at launcher %s\n", execFile, path)
			}
			return
		}
	}
}

// A launcher is considered obsolete if its name starts with "parrot-" or "serv-"
// but it no longer exists in the source directory.
func removeOldLaunchers() {
	err := filepath.WalkDir(dirLauncherDest, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		shouldCheckBinary := true
		currentLauncher := d.Name()

		if (strings.HasPrefix(currentLauncher, "parrot-") ||
			strings.HasPrefix(currentLauncher, "serv-")) &&
			strings.HasSuffix(currentLauncher, ".desktop") {
			// Build the path to the corresponding file in the source directory.
			srcToCheck := filepath.Join(dirLauncherSource, currentLauncher)
			if _, err := os.Stat(srcToCheck); os.IsNotExist(err) {
				shouldCheckBinary = false
				if err := os.Remove(path); err != nil {
					log.Printf("Failed to remove %s: %v", path, err)
				}
			}
		}

		if shouldCheckBinary {
			checkValidBinary(path)
		}
		return nil
	})

	if err != nil {
		log.Printf("Error walking directory %s: %v", dirLauncherDest, err)
	}
}

func fixDebLaunchers() {
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
		finalPath := filepath.Join(dirLauncherDest, fileName)
		// If a file from the blacklist exists, remove it.
		if _, err := os.Stat(finalPath); err == nil {
			if err := os.Remove(finalPath); err != nil {
				log.Printf("Error while removing %s: %v", finalPath, err)
			}
		}
	}
}

func getXPackageName(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("Error closing file %s: %v", path, err)
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
	}
	return "", nil
}

func queryInstalled() (map[string]struct{}, error) {
	installed := make(map[string]struct{})
	file, err := os.Open(dpkgStatusPath)
	if err != nil {
		return nil, fmt.Errorf("could not open dpkg status file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("Error closing file %s: %v", dpkgStatusPath, err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	var pkgName string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Package: ") {
			pkgName = strings.TrimPrefix(line, "Package: ")
		} else if line == "Status: install ok installed" && pkgName != "" {
			installed[pkgName] = struct{}{}
			pkgName = ""
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading dpkg status file: %w", err)
	}
	return installed, nil
}

// If a new launcher (e.g., "serv-tool.desktop") is installed, this function
// ensures that the older version (e.g., "parrot-toolname.desktop") is removed
// to avoid duplicates in the application menu.
func fixOldLaunchers(fileName string) {
	newNamePrefixes := []string{"serv-"}
	for _, checkName := range newNamePrefixes {
		if strings.HasPrefix(fileName, checkName) {
			oldFileName := "parrot-" + strings.TrimPrefix(fileName, checkName)
			destPath := filepath.Join(dirLauncherDest, oldFileName)
			if _, err := os.Stat(destPath); err == nil {
				if err := os.Remove(destPath); err != nil {
					log.Printf("Error while removing duplicate launcher %s: %v", destPath, err)
				}
			}
			break
		}
	}
}

// updateLaunchers is the main function that synchronizes launchers.
// 1. Gets the list of installed packages.
// 2. Scans the Parrot source directory for launchers.
// 3. For each launcher, it determines the associated Debian package via the "X-Parrot-" field.
// 4. If the package is installed, it copies/updates the launcher to the destination directory.
// 5. If the package is not installed, it removes the launcher from the destination.
func updateLaunchers() {
	installed, err := queryInstalled()
	if err != nil {
		log.Fatalf("Fatal error querying installed packages: %v", err)
	}

	err = filepath.WalkDir(dirLauncherSource, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		fileName := d.Name()
		if !strings.HasPrefix(fileName, "parrot-") && !strings.HasPrefix(fileName, "serv-") {
			return nil
		}

		aptParrotPackage, err := getXPackageName(path)
		if err != nil {
			log.Printf("Error reading package name from %s: %v", path, err)
			return nil
		}
		if aptParrotPackage == "" {
			return nil
		}

		finalDestPath := filepath.Join(dirLauncherDest, fileName)
		_, isInstalled := installed[aptParrotPackage]

		if isInstalled {
			srcInfo, err := d.Info()
			if err != nil {
				log.Printf("Could not get source file info for %s: %v", path, err)
				return nil
			}
			destInfo, err := os.Stat(finalDestPath)
			if os.IsNotExist(err) || srcInfo.Size() != destInfo.Size() || srcInfo.ModTime() != destInfo.ModTime() {
				if err := copyFile(path, finalDestPath); err != nil {
					log.Printf("Error updating file %s: %v", path, err)
				}
			}
		} else {
			if _, err := os.Stat(finalDestPath); err == nil {
				if err := os.Remove(finalDestPath); err != nil {
					log.Printf("Error removing old launcher %s: %v", finalDestPath, err)
				}
			}
		}
		fixOldLaunchers(fileName)
		return nil
	})

	if err != nil {
		log.Printf("Error walking source directory %s: %v", dirLauncherSource, err)
	}
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(source *os.File) {
		err := source.Close()
		if err != nil {
			log.Printf("Error closing file %s: %v", src, err)
		}
	}(source)

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(destination *os.File) {
		err := destination.Close()
		if err != nil {
			log.Printf("Error closing file %s: %v", dst, err)
		}
	}(destination)
	_, err = io.Copy(destination, source)
	return err
}

func main() {
	// It starts the update and cleanup operations in parallel.
	// We are using goroutines to improve performance.
	var wg sync.WaitGroup

	fmt.Println("Scanning application launchers")
	wg.Add(1)
	go func() {
		defer wg.Done()
		updateLaunchers()
	}()

	fmt.Println("Removing duplicate or broken launchers")
	wg.Add(2)
	go func() {
		defer wg.Done()
		removeOldLaunchers()
	}()
	go func() {
		defer wg.Done()
		fixDebLaunchers()
	}()

	wg.Wait()

	fmt.Println("Launchers are updated")
}
