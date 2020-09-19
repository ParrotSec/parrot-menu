# Update the lauchers in /usr/share/parrot-menu/applications
# Copy launchers from parrot-menu (/usr/share/parrot-menu/applications) to dest folder if it was installed
#   1. Check package name from X-Parrot-Package=
#   2. Check if it is installed using apt list
#   3.
#     a. If it is not in the list -> Remove it in destination folder
#     b. If it is in the list -> copy it to dest folder if it doesnt exists
#   4. Fix duplicate launchers
# Remove old launchers that removed or uninstalled

import os, osproc, strutils, re

const dirLaucherDest = "/usr/share/applications/"

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
    "parrot-sipsak.desktop", # start removing broken launchers from here
    "parrot-ragg2-cc.desktop",
    "parrot-acccheck.desktop",
    "parrot-blindelephant.desktop",
    "parrot-bluemaho.desktop",
    "parrot-cachedump.desktop",
    "parrot-dbpwaudit.desktop",
    "parrot-deblaze.desktop",
    "parrot-dff.desktop",
    "parrot-dff-gui.desktop",
    "parrot-dnmap-client.desktop",
    "parrot-dnmap-server.desktop",
  ]
  for fileName in blacklistLauncherName:
    let finalPath = dirLaucherDest & fileName
    if fileExists(finalPath):
      if not tryRemoveFile(finalPath):
        stderr.write("[x] Error while removing " & finalPath & "\n")


proc fixOldLaunchers(path: string) =
  let
    fileName = path.split("/")[^1]
    newNameArr = ["serv-", "native-"]
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
  let installed = execProcess("apt list --installed | cut -d '/' -f 1")

  # Get all file in applications
  for kind, path in walkDir(dirLauncherSource):
    # Use regex to get X-Parrot-Package value
    let fileData = readFile(path)
    var aptParrotPackage = ""
    # Try get package name from X-parrot-package section
    try:
      aptParrotPackage = findAll(fileData, re("X-Parrot-[Pp]ackage=(\\S+)"))[0].split("=")[1]
    except IndexError:
      aptParrotPackage = findAll(fileData, re("Name=(\\S+)"))[0].split("=")[1].toLower() # TODO packages may have Upper char?
    except:
      stderr.write("[x] Error while getting package name from " & path & "\n")

    #[
      1. Case 1: if the package is installed but
        a) Add launcher if it isn't there
        b) Compare launcher's data and update it
      2. Case 2: if the package isn't installed, remove launcher
    ]#

    try:
      let finalDestPath = dirLaucherDest & splitPath(path).tail
      # If the package is installed
      # Check if package name is in installed list. The name matches a line exactly
      if contains(installed, re("(^|\\n)" & aptParrotPackage & "($|\\n)")):
        # Check if file is not in the final directory
        if not fileExists(finalDestPath):
          # Update new launcher
          try:
            # If file does not exists in dest folder, copy it
            copyFile(path, finalDestPath)
          except:
            stderr.write("[x] Error while copying file " & path & " to " & finalDestPath & "\n")
        # If it is in there, check to upgrade it
        else:
          # Compare files and update launcher or discard
          # if readFile(path) != readFile(finalDestPath):
          if not sameFileContent(path, finalDestPath):
            try:
              copyFile(path, finalDestPath)
            except:
              stderr.write("[x] Error while updating launcher " & path & " to " & finalDestPath & "\n")
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

echo "Scanning application launchers"
update_launchers()
echo "Removing duplicate launchers or broken launchers"
fixDebLaunchers()
echo "Launchers are updated"
