import os
import strutils
import program_consts
import parseutils


proc checkValidBinary(path: string) =
  #[
    When the script is invoked by APT, the PATH is the system's default of root user
    Therefore /usr/games/ is not in the path. We have to set our custom PATH manually
    to avoid the false positive
  ]#
  putEnv("PATH", "/usr/local/bin:/usr/bin:/bin:/usr/local/games:/usr/games:/usr/local/sbin:/usr/sbin:/sbin")
  for line in lines(path):
    if line.startsWith("Exec="):
      let execFile = line.captureBetween('=', ' ').replace("\"", "").replace("\'", "")
      if findExe(execFile) == "":
        echo " [-] Missing executable file ", execFile, " at launcher ", path
        return
  return


proc removeOldLaunchers*() =
  #[
    Remove all old launchers that isn't in Parrot's launchers db anymore
  ]#
  for kind, path in walkDir(dirLaucherDest):
    if kind != pcFile:
      continue
    var isCheckBinary = true
    let currentLauncher = splitPath(path).tail
    # Check if the launcher is Parrot's specific
    if (currentLauncher.startsWith("parrot-") or currentLauncher.startsWith("serv-")) and currentLauncher.endsWith(".desktop"):
      # Get package name from launcher. If package name != "" then it belongs to parrot-menu (or old one)
      # We can use the fileExist from source method because some packages are having custom launcher in the package
      let srcToCheck = dirLauncherSource & path.splitPath().tail
      if not fileExists(srcToCheck):
        isCheckBinary = false
        if not tryRemoveFile(path):
          echo "Failed to remove ", path
    if isCheckBinary:
      checkValidBinary(path)
