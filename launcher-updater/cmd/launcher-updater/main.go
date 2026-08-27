package main

import (
	"fmt"
	"launcher-updater/internal/config"
	"launcher-updater/internal/dpkg"
	"launcher-updater/internal/launcher"
	"launcher-updater/internal/runlock"
	"log/slog"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

func main() {
	mode := ""
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	skipKsycoca := mode == "wait_dpkg" || mode == "remove_all"

	lock, err := runlock.Acquire()
	if err != nil {
		slog.Error("error acquiring launcher updater lock", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err := lock.Release(); err != nil {
			slog.Error("error releasing launcher updater lock", "err", err)
		}
	}()

	fmt.Println("--------------------------------------------------")
	if mode == "remove_all" {
		fmt.Println("[!] Removing Parrot application launchers")
		if err := launcher.RemoveAllManagedLaunchers(); err != nil {
			slog.Error("error removing managed launchers", "err", err)
			os.Exit(1)
		}
		fmt.Println("[!] Launchers have been successfully removed!")
		fmt.Println("--------------------------------------------------")
		return
	}

	fmt.Println("[!] Scanning application launchers")

	installed, err := dpkg.QueryInstalled()
	if err != nil {
		slog.Error("error querying installed packages", "err", err)
		os.Exit(1)
	}

	options, err := config.Load()
	if err != nil {
		slog.Error("error loading launcher configuration", "err", err)
		os.Exit(1)
	}

	if _, err := launcher.CatalogPackageNames(); err != nil {
		slog.Error("error validating launcher catalog", "err", err)
		os.Exit(1)
	}

	total, notInstalled, err := launcher.SyncLaunchers(
		installed,
		options.ShowInstallableLaunchers,
	)
	if err != nil {
		slog.Error("error synchronizing launchers", "err", err)
		os.Exit(1)
	}

	fmt.Println("Removing duplicate or broken launchers...")
	launcher.RemoveOldLaunchers()
	launcher.FixDebLaunchers()

	fmt.Printf("[i] %d launcher(s) processed, %d launcher(s) for uninstalled packages\n", total, notInstalled)

	fmt.Println("[!] Launchers have been successfully updated!")
	fmt.Println("--------------------------------------------------")

	if !skipKsycoca {
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
