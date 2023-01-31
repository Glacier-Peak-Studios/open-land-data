package main

import (
	"flag"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"glacierpeak.app/openland/pkg/proc_runners"
	"glacierpeak.app/openland/pkg/utils"
)

func main() {
	flag.Usage = utils.CliUsage("generate_tile_overview",
		`This is a tool for creating various overview zoom levels from a base tileset.
		It takes in a tileset and a range of zoom levels (where the largest zoom in the range
		is the zoom of the curren ttileset), and generates sucessively "zoomed out" tile sets.
		Ex: a zoom range of "2-17" would assume the input tileset was generated at a level 17,
		and it wouuld generate tilesets at zoom levels 2-16.`)

	workersOpt := flag.Int("t", 4, "The number of concurrent jobs being processed")
	inDir := flag.String("i", "", "The root directory of the source files")
	zRange := flag.String("z", "", "Zoom levels to generate. (Ex. \"2-17\") Must start with current zoom level")
	verboseOpt := flag.Int("v", 1, "Set the verbosity level:\n"+
		" 0 - Only prints error messages\n"+
		" 1 - Adds run specs and error details\n"+
		" 2 - Adds general progress info\n"+
		" 3 - Adds debug info and details more detail\n")
	flag.Parse()

	utils.CheckRequiredFlags("i", "z")

	utils.SetupLogByLevel(*verboseOpt)

	rng := strings.Split(*zRange, "-")
	zMin, _ := strconv.Atoi(rng[0])
	zMax := zMin
	if len(rng) >= 2 {
		zMax, _ = strconv.Atoi(rng[1])
	}
	log.Info().Msgf("Generating zoom from %v to %v", zMax, zMin)

	proc_runners.CreateOverviewRange(zMax, zMin, *inDir, *workersOpt)

}

// func logMsg(results chan<- string, source, msg string) {
// 	toSend := source + ": " + msg
// 	results <- toSend
// }

// func queueSources(sources []string, jobs chan<- string) {
// 	for _, source := range sources {
// 		jobs <- source
// 	}
// }
