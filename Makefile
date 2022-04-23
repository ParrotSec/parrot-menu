install:
	mkdir -p build
	nim c --nimcache:/tmp -d:release -o:build/update-launchers launcher-updater/update_launchers.nim
