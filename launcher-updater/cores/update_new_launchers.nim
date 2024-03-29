import os
import strutils
import program_consts


proc getXPackageName(path: string): string =
  for line in lines(path):
    # Normaly it starts with X-Parrot-Packages
    # But some launchers might contains X-Parrot-package (typo mistake)
    # So we use keyword X-Parrot only
    if line.startsWith("X-Parrot-"):
      return line.split("=")[1]
  return ""


proc query_installed(): seq[string] =
  var
    pkg_name: string
  for line in lines(path):
    if line.startsWith("Package: "):
      pkg_name = line.split(": ")[1]
    elif line.startsWith("Status: "):
      if line == "Status: install ok installed":
        result.add(pkg_name)


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


proc update_launchers*() =
  # Get all installed packages
  let installed = query_installed()

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
