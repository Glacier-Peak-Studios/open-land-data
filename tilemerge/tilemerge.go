package main

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/schollz/progressbar/v3"

	"solidsilver.dev/openland/utils"
)

func main() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	workersOpt := flag.Int("t", 1, "The number of concurrent jobs being processed")
	src1 := flag.String("src1", "/Users/solidsilver/Code/USFS/TilemergeTest/GPEastFSTopo", "The root directory of the source files")
	src2 := flag.String("src2", "/Users/solidsilver/Code/USFS/TilemergeTest/GPWestFSTopo", "The root directory of the source files")
	massMerge := flag.Bool("m", false, "Merge all tilesets in a given directory (ignores -src2 flag)")
	outDir := flag.String("o", "/Users/solidsilver/Code/USFS/TilemergeTest/GPTilemerge", "The root directory of the source files")
	zLevel := flag.String("z", "17", "Z level of tiles to process")
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
	case 1:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
		// log.SetLevel(log.WarnLevel)
	case 2:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case 3:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		// log.SetReportCaller(true)
	default:
		break
	}

	if *massMerge {
		MassTileMerge(*src1, *outDir, *zLevel, *workersOpt)
	} else {
		TileMerge(*src1, *src2, *outDir, *workersOpt)
	}

}


func MassTileMerge(setsDir string, out string, zLevel string, workers int) {

	m, tileList := utils.GetAllTiles0(setsDir, zLevel, workers)

	log.Warn().Msg("Done gathering tiles.")

	rsltLen := len(tileList)
	jobCount := 32
	jobs := make(chan utils.Tile, jobCount)
	results := make(chan string, jobCount)
	readChan := make(chan int, 1)

	log.Warn().Msgf("Running with %v workers", workers)
	go resultReaderWorker(results, jobs, rsltLen, readChan)

	mapLock := sync.RWMutex{}
	
	for i := 0; i < workers; i++ {
		go utils.TilesetMergeWorker0(jobs, results, m, out, setsDir, &mapLock)
	}
	for _, tile := range tileList {
		jobs <- tile
	}

	<-readChan

	log.Warn().Msg("Done with all jobs")

}


func resultReaderWorker(toRead <-chan string, jobs chan utils.Tile, resultCount int, result chan<- int) {

	progBar := progressbar.NewOptions(resultCount,
    progressbar.OptionSetDescription("Merging tiles..."),
		progressbar.OptionSetItsString("tiles"),
		progressbar.OptionShowIts(),
		progressbar.OptionThrottle(1*time.Second),
		progressbar.OptionSetPredictTime(true),
    progressbar.OptionSetTheme(progressbar.Theme{
        Saucer:        "=",
        SaucerHead:    ">",
        SaucerPadding: " ",
        BarStart:      "[",
        BarEnd:        "]",
    }),
	)

	for i := 0; i < resultCount; i++ {
		<-toRead
		progBar.Add(1)
	}
	close(jobs)
	progBar.Finish()
	result <- 1

}

func TileMerge(src1 string, src2 string, out string, workers int) {
	src1BBox, _ := utils.BBoxFromTileset(src1 + "/18")
	src2BBox, _ := utils.BBoxFromTileset(src2 + "/18")

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


func queueSources(sources []string, jobs chan<- string) {
	for _, source := range sources {
		jobs <- source
	}
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
