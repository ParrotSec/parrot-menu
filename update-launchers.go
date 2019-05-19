package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
	fmt.Println("Scanning application launchers")

	// get list of files in /usr/share/parrot-menu/applications (source)
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
	fmt.Println(len(packages))

	// create source and destination structures
	source := make([]launcher, len(sfiles))
	destination := make([]launcher, len(dfiles))

	// initialize source structure
	sfinished := make(chan bool)
	var spath string
	for i, f := range sfiles {
		spath = fmt.Sprintf("/usr/share/parrot-menu/applications/%s", f.Name())
		go getLauncherContent(sfinished, &source[i], spath, f.Name())
	}

	// initialize destination structure
	dfinished := make(chan bool)
	var dpath string
	for i, f := range dfiles {
		dpath = fmt.Sprintf("/usr/share/applications/%s", f.Name())
		go getLauncherContent(dfinished, &destination[i], dpath, f.Name())
	}

	// wait for initialization goroutines to finish
	for i := 0; i < len(dfiles); i++ {
		<-dfinished
	}
	for i := 0; i < len(sfiles); i++ {
		<-sfinished
	}

	fmt.Println("Updating active launchers")

	removeOldLaunchers(destination)
	copyActiveLaunchers(source, packages)
	exec.Command("sync")

	fmt.Println("Done")
}

func getLauncherContent(finished chan bool, s *launcher, path string, filename string) {
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

	// send signal
	finished <- true
}

func removeOldLaunchers(files []launcher) {
	for _, f := range files {
		if f.deb != "" {
			os.Remove(f.path)
		}
	}
}

func copyActiveLaunchers(launchers []launcher, packages []string) {
	for _, l := range launchers {
		for _, p := range packages {
			if l.deb == p {
				in, err := os.Open(l.path)
				if err != nil {
					fmt.Printf("[WARNING] Can't find source launcher to copy: %s\n", l.path)
				}
				defer in.Close()

				out, err := os.Create(fmt.Sprintf("/usr/share/applications/%s", l.filename))
				if err != nil {
					fmt.Printf("[WARNING] Can't create target launcher to write: %s%s\n", "/usr/share/applications/", l.filename)
				}
				defer out.Close()

				if _, err := io.Copy(out, in); err != nil {
					fmt.Printf("[WARNING] Can't create new launcher for %s\n", l.name)
				}
			}
		}
	}
}

