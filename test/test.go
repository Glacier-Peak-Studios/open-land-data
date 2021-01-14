package main

import (
	"flag"
	"strings"

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

	daw := utils.GetGeoPDFLayers(*in)

	lrs := utils.Filter(daw, LayerFilter)

	for _, val := range lrs {
		println(val)
	}
}

func LayerFilter(layer string) bool {
	if strings.HasPrefix(layer, "Quadrangle.Neatline") {
		return false
	}
	if strings.HasPrefix(layer, "Quadrangle.2_5") {
		return false
	}
	if strings.HasPrefix(layer, "Quadrangle_Ext") {
		return false
	}
	if strings.HasPrefix(layer, "Adjacent") {
		return false
	}
	if strings.HasPrefix(layer, "Other") {
		return false
	}
	if strings.HasPrefix(layer, "Quadrangle.UTM") {
		return false
	}

	return true
}
