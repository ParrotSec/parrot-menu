#[
  A simple nim script to auto parse package name and command execution in desktop launchers
  To auto check broken launcher
  Author: Nong Hoang Tu
]#
import os
import strutils

# const path = "desktop-files" # Use this path for full launchers
const path = "/usr/share/applications/" # Use this path for launchers in system
let env_path = getEnv("PATH").split(":")


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
        echo "[" & thisCmd, "] [", thisName, "] Not found"
