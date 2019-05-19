all:
	/usr/bin/go build update-launchers.go
	/usr/bin/strip update-launchers
