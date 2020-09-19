#[
  A simple nim script to auto parse package name and command execution in desktop launchers
  To auto check broken launcher
  Author: Nong Hoang Tu
]#
import os
import strutils

const
  path = "desktop-files" # Use this path for full launchers
  # const path = "/usr/share/applications/" # Use this path for launchers in system
  index = "/var/lib/apt/lists/" # cache file of apt update
  rMain = "vietnam.deb.parrot.sh_parrot_dists_rolling-testing_main_binary-amd64_Packages" # Index name
  rContrib = "vietnam.deb.parrot.sh_parrot_dists_rolling-testing_contrib_binary-amd64_Packages"
  rNonFree = "vietnam.deb.parrot.sh_parrot_dists_rolling-testing_non-free_binary-amd64_Packages"
let
  env_path = getEnv("PATH").split(":")
  dataMain = readFile(index & rMain)
  dataContrib = readFile(index & rContrib)
  dataNonFree = readFile(index & rNonFree)
var
  checkPackages: seq[string]


proc checkPath(name: string): bool =
  for each_env_path in env_path:
    if fileExists(each_env_path & "/" & name):
      return true
  return false


proc getCmd(data: string): string =
  var testData = data
  if testData.startsWith("\""):
    testData = testData.replace("\"", "")
  if testData.startsWith("menuexec"):
    return testData.replace("\"", "").split(" ")[1]
  else:
    return testData.split(" ")[0]


for file, path in walkDir(path):
  var
    thisName: string
    thisCmd: string
  if path.split("/")[^1].endsWith(".desktop"):
    for line in lines(path):
      if line.startsWith("Exec"):
        thisCmd = getCmd(line.split("=")[1])
      elif line.startsWith("X-Parrot-Package"):
        thisName = line.split("=")[1]
    if thisName != "":
      if not checkPath(thisCmd) and not fileExists(thisCmd):
        # echo "[" & thisCmd, "] [", thisName, "] Not found"
        let searchFor = "Package: " & thisName
        if not contains(dataMain, searchFor) and not contains(dataContrib, searchFor) and not contains(dataNonFree, searchFor):
          if thisName in checkPackages:
            discard
          else:
            checkPackages.add(thisName)


writeFile("/tmp/allPackages", join(checkPackages, "\n"))
