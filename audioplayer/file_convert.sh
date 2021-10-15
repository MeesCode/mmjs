#!/bin/bash

# move file and try to convert
mv "$1" "$1_old"
ffmpeg -i "$1_old" -f flac "$1.flac" &> /dev/null # turn into a flac

if [ $? -eq 0 ]; then # conversion succeeded
    rm "$1_old"
    exit 0
else # conversion failed, restore file
    mv "$1_old" "$1"
    exit 1
fi
