package main

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"

	"solidsilver.dev/openland/utils"
)

func main() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	workersOpt := flag.Int("t", 1, "The number of concurrent jobs being processed")
	src1 := flag.String("src1", "/Users/solidsilver/Code/USFS/TilemergeTest/GPEastFSTopo", "The root directory of the source files")
	src2 := flag.String("src2", "/Users/solidsilver/Code/USFS/TilemergeTest/GPWestFSTopo", "The root directory of the source files")
	outDir := flag.String("o", "/Users/solidsilver/Code/USFS/TilemergeTest/GPTilemerge", "The root directory of the source files")
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

	TileMerge(*src1, *src2, *outDir, *workersOpt)

}

func TileMerge(src1 string, src2 string, out string, workers int) {
	src1BBox := utils.BBoxFromTileset(src1 + "/18")
	src2BBox := utils.BBoxFromTileset(src2 + "/18")

	intersect, _ := utils.GetBBoxIntersect(src1BBox, src2BBox)
	intersect.ExpandBy(1)

	var sources []string

	for ix := src1BBox.Origin().X; ix <= src1BBox.Extent().X; ix++ {
		for iy := src1BBox.Origin().Y; iy <= src1BBox.Extent().Y; iy++ {
			sources = append(sources, filepath.Join(out, "18", strconv.Itoa(ix), strconv.Itoa(iy)+".png"))
		}
	}

	for ix := src2BBox.Origin().X; ix <= src2BBox.Extent().X; ix++ {
		for iy := src2BBox.Origin().Y; iy <= src2BBox.Extent().Y; iy++ {
			curP := utils.Point{X: ix, Y: iy}
			if !intersect.PointInBBox(curP) {
				sources = append(sources, filepath.Join(out, "18", strconv.Itoa(ix), strconv.Itoa(iy)+".png"))
			} else {
				println("In bbox: ", ix, "/", iy)
			}
		}
	}

	jobCount := len(sources)
	jobs := make(chan string, jobCount)
	results := make(chan string, jobCount)

	log.Warn().Msgf("Running with %v workers", workers)
	for i := 0; i < workers; i++ {
		go utils.TilesetMergeWorker(jobs, results, src1, src2)
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

func Filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}
