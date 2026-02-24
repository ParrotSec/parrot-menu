package main

import (
	"fmt"
	"launcher-updater/internal"
	"launcher-updater/internal/blacklist"
	"launcher-updater/internal/desktop"
	"launcher-updater/internal/dpkg"
	"launcher-updater/internal/launcher"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func main() {
	// It starts the update and cleanup operations in parallel.
	// We are using goroutines to improve performance.
	var wg sync.WaitGroup

	fmt.Println("--------------------------------------------------")
	fmt.Println("[!] Scanning application launchers")
	updateLaunchers()

	fmt.Println("Removing duplicate or broken launchers...")
	wg.Add(2)
	go func() {
		defer wg.Done()
		launcher.RemoveOldLaunchers()
	}()
	go func() {
		defer wg.Done()
		blacklist.FixDebLaunchers()
	}()

	wg.Wait()

	fmt.Println("[!] Launchers have been successfully updated!")
	fmt.Println("--------------------------------------------------")
}

func updateLaunchers() {
	installed, err := dpkg.QueryInstalled()
	if err != nil {
		log.Fatalf("Fatal error querying installed packages: %v", err)
	}

	err = filepath.WalkDir(internal.DirLauncherSource, func(path string, d os.DirEntry, err error) error {
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

		aptParrotPackage, err := desktop.GetXPackageName(path)
		if err != nil {
			log.Printf("Error reading package name from %s: %v", path, err)
			return nil
		}
		if aptParrotPackage == "" {
			return nil
		}

		finalDestPath := filepath.Join(internal.DirLauncherDest, fileName)
		_, isInstalled := installed[aptParrotPackage]

		if isInstalled {
			srcInfo, err := d.Info()
			if err != nil {
				log.Printf("Could not get source file info for %s: %v", path, err)
				return nil
			}
			destInfo, destErr := os.Stat(finalDestPath)
			if destErr != nil || srcInfo.Size() != destInfo.Size() || srcInfo.ModTime() != destInfo.ModTime() {
				if err := desktop.CopyFile(path, finalDestPath); err != nil {
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
		desktop.FixOldLaunchers(fileName)
		return nil
	})

	if err != nil {
		log.Printf("Error walking source directory %s: %v", internal.DirLauncherSource, err)
	}
}
