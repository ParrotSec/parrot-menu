package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

type launcher struct {
	name      string
	path      string
	filename  string
	deb       string
	sandboxed bool
	todelete  bool
	toupdate  bool
}

func main() {
	if uid, _ := user.Current(); uid.Uid != "0" {
		os.Stderr.WriteString(fmt.Sprintf("Wrong user %s! - Run this program as root\n", uid.Name))
		os.Exit(1)
	}

	fmt.Println("Scanning application launchers")

	// get list of files in /usr/share/parrot-menu/applications (source)""
	sfiles, err := ioutil.ReadDir("/usr/share/parrot-menu/applications/")
	if err != nil {
		log.Fatal(err)
	}

	// get list of files in /usr/share/applications (destination)
	dfiles, err := ioutil.ReadDir("/usr/share/applications/")
	if err != nil {
		log.Fatal(err)
	}

	// get list of installed packages
	tmp, _ := exec.Command("bash", "-c", "apt list --installed | cut -d'/' -f1").Output()
	packages := strings.Split(string(tmp), "\n")

	// create source and destination structures
	source := make([]launcher, len(sfiles))
	destination := make([]launcher, len(dfiles))

	// initialize source structure
	var spath string
	for i, f := range sfiles {
		spath = fmt.Sprintf("/usr/share/parrot-menu/applications/%s", f.Name())
		getLauncherContent(&source[i], spath, f.Name())
	}

	// initialize destination structure
	var dpath string
	for i, f := range dfiles {
		dpath = fmt.Sprintf("/usr/share/applications/%s", f.Name())
		getLauncherContent(&destination[i], dpath, f.Name())
	}

	fmt.Println("Updating active launchers")

	removeOldLaunchers(destination)
	copyActiveLaunchers(source, packages)
	exec.Command("sync")

	fmt.Println("Done")
}

func getLauncherContent(s *launcher, path string, filename string) {
	// set path
	s.path = path

	// set filename
	s.filename = filename

	// set name
	tmpname, _ := exec.Command("bash", "-c",
		fmt.Sprintf("grep Name= %s | sed -e 's/Name=//g'", path)).Output()
	s.name = strings.Trim(string(tmpname), "\n")

	// set package name
	tmpdeb, _ := exec.Command("bash", "-c",
		fmt.Sprintf("grep X-Parrot-Package= %s | sed -e 's/X-Parrot-Package=//g'", path)).Output()
	s.deb = strings.Trim(string(tmpdeb), "\n")
}

func removeOldLaunchers(files []launcher) {
	for _, f := range files {
		if f.deb != "" {
			os.Remove(f.path)
		}
	}
}

func copyLauncher(source string, destination string) (error, error, error) {
	in, errOpen := os.Open(source)
	if errOpen != nil {
		fmt.Printf("[WARNING] Can't open: %s - %v\n", source, errOpen)
	}

	out, errClose := os.Create(destination)
	if errClose != nil {
		fmt.Printf("[WARNING] Can't create: %s - %v\n", destination, errClose)
	}

	_, errCopy := io.Copy(out, in)
	if errCopy != nil {
		fmt.Printf("[WARNING] Can't install launcher - %v\n", errCopy)
	}

	defer in.Close()
	defer out.Close()
	return errOpen, errClose, errCopy
}

func copyActiveLaunchers(launchers []launcher, packages []string) {
	for _, l := range launchers {
		for _, p := range packages {
			if l.deb == p {
				copyLauncher(l.path, fmt.Sprintf("/usr/share/applications/%s", l.filename))
			}
		}
	}
}
