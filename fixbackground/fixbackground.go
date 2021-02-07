package main

import (
	"flag"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"

	"solidsilver.dev/openland/utils"
)

func main() {

	workersOpt := flag.Int("t", 1, "The number of concurrent jobs being processed")
	inDir := flag.String("i", "", "The root directory of the source files")
	outDir := flag.String("o", "", "The output directory of the source files")
	// zRange := flag.String("z", "17", "Zoom level to fix. (Ex. \"2-16\") Must start with current zoom level")
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

	// rng := strings.Split(*zRange, "-")
	// zMin, _ := strconv.Atoi(rng[0])
	// zMax := zMin
	// if len(rng) >= 2 {
	// 	zMax, _ = strconv.Atoi(rng[1])
	// }
	// log.Info().Msgf("Generating zoom from %v to %v", zMax, zMin)
	FixBackground(*inDir, *outDir, *workersOpt)
	// CreateOverviewRange(zMax, zMin, *inDir, *workersOpt)

}

func FixBackground(dir string, out string, workers int) {
	m := make(map[string]bool)
	// sources, _ := utils.WalkMatch(dir, "*.png")
	sources := utils.GetAllTiles2(filepath.Join(dir, "17"), workers)
	var tileList []utils.Tile

	for _, source := range sources {
		tile, _ := utils.PathToTile(source)
		m[tile.GetPathXY()] = true
		tileList = append(tileList, tile)
		// tileList = utils.AppendSetT(tileList, tile)
	}

	jobCount := len(tileList)
	jobs := make(chan utils.Tile, jobCount)
	results := make(chan string, jobCount)

	log.Warn().Msgf("Running with %v workers", workers)
	for i := 0; i < workers; i++ {
		go utils.FixBackgroundWorker(jobs, results, m, dir, out)
		// go utils.TilesetMergeWorker2(jobs, results, m, out)
	}
	for _, tile := range tileList {
		jobs <- tile
	}
	for i := 0; i < jobCount; i++ {
		var rst = <-results
		log.Debug().Msg(rst)
	}
	close(jobs)
	log.Warn().Msg("Done with all jobs")
}
