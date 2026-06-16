package main

import (
	"fmt"
	"launcher-updater/internal/blacklist"
	"launcher-updater/internal/dpkg"
	"launcher-updater/internal/launcher"
	"log/slog"
	"os"
	"os/exec"
)

func main() {
	skipKsycoca := len(os.Args) > 1 && os.Args[1] == "wait_dpkg"

	fmt.Println("--------------------------------------------------")
	fmt.Println("[!] Scanning application launchers")

	installed, err := dpkg.QueryInstalled()
	if err != nil {
		slog.Error("error querying installed packages", "err", err)
		os.Exit(1)
	}

	removed := launcher.SyncLaunchers(installed)

	fmt.Println("Removing duplicate or broken launchers...")
	launcher.RemoveOldLaunchers()
	blacklist.FixDebLaunchers()

	if len(removed) > 0 {
		fmt.Println()
		fmt.Println("[i] The following tools were uninstalled and their launchers have been removed:")
		for _, rt := range removed {
			fmt.Printf("  - %s → sudo apt install %s\n", rt.Name, rt.Package)
		}
		fmt.Println()
	}

	fmt.Println("[!] Launchers have been successfully updated!")
	fmt.Println("--------------------------------------------------")

	if !skipKsycoca {
		if _, err := exec.LookPath("kbuildsycoca6"); err == nil {
			user := os.Getenv("SUDO_USER")
			if user != "" {
				_ = exec.Command("sudo", "-u", user, "kbuildsycoca6").Run()
			}
		}
	}
}
