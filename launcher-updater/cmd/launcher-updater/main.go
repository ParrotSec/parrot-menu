package main

import (
	"fmt"
	"launcher-updater/internal/blacklist"
	"launcher-updater/internal/dpkg"
	"launcher-updater/internal/launcher"
	"log/slog"
	"os"
	"os/exec"
	"strings"
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
			if user == "" {
				user = findUser()
			}
			if user != "" {
				_ = exec.Command("sudo", "-u", user, "kbuildsycoca6").Run()
			}
		}
	}
}

func findUser() string {
	out, err := exec.Command("logname").Output()
	if err == nil {
		user := strings.TrimSpace(string(out))
		if user != "" && user != "root" {
			return user
		}
	}
	for _, candidate := range []string{"USER", "USERNAME"} {
		if user := os.Getenv(candidate); user != "" && user != "root" {
			return user
		}
	}
	return ""
}
