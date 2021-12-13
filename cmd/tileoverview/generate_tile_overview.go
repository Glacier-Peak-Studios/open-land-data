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

	workersOpt := flag.Int("t", 1, "The number of concurrent jobs being processed")
	inDir := flag.String("i", "", "The root directory of the source files")
	zRange := flag.String("z", "17", "Zoom levels to generate. (Ex. \"2-16\") Must start with current zoom level")
	verboseOpt := flag.Int("v", 1, "Set the verbosity level:\n"+
		" 0 - Only prints error messages\n"+
		" 1 - Adds run specs and error details\n"+
		" 2 - Adds general progress info\n"+
		" 3 - Adds debug info and details more detail\n")
	flag.Parse()

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
