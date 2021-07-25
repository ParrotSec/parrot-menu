# Update the lauchers in /usr/share/parrot-menu/applications
# Copy launchers from parrot-menu (/usr/share/parrot-menu/applications) to dest folder if it was installed
#   1. Check package name from X-Parrot-Package=
#   2. Check if it is installed using apt list
#   3.
#     a. If it is not in the list -> Remove it in destination folder
#     b. If it is in the list -> copy it to dest folder if it doesnt exists
#   4. Fix duplicate launchers
# Remove old launchers that removed or uninstalled

import os, osproc, strutils

const dirLaucherDest = "/usr/share/applications/"


proc getXPackageName(path: string): string =
  for line in lines(path):
    # Normaly it starts with X-Parrot-Packages
    # But some launchers might contains X-Parrot-package (typo mistake)
    # So we use keyword X-Parrot only
    if line.startsWith("X-Parrot-"):
      return line.split("=")[1]
  return ""


proc fixDebLaunchers() =
  #[
    There are packages from Debian that has custom launchers
    It makes error after install pentest tools from Home Edition
    or create duplicate launchers in menu.
    We are removing them here
    chirp will be removed because of python2 removal problem
  ]#
  let blacklistLauncherName = [
    "org.radare.Cutter.desktop",
    "gpa.desktop",
    "rtlsdr-scanner.desktop",
    "gnuradio-grc.desktop",
    "arduino.desktop",
    "gqrx.desktop",
    "zulucrypt-gui.desktop",
    "zulumount-gui.desktop",
    "ophcrack.desktop",
    "xsser.desktop",
    "io.github.mhogomchungu.sirikali.desktop",
    "etherape.desktop",
    "edb.desktop",
    "lynis.desktop",
    "wireshark.desktop",
    "ettercap.desktop",
    "chirp.desktop",
    "hydra-gtk.desktop",
    "driftnet.desktop",
    "gscriptor.desktop",
    "spectool_gtk.desktop",
    "gksu.desktop",
    "re.rizin.cutter.desktop", # Duplicate launcher of rizin's cutter
    "openjdk-8-policytool.desktop",
  ]
  for fileName in blacklistLauncherName:
    let finalPath = dirLaucherDest & fileName
    if fileExists(finalPath):
      if not tryRemoveFile(finalPath):
        stderr.write("[x] Error while removing " & finalPath & "\n")


proc fixOldLaunchers(path: string) =
  #[
    Some new launchers have been changed name to parrot-* to serv-*
    We try to remove duplicated launchers
  ]#
  let
    fileName = splitPath(path).tail
    newNameArr = ["serv-"]
  var
    destName: string = dirLaucherDest
    isDeleteNeeded: bool

  for checkName in newNameArr:
    isDeleteNeeded = false
    if fileName.startsWith(checkName):
      isDeleteNeeded = true
      destName &= "parrot-" & fileName.replace(checkName, "")
      break

  if isDeleteNeeded:
    if not tryRemoveFile(destName):
      stderr.write("[x] Error while removing " & destName & "\n")


proc update_launchers() =
  const
    # Path must have / at the end of string or it makes error
    dirLauncherSource = "/usr/share/parrot-menu/applications/"

  # Get all installed packages
  let installed = execProcess("apt list --installed | cut -d '/' -f 1").split("\n")

  # Get all file in applications
  var allLaunchers: seq[string]
  for kind, path in walkDir(dirLauncherSource):
    let fileName = splitPath(path).tail
    allLaunchers.add(fileName)
    if fileName.startsWith("parrot-") or fileName.startsWith("serv-"):
      let
        aptParrotPackage = getXPackageName(path)
      if aptParrotPackage != "":
        #[
          1. Case 1: if the package is installed but
            a) Add launcher if it isn't there
            b) Compare launcher's data and update it
          2. Case 2: if the package isn't installed, remove launcher
        ]#

        try:
          let finalDestPath = dirLaucherDest & fileName
          # If the package is installed
          # Check if package name is in installed list. The name matches a line exactly
          # if contains(installed, re("(^|\\n)" & aptParrotPackage & "($|\\n)")):
          if installed.contains(aptParrotPackage):
            # Check if file is not in the final directory
            if not fileExists(finalDestPath) or not sameFileContent(path, finalDestPath):
              # Update new launcher
              try:
                # If file does not exists in dest folder, copy it
                copyFile(path, finalDestPath)
              except:
                stderr.write("[x] Error while copying file " & path & " to " & finalDestPath & "\n")
          else:
            if fileExists(finalDestPath):
              # Remove old launchers here
              if not tryRemoveFile(finalDestPath):
                stderr.write("[x] Error while processing " & path & "\n")
          
          # In this version, we are moving name to serv-, native and being more with different categories
          fixOldLaunchers(path)
        except:
          stderr.write("[x] Error while processing " & path & "\n")
          echo getCurrentExceptionMsg()

  for kind, path in walkDir(dirLaucherDest):
    let currentLauncher = splitPath(path).tail
    # Check if the launcher is Parrot's specific
    if (currentLauncher.startsWith("parrot-") or currentLauncher.startsWith("serv-")) and currentLauncher.endsWith(".desktop"):
      # Get package name from launcher. If package name != "" then it belongs to parrot-menu (or old one)
      let packageOfLauncher = getXPackageName(path)
      if packageOfLauncher != "":
        # We must test if package is not installed here
        if not allLaunchers.contains(currentLauncher) and not installed.contains(packageOfLauncher):
          if not tryRemoveFile(path):
            stderr.write("[x] Error while removing " & path & "\n")

echo "Scanning application launchers"
update_launchers()
echo "Removing duplicate launchers or broken launchers"
fixDebLaunchers()
echo "Launchers are updated"
