package main

import (
	"flag"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"glacierpeak.app/openland/pkg/proc_runners"
	"glacierpeak.app/openland/pkg/utils"
)

func main() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	workersOpt := flag.Int("t", 1, "The number of concurrent jobs being processed")
	src1 := flag.String("src1", "/Users/solidsilver/Code/USFS/TilemergeTest/GPEastFSTopo", "The root directory of the source files")
	src2 := flag.String("src2", "/Users/solidsilver/Code/USFS/TilemergeTest/GPWestFSTopo", "The root directory of the source files")
	massMerge := flag.Bool("m", false, "Merge all tilesets in a given directory (ignores -src2 flag)")
	outDir := flag.String("o", "/Users/solidsilver/Code/USFS/TilemergeTest/GPTilemerge", "The root directory of the source files")
	zLevel := flag.String("z", "17", "Z level of tiles to process")
	verboseOpt := flag.Int("v", 1, "Set the verbosity level:\n"+
		" 0 - Only prints error messages\n"+
		" 1 - Adds run specs and error details\n"+
		" 2 - Adds general progress info\n"+
		" 3 - Adds debug info and details more detail\n")
	flag.Parse()

	utils.SetupLogByLevel(*verboseOpt)

	if *massMerge {
		proc_runners.MassTileMerge(*src1, *outDir, *zLevel, *workersOpt)
	} else {
		proc_runners.TileMerge(*src1, *src2, *outDir, *workersOpt)
	}

}
