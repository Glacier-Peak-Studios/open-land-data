package main

import (
	"flag"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"glacierpeak.app/openland/pkg/proc_runners"
	"glacierpeak.app/openland/pkg/utils"
)

func main() {
	flag.Usage = utils.CliUsage("fixbackground",
		`This is a tool for fixing the background at the edge of tile sets.
		It looks at the edges of the tilets in order to get rid of any white borders
		that start in the middle of a tile. It converts all white borders to be transparent.
		It then saves this fixed tileset to the specified output directory.
		Note: 'cleanup_tiles' should be run prior to running this command`)

	workersOpt := flag.Int("t", utils.GetDefaultWorkers(), "The number of concurrent jobs being processed")
	inDir := flag.String("i", "", "The root directory of the source files")
	outDir := flag.String("o", "", "The output directory of the source files")
	zLevel := flag.String("z", "17", "Z level of tiles to process")
	// zRange := flag.String("z", "17", "Zoom level to fix. (Ex. \"2-16\") Must start with current zoom level")
	verboseOpt := flag.Int("v", 1, "Set the verbosity level:\n"+
		" 0 - Only prints error messages\n"+
		" 1 - Adds run specs and error details\n"+
		" 2 - Adds general progress info\n"+
		" 3 - Adds debug info and details more detail\n")
	flag.Parse()

	utils.CheckRequiredFlags("i", "o")

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
	proc_runners.FixBackground(*inDir, *outDir, *workersOpt, *zLevel)
	// CreateOverviewRange(zMax, zMin, *inDir, *workersOpt)
}
