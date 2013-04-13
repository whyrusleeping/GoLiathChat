all:
	echo "Please Specify which build GoliathWK, GoliathLite or server"

GoliathWK:
	cd Go
	make GoliathWK

GoliathLite:
	cd Go
	make GoliathLite

server:
	cd go
	make server

