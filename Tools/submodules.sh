#!/bin/sh
echo "Project Submodules:"
cd ../; 
git submodule
echo "Initializing Submodules:"
git submodule init
echo "Updating Submodules:"
git submodule update
echo "Submodule setup complete..."

