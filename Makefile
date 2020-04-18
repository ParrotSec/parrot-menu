install:
	nim c --nimcache:/tmp -d:release -o:update-launchers update_launchers.nim
