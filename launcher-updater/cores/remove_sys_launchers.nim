import os
import program_consts


proc fixDebLaunchers*() =
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
    "org.keepassxc.KeePassXC.desktop",
  ]

  for fileName in blacklistLauncherName:
    let finalPath = dirLaucherDest & fileName
    if fileExists(finalPath):
      if not tryRemoveFile(finalPath):
        stderr.write("[x] Error while removing " & finalPath & "\n")
