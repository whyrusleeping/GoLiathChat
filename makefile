all: makeGo-debug

makeGo-debug:
	cd Go; make GoliathWK

install: makeGo-debug
