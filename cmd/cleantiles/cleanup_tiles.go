package main

import (
	"flag"

	"glacierpeak.app/openland/pkg/proc_runners"
	"glacierpeak.app/openland/pkg/utils"
)

func main() {
	flag.Usage = utils.CliUsage("clean_tiles",
		`This is a tool to remove unnecessary image tiles from a directory of image tile sources.
		It takes in a directory (inDir) and a zoom level (zoomLvl), traverses each directory for that zoom level, 
		and determines the bounds for each tileset such that there are no empty tiles (only white or transparent).
		It removes all files and directories outside of these bounds.
		Note: this should be run prior to the 'fixbackground' command`)

	workersOpt := flag.Int("t", 1, "The number of concurrent jobs being processed")
	zoomLvl := flag.Int("z", 17, "Zoom level of the tileset to clean")
	inDir := flag.String("i", "", "Folder with the pdf files")
	verboseOpt := flag.Int("v", 1, "Set the verbosity level:\n"+
		" 0 - Only prints error messages\n"+
		" 1 - Adds run specs and error details\n"+
		" 2 - Adds general progress info\n"+
		" 3 - Adds debug info and details more detail\n")
	flag.Parse()
	utils.CheckRequiredFlags("i")

	utils.SetupLogByLevel(*verboseOpt)
	proc_runners.CleanupTiles(*inDir, *zoomLvl, *workersOpt)

}
