package main

import (
	"flag"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"

	"solidsilver.dev/openland/utils"
)

func main() {
	// workersOpt := flag.Int("t", 1, "The number of concurrent jobs being processed")
	in := flag.String("i", "/Users/solidsilver/Code/USFS/TilemergeTest/GPTilemerge", "The root directory of the source files")
	// zRange := flag.String("z", "18", "Zoom levels to generate. (Ex. \"2-16\") Must start with current zoom level")
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

	daw := utils.GetGeoPDFLayers(*in)

	lrs := utils.Filter(daw, utils.RemoveLayer)

	println("Layers to Remove:")
	for _, val := range lrs {
		println(val)
	}
}
