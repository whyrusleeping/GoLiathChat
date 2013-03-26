all: all-release

all-debug: makeGo-debug makeWebkit-debug makeConsole-debug
all-release: makeGo-release makeWebkit-release makeConsole-release

makeGo-debug:
	cd Go; make all

makeGo-release: makeGo-debug

makeWebkit-debug:
	cd WebkitUI/src; qmake "CONFIG+=debug" GoChat_QtUI.pro; make

makeWebkit-release:
	cd WebkitUI/src; qmake "CONFIG+=release" GoChat_QtUI.pro; make

makeConsole-debug:

makeConsole-release:

