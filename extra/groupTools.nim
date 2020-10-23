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
    var
      pkgName = ""
      categories = ""
    for line in lines(path):
      if line.startsWith("X-Parrot-Package"):
        pkgName = line.split("=")[^1]
      elif line.startsWith("Categories"):
        categories = line.split("=")[^1]
    for thisCat in categories.split(";"):
      if thisCat.startsWith("01-"):
        seqInfo.add(thisCat)
      elif thisCat.startsWith("02-"):
        seqVuln.add(thisCat)
      elif thisCat.startsWith("03-"):
        seqWeb.add(thisCat)
      elif thisCat.startsWith("04-"):
        seqPwn.add(thisCat)
      elif thisCat.startsWith("05-"):
        seqMaintain.add(thisCat)
      elif thisCat.startsWith("06-"):
        seqPost.add(thisCat)
      elif thisCat.startsWith("07-"):
        seqPwd.add(thisCat)
      elif thisCat.startsWith("08-"):
        seqWireless.add(thisCat)
      elif thisCat.startsWith("09-"):
        seqSniff.add(thisCat)
      elif thisCat.startsWith("10-"):
        seqFor.add(thisCat)
      elif thisCat.startsWith("11-"):
        seqAuto.add(thisCat)
      elif thisCat.startsWith("12-"):
        seqRev.add(thisCat)
      elif thisCat.startsWith("13-"):
        seqReport.add(thisCat)
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
