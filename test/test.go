package main

import (
	"flag"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"

	"solidsilver.dev/openland/utils"
)

func main() {
	// workersOpt := flag.Int("t", 1, "The number of concurrent jobs being processed")
	in := flag.String("i", "", "Filepath to use for testing")
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

	tileImg, _ := utils.DecodePNGFromPath(*in)

	rects, _ := utils.GetCoverageRectCorner(tileImg, 1)

	// covRect, _ := utils.GetCoverageRectSide(tileImg, 1)
	// covRect1, _ := utils.GetCoverageRectSide(tileImg, 3)

	// covRect := rects[0]
	// covRect1 := rects[1]

	newImg := utils.ImgOverRects(tileImg, rects)

	utils.EncodePNGToPath("test.png", newImg)

	// intersect := covRect.Intersect(covRect1)

	// println("rect: %v %v", intersect.Dx(), intersect.Dy())
}
