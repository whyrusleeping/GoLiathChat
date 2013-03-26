#!/bin/bash
echo "Starting documentation server..."
echo "open 127.0.0.1:6060 in your browser"
echo "Ctrl + c to stop."
godoc -http=":6060" -path="."
