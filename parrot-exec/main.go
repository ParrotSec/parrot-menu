/*
parrot-exec is the execution gateway for ParrotOS desktop entries.

It dispatches into four modes:

  - `--sudo`: runs a command inside the terminal the DE has already
    opened (Terminal=true in the .desktop file).

  - `--ls`: delegates to ls rather than os.ReadDir for visual consistency
    with terminal output.

  - `--gui`: runs via pkexec while preserving DISPLAY and XAUTHORITY, which
    pkexec strips for security reasons by default.

  - `--install`: runs apt update + apt install, then triggers launcher-updater
    so the template desktop entry is replaced with the real one.

The `--keep` flag (default set to **true**) calls runShell() after execution so
the terminal stays open otherwise a Terminal=true entry would close before the
user can read output or errors. runShell() whitelists known shells to prevent
executing untrusted binaries injected via $SHELL.
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	banner = `
 ___                  _   ___         
| _ \__ _ _ _ _ _ ___| |_/ __| ___ __ 
|  _/ _' | '_| '_/ _ \  _\__ \/ -_) _|
|_| \__,_|_| |_| \___/\__|___/\___\__|
`
	colorReset   = "\033[0m"
	colorRed     = "\033[0;31m"
	colorMagenta = "\033[1;95m"
	colorCyan    = "\033[1;96m"
	parrotEmail  = "team@parrotsec.org"
)

func main() {
	isSudo := flag.Bool("sudo", false, "Run with sudo")
	isGui := flag.Bool("gui", false, "Run with pkexec and show notifications")
	isLs := flag.Bool("ls", false, "Run as directory lister")
	isInstall := flag.Bool("install", false, "Install the specified package")
	noBanner := flag.Bool("no-banner", false, "Do not show banner")
	keepOpen := flag.Bool("keep", true, "Keep shell open after execution")

	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		fmt.Println("Usage: parrot-exec [flags] <command>")
		os.Exit(1)
	}

	if !*noBanner && !*isGui {
		fmt.Print(banner)
	}

	commandStr := strings.Join(args, " ")
	execName := args[0]

	if !*isLs {
		if _, err := exec.LookPath(execName); err != nil {
			handleError(execName, *isGui)
			return
		}
	}

	if *isGui {
		runGui(commandStr, args)
	} else if *isLs {
		runLs(commandStr, *keepOpen)
	} else if *isInstall {
		runInstall(execName, *keepOpen)
	} else if *isSudo {
		runCommand(args, true, *keepOpen)
	} else {
		runCommand(args, false, *keepOpen)
	}
}

func runInstall(pkgName string, keep bool) {
	fmt.Printf("%sInstalling package %s...%s\n\n", colorCyan, pkgName, colorReset)

	cmd := exec.Command("apt-cache", "show", pkgName)
	if cmd.Run() != nil {
		cmd = exec.Command("sudo", "apt", "update")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			fmt.Printf("\n%sERROR:%s Failed to update package list: %v\n\n", colorRed, colorReset, err)
			if keep { runShell() }
			return
		}
	}

	cmd = exec.Command("sudo", "apt", "install", "-y", pkgName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("\n%sERROR:%s Failed to install '%s': %v\n\n", colorRed, colorReset, pkgName, err)
	} else {
		fmt.Printf("\n%sSUCCESS:%s '%s' installed correctly. The menu will now be updated.\n\n", colorCyan, colorReset, pkgName)
		updateCmd := exec.Command("sudo", "launcher-updater")
		updateCmd.Stdout = os.Stdout
		updateCmd.Stderr = os.Stderr
		if err := updateCmd.Run(); err != nil {
			fmt.Printf("\n%sWARNING:%s Menu update failed: %v\n", colorRed, colorReset, err)
		}
	}

	if keep {
		runShell()
	}
}

func handleError(name string, gui bool) {
	msg := fmt.Sprintf("Command '%s' cannot be found.\nPlease report this bug to %s%s%s", name, colorCyan, parrotEmail, colorReset)
	if gui {
		exec.Command("notify-send", "-i", "security-low", "Execution Failed", msg).Run()
	} else {
		fmt.Printf("%sERROR:%s %s\n\n", colorRed, colorReset, msg)
		runShell()
	}
}

func runGui(commandStr string, args []string) {
	exec.Command("notify-send", "ParrotSec", "Starting "+commandStr).Run()

	fullArgs := append([]string{"env", "DISPLAY=" + os.Getenv("DISPLAY"), "XAUTHORITY=" + os.Getenv("XAUTHORITY")}, args...)
	cmd := exec.Command("pkexec", fullArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		handleError(commandStr, true)
	}
}

func runCommand(args []string, sudo bool, keep bool) {
	fmt.Printf("Executing %s%s%s\n", colorMagenta, strings.Join(args, " "), colorReset)

	var cmd *exec.Cmd
	if sudo {
		cmd = exec.Command("sudo", args...)
	} else {
		cmd = exec.Command(args[0], args[1:]...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		handleError(strings.Join(args, " "), false)
	}

	if keep {
		runShell()
	}
}

func runLs(path string, keep bool) {
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		fmt.Printf("%sPath '%s' doesn't exist.%s\nPlease report this bug to %s%s%s\n", colorMagenta, path, colorReset, colorCyan, parrotEmail, colorReset)
		if keep {
			runShell()
		}
		return
	}

	fmt.Printf("Listing %s%s%s\n", colorMagenta, path, colorReset)
	cmd := exec.Command("ls", "-laH", "--color=auto", "--", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("%sERROR:%s Failed to list directory: %v\n", colorRed, colorReset, err)
	}

	if keep {
		runShell()
	}
}

var allowedShells = map[string]bool{
	"/bin/bash":     true,
	"/bin/zsh":      true,
	"/bin/sh":       true,
	"/bin/fish":     true,
	"/usr/bin/bash": true,
	"/usr/bin/zsh":  true,
	"/usr/bin/fish": true,
}

func runShell() {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}
	if !allowedShells[shell] {
		fmt.Printf("%sWARNING:%s SHELL '%s' is not recognized, falling back to /bin/bash\n", colorRed, colorReset, shell)
		shell = "/bin/bash"
	}
	cmd := exec.Command(shell, "-i")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Printf("%sERROR:%s Could not start shell %s: %v\n", colorRed, colorReset, shell, err)
		os.Exit(1)
	}
}
