# Update the lauchers in /usr/share/parrot-menu/applications
# Copy launchers from parrot-menu (/usr/share/parrot-menu/applications) to dest folder if it was installed
#   1. Check package name from X-Parrot-Package=
#   2. Check if it is installed using apt list
#   3.
#     a. If it is not in the list -> Remove it in destination folder
#     b. If it is in the list -> copy it to dest folder if it doesnt exists
# Remove old launchers that removed or uninstalled

import os, osproc, strutils, re

proc update_launchers() =
  const
    # Path must have / at the end of string or it makes error
    dirLauncherSource = "/usr/share/parrot-menu/applications/"
    dirLaucherDest = "/usr/share/applications/" # /usr/share/applications/
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
      echo "[ERROR] Error while getting package name from " & path

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
            echo "[WARNING] Error while copying file " & path & " to " & finalDestPath
        # If it is in there, check to upgrade it
        else:
          # Compare files and update launcher or discard
          if readFile(path) != readFile(finalDestPath):
            try:
              copyFile(path, finalDestPath)
            except:
              echo "[WARNING] Error while copying file " & path & " to " & finalDestPath
      else:
        try:
          # If file is in dest folder -> remove
          if fileExists(finalDestPath):
            # Remove old launchers here
            removeFile(finalDestPath)
        except:
          echo "[WARNING] Error while removing file " & finalDestPath
    except:
      echo "[ERROR] Error while processing " & path

echo "Scanning application launchers"
update_launchers()
echo "Launchers are updated"
