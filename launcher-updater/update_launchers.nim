# Update the lauchers in /usr/share/parrot-menu/applications
# Copy launchers from parrot-menu (/usr/share/parrot-menu/applications) to dest folder if it was installed
#   1. Check package name from X-Parrot-Package=
#   2. Check if it is installed using apt list
#   3.
#     a. If it is not in the list -> Remove it in destination folder
#     b. If it is in the list -> copy it to dest folder if it doesnt exists
#   4. Fix duplicate launchers
# Remove old launchers that removed or uninstalled

import cores / [remove_sys_launchers, check_broken_launchers, update_new_launchers]


echo "Scanning application launchers"
update_launchers()
echo "Removing duplicate launchers or broken launchers"
removeOldLaunchers()
fixDebLaunchers()
echo "Launchers are updated"
