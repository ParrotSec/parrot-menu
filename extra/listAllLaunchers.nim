# Show all launchers of parrot menu
# Idea: show list in json
import os
import strutils
# import sequtils

type
  FileInfo = object
    # pname: string # package name. Use to group all launchers from 1 package
    cname: string # command exec or other name to define different package
    desc: string # description of this launcher
    tname: string # Too name: name of toolin name section
  Launcher = object
    pname: string
    subInfo*: seq[FileInfo]    


var allLaunchers: seq[Launcher]
const path = "/home/dmknght/ParrotProject/parrot-menu/desktop-files/"

for kind, path in walkDir(path):
  var
    thisInfo: FileInfo
    pname: string
    cname: string
    desc: string
    tname: string
    # isBreak = false
  for line in lines(path):
    if line.startsWith("Name"):
      tname = line.split("=")[1]
    elif line.startsWith("Comment"):
      desc = line.split("=")[1]
    elif line.startsWith("Exec"):
      cname = line.split("=")[1]
    elif line.startsWith("X-Parrot-Package") or line.startsWith("X-parrot-Package"):
      pname = line.split("=")[1]
    else:
      discard
  # echo pname shows empty lines
  thisInfo = FileInfo(
    cname: cname,
    desc: desc,
    tname: tname,
  )
  # for eachLauncher in allLaunchers:
  #   if eachLauncher.pname == pname:
  #     eachLauncher.subInfo.add(thisInfo)
  #     isBreak = true
  # if isBreak:
  #   continue
  var thisLauncher: Launcher
  
  thisLauncher = Launcher(
    pname: pname,
  )
  thisLauncher.subInfo.add(thisInfo)
  allLaunchers.add(thisLauncher)

echo len(allLaunchers)