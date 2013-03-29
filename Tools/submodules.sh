#!/bin/sh
echo "Project Submodules:"
cd ../; 
git submodule
echo "Initializing Submodules:"
git submodule init
echo "Updating Submodules:"
git submodule update
git submodule foreach git pull origin master
echo "Submodule setup complete..."

