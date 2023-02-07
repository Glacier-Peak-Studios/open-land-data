#!/bin/bash
for filename in ./TIFs/*.tif; do
	[ -e "$filename" ] || continue
	#echo "$filename"
	gdal2tiles.py "$filename" --zoom="17" --processes=28 --xyz --resume
done