#!/bin/bash
RESULT_DIR=generated

swap_coords() {
    echo "Swapping Coords"
    # rm -r $SOURCEDIR/swapped
    # mkdir $SOURCEDIR/swapped
    [ ! -d "$SOURCEDIR/swapped" ] && mkdir $SOURCEDIR/swapped
    for dir in $SOURCEDIR/*/ ; do
        swpdir="swapped/$(basename $dir)"
        [ ! -d "$SOURCEDIR/$swpdir" ] && mkdir $SOURCEDIR/$swpdir
        for file in $dir*.geojson; do
            filefull=$(basename -- "$file")
            extension="${filefull##*.}"
            filename="${filefull%.*}"
            # echo "[dir:$dir][swpdir:$swpdir][filename:$filename][file:$file]"
            [ ! -f $SOURCEDIR/$swpdir/$filename.geojson ] && python swapgeojson.py -f $file -o $SOURCEDIR/$swpdir && echo "Swapped $swpdir"
        done
    done
}