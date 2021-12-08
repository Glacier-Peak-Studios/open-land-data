package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/schollz/progressbar/v3"
	"glacierpeak.app/openland/utils"
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
	sources := utils.GetAllTiles2(dir, workers)
	// sources, _ := utils.WalkMatch(dir, "*")
	m := make(map[string]bool)
	var overviews []string

	for _, source := range sources {
		over := utils.OverviewRoot(source)
		tile, _ := utils.PathToTile(over)
		if !m[tile.GetPathXY()] {
			overviews = append(overviews, over)
			m[tile.GetPathXY()] = true
		}

		// tileList = utils.AppendSetT(tileList, tile)
	}

	progBar := progressbar.NewOptions(len(overviews),
    progressbar.OptionSetDescription(fmt.Sprintf("Generating overview for level %s...", dir)),
		progressbar.OptionSetItsString("tiles"),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
    progressbar.OptionSetTheme(progressbar.Theme{
        Saucer:        "=",
        SaucerHead:    ">",
        SaucerPadding: " ",
        BarStart:      "[",
        BarEnd:        "]",
    }),
	)
	// sources = utils.SetMap(sources, utils.OverviewRoot)

	jobCount := len(overviews)
	jobs := make(chan string, jobCount)
	results := make(chan string, jobCount)

	log.Warn().Msgf("Running with %v workers", workers)
	for i := 0; i < workers; i++ {
		go utils.OverviewWorker(jobs, results)
	}
	go queueSources(overviews, jobs)

	for i := 0; i < jobCount; i++ {
		var rst = <-results
		log.Debug().Msg(rst)
		progBar.Add(1)
	}
	close(jobs)
	progBar.Finish()
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
