#[
  Group all pentest tools based on installed launchers to system
  this tool is using for regrouping meta package
]#

import os
import strutils

const pathLaunchers = "/usr/share/applications/"

var
  seqInfo: seq[string]
  seqVuln: seq[string]
  seqWeb: seq[string]
  seqPwn: seq[string]
  seqMaintain: seq[string]
  seqPost: seq[string]
  seqPwd: seq[string]
  seqWireless: seq[string]
  seqSniff: seq[string]
  seqFor: seq[string]
  seqAuto: seq[string]
  seqRev: seq[string]
  seqReport: seq[string]


proc parser() =
  for kind, path in walkDir(pathLaunchers):
    if not path.split("/")[^1].startsWith("parrot-"):
      continue
    var
      pkgName = ""
      categories = ""
    for line in lines(path):
      if line.startsWith("X-Parrot-Package"):
        pkgName = line.split("=")[^1]
      elif line.startsWith("Categories"):
        categories = line.split("=")[^1]
    for thisCat in categories.split(";"):
      if thisCat.startsWith("01-") and not contains(seqInfo, pkgName):
        seqInfo.add(pkgName)
      elif thisCat.startsWith("02-") and not contains(seqVuln, pkgName):
        seqVuln.add(pkgName)
      elif thisCat.startsWith("03-") and not contains(seqWeb, pkgName):
        seqWeb.add(pkgName)
      elif thisCat.startsWith("04-") and not contains(seqPwn, pkgName):
        seqPwn.add(pkgName)
      elif thisCat.startsWith("10-") and not contains(seqMaintain, pkgName):
        seqMaintain.add(pkgName)
      elif thisCat.startsWith("08-") and not contains(seqPost, pkgName):
        seqPost.add(pkgName)
      elif thisCat.startsWith("05-") and not contains(seqPwd, pkgName):
        seqPwd.add(pkgName)
      elif thisCat.startsWith("06-") and not contains(seqWireless, pkgName):
        seqWireless.add(pkgName)
      elif thisCat.startsWith("09-") and not contains(seqSniff, pkgName):
        seqSniff.add(pkgName)
      elif thisCat.startsWith("11-") and not contains(seqFor, pkgName):
        seqFor.add(pkgName)
      elif thisCat.startsWith("13-") and not contains(seqAuto, pkgName):
        seqAuto.add(pkgName)
      elif thisCat.startsWith("07-") and not contains(seqRev, pkgName):
        seqRev.add(pkgName)
      elif thisCat.startsWith("12-") and not contains(seqReport, pkgName):
        seqReport.add(pkgName)
  echo "Info"
  for this in seqInfo:
    echo "  ", this, ","
  echo "Vuln"
  for this in seqVuln:
    echo "  ", this, ","
  echo "Web"
  for this in seqWeb:
    echo "  ", this, ","
  echo "Pwn"
  for this in seqPwn:
    echo "  ", this, ","
  echo "Maintain"
  for this in seqMaintain:
    echo "  ", this, ","
  echo "Post Exploit"
  for this in seqPost:
    echo "  ", this, ","
  echo "Password"
  for this in seqPwd:
    echo "  ", this, ","
  echo "Wireless"
  for this in seqWireless:
    echo "  ", this, ","
  echo "Sniff"
  for this in seqSniff:
    echo "  ", this, ","
  echo "Forensic"
  for this in seqFor:
    echo "  ", this, ","
  echo "Car"
  for this in seqAuto:
    echo "  ", this, ","
  echo "Reversing"
  for this in seqRev:
    echo "  ", this, ","
  echo "Report"
  for this in seqReport:
    echo "  ", this, ","

parser()
