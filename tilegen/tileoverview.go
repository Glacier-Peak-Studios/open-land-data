package main

import (
	"flag"
	"os"
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
	outDir := flag.String("i", "/Users/solidsilver/Code/USFS/TilemergeTest/GPTilemerge", "The root directory of the source files")
	zRange := flag.String("z", "18", "Zoom levels to generate. (Ex. \"2-16\") Must start with current zoom level")
	verboseOpt := flag.Int("v", 1, "Set the verbosity level:\n"+
		" 0 - Only prints error messages\n"+
		" 1 - Adds run specs and error details\n"+
		" 2 - Adds general progress info\n"+
		" 3 - Adds debug info and details more detail\n")
	flag.Parse()

	switch *verboseOpt {
	case 0:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		// log.SetLevel(log.ErrorLevel)
		break
	case 1:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
		// log.SetLevel(log.WarnLevel)
		break
	case 2:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		break
	case 3:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		// log.SetReportCaller(true)
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

	CreateOverviewRange(zMax, zMin, *outDir, *workersOpt)

}

func CreateOverviewRange(zMax int, zMin int, dir string, workers int) {
	for i := zMax; i > zMin; i-- {
		CreateOverview(filepath.Join(dir, strconv.Itoa(i)), workers)
	}
}

func CreateOverview(dir string, workers int) {
	log.Warn().Msgf("Searching sources dir: %v", dir)
	sources, _ := utils.WalkMatch(dir, "*.png")
	sources = utils.Filter(sources, isEvenTile)

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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
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

func isEvenTile(path string) bool {
	// println("Checking path: ", path)
	yStr := utils.StripExt(filepath.Base(path))
	fdir := filepath.Dir(path)
	xStr := filepath.Base(fdir)

	// println("X: ", xStr, " - Y: ", yStr)

	x, err := strconv.Atoi(xStr)
	y, err := strconv.Atoi(yStr)

	if err != nil {
		log.Error().Msg("Could not parse string to int")
	}

	return x%2 == 0 && y%2 == 0
}
