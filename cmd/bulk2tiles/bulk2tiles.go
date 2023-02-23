package main

import (
	"flag"

	"glacierpeak.app/openland/pkg/proc_runners"
	"glacierpeak.app/openland/pkg/utils"
)

func main() {
	flag.Usage = utils.CliUsage("bulk2tiles",
		`This is a tool for converting large TIFF files into tile sets.
			It takes in a directory of TIFF files, processes each file into a VRT format,
			and then converts the VRT files into tiles at a specified zoom level and number of worker processes (threads).
			The tool outputs the tile sets into a specified output directory.`)

	workersOpt := flag.Int("t", utils.GetDefaultWorkers(), "The number of concurrent jobs being processed")
	zoomLvl := flag.Int("z", 17, "Zoom level to create tiles at")
	inDir := flag.String("i", "", "Folder with the .tif files")
	outDir := flag.String("o", "", "Folder to output the tiles files")
	verboseOpt := flag.Int("v", 1, "Set the verbosity level:\n"+
		" 0 - Only prints error messages\n"+
		" 1 - Adds run specs and error details\n"+
		" 2 - Adds general progress info\n"+
		" 3 - Adds debug info and details more detail\n")

	flag.Parse()

	utils.CheckRequiredFlags("i", "o")

	utils.SetupLogByLevel(*verboseOpt)
	proc_runners.Bulk2Tiles(*inDir, *outDir, *workersOpt, *zoomLvl)
}
