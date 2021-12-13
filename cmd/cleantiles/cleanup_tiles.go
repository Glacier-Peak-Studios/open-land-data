package main

import (
	"flag"

	"glacierpeak.app/openland/pkg/proc_runners"
	"glacierpeak.app/openland/pkg/utils"
)

func main() {

	workersOpt := flag.Int("t", 1, "The number of concurrent jobs being processed")
	// outDir := flag.String("o", "./", "Folder to output the tiff files")
	zoomLvl := flag.Int("z", 17, "Zoom level of the tileset to clean")
	inDir := flag.String("i", "./", "Folder with the pdf files")
	verboseOpt := flag.Int("v", 1, "Set the verbosity level:\n"+
		" 0 - Only prints error messages\n"+
		" 1 - Adds run specs and error details\n"+
		" 2 - Adds general progress info\n"+
		" 3 - Adds debug info and details more detail\n")
	flag.Parse()

	utils.SetupLogByLevel(*verboseOpt)
	proc_runners.CleanupTiles(*inDir, *zoomLvl, *workersOpt)

}
