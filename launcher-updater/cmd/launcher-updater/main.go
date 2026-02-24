package main

import (
	"fmt"
	"launcher-updater/internal/blacklist"
	"launcher-updater/internal/dpkg"
	"launcher-updater/internal/launcher"
	"log"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	fmt.Println("--------------------------------------------------")
	fmt.Println("[!] Scanning application launchers")

	installed, err := dpkg.QueryInstalled()
	if err != nil {
		log.Fatalf("Fatal error querying installed packages: %v", err)
	}

	launcher.SyncLaunchers(installed)

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
