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
	flag.Usage = utils.CliUsage("tilemerge",
		`This is a tool for merging (typically spatially adjacent) tilesets into one tileset.
		In default mode, it takes in two source tileset directories and merges them.
		In massMerge mode, it takes in a directory that contains many tileset directories, and merges them all.
		The output from either mode is put into the spcified 'o' directory.`)

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	workersOpt := flag.Int("t", 4, "The number of concurrent jobs being processed")
	src1 := flag.String("src1", "", "The root directory of the source files")
	src2 := flag.String("src2", "", "The root directory of the source files")
	massMerge := flag.Bool("m", false, "Merge all tilesets in a given directory (ignores -src2 flag)")
	outDir := flag.String("o", "", "The root directory of the source files")
	zLevel := flag.String("z", "17", "Z level of tiles to process")
	verboseOpt := flag.Int("v", 1, "Set the verbosity level:\n"+
		" 0 - Only prints error messages\n"+
		" 1 - Adds run specs and error details\n"+
		" 2 - Adds general progress info\n"+
		" 3 - Adds debug info and details more detail\n")
	flag.Parse()

	utils.CheckRequiredFlags("src1", "o")

	utils.SetupLogByLevel(*verboseOpt)

	if *massMerge {
		proc_runners.MassTileMerge(*src1, *outDir, *zLevel, *workersOpt)
	} else {
		proc_runners.TileMerge(*src1, *src2, *outDir, *workersOpt)
	}

}
