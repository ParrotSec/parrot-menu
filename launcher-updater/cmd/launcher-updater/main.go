package main

import (
	"fmt"
	"launcher-updater/internal/desktop"
	"launcher-updater/internal/dpkg"
	"launcher-updater/internal/launcher"
	"log/slog"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"
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

	total, notInstalled := launcher.SyncLaunchers(installed)

	fmt.Println("Removing duplicate or broken launchers...")
	launcher.RemoveOldLaunchers()
	launcher.FixDebLaunchers()

	fmt.Printf("[i] %d launcher(s) processed, %d package(s) not installed\n", total, notInstalled)

	fmt.Println("[!] Launchers have been successfully updated!")
	fmt.Println("--------------------------------------------------")

	if !skipKsycoca {
		// Touch the dest directory so kbuildsycoca detects the mtime change.
		_ = os.Chtimes(desktop.DirLauncherDest, time.Now(), time.Now())

		if _, err := exec.LookPath("kbuildsycoca6"); err == nil {
			userName := os.Getenv("SUDO_USER")
			if userName == "" {
				userName = findUser()
			}
			if userName != "" {
				uid := os.Getenv("SUDO_UID")
				if uid == "" {
					if u, err := user.Lookup(userName); err == nil {
						uid = u.Uid
					}
				}

				// Pass the user's D-Bus session address so kbuildsycoca6
				// can notify Kicker that the cache was rebuilt.
				var args []string
				if uid != "" {
					args = append(args,
						fmt.Sprintf("DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/%s/bus", uid))
				}
				args = append(args, "kbuildsycoca6")

				out, err := exec.Command("sudo", append([]string{"-u", userName}, args...)...).CombinedOutput()
				if err != nil {
					fmt.Printf("[!] kbuildsycoca6 error: %v\n%s\n", err, string(out))
				}
			} else {
				fmt.Println("[!] Could not determine user to run kbuildsycoca6")
			}
		} else {
			fmt.Println("[!] kbuildsycoca6 not found on the system")
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
