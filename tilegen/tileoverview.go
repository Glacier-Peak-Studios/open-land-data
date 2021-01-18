package main

import (
	"flag"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"

	"solidsilver.dev/openland/utils"
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

	switch *verboseOpt {
	case 0:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		break
	case 1:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
		break
	case 2:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		break
	case 3:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		break
	default:
		break
	}

	rng := strings.Split(*zRange, "-")
	zMin, _ := strconv.Atoi(rng[0])
	zMax := zMin
	if len(rng) >= 2 {
		zMax, _ = strconv.Atoi(rng[1])
	}
	log.Info().Msgf("Generating zoom from %v to %v", zMax, zMin)

	CreateOverviewRange(zMax, zMin, *inDir, *workersOpt)

}

func CreateOverviewRange(zMax int, zMin int, dir string, workers int) {
	for i := zMax; i > zMin; i-- {
		CreateOverview(filepath.Join(dir, strconv.Itoa(i)), workers)
	}
}

func CreateOverview(dir string, workers int) {
	log.Warn().Msgf("Searching sources dir: %v", dir)
	sources, _ := utils.WalkMatch(dir, "*.png")
	sources = utils.SetMap(sources, utils.OverviewRoot)

	jobCount := len(sources)
	jobs := make(chan string, jobCount)
	results := make(chan string, jobCount)

	log.Warn().Msgf("Running with %v workers", workers)
	for i := 0; i < workers; i++ {
		go utils.OverviewWorker(jobs, results)
	}
	queueSources(sources, jobs)

	for i := 0; i < jobCount; i++ {
		var rst = <-results
		log.Debug().Msg(rst)
	}
	close(jobs)
	log.Warn().Msg("Done with all jobs")
}

func logMsg(results chan<- string, source, msg string) {
	toSend := source + ": " + msg
	results <- toSend
}

func queueSources(sources []string, jobs chan<- string) {
	for _, source := range sources {
		jobs <- source
	}
}
