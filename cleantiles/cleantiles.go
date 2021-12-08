package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/schollz/progressbar/v3"
	"glacierpeak.app/openland/utils"
)

func main() {

	workersOpt := flag.Int("t", 1, "The number of concurrent jobs being processed")
	// outDir := flag.String("o", "./", "Folder to output the tiff files")
	zoomLvl := flag.Int("z", 17, "Zoom level of the tileset to clean")
	inDir := flag.String("i", "./", "Folder with the pdf files")
	verboseOpt := flag.Int("v", 1, "Set the verbosity level:\n"+
		" 0 - Only prints error messages\n"+
		" 1 - Adds run specs and error details\n"+
		" 2 - Adds general progress info\n"+
		" 3 - Adds debug info and details more detail\n")
	flag.Parse()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

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

	// log.Warn().Msgf("Searching sources dir: %v", *outDir)
	// sources, _ := utils.WalkMatch(*inDir, "*.pdf")
	files, _ := ioutil.ReadDir(*inDir)
	files = utils.FilterFI(files, func(file os.FileInfo) bool {
		return file.IsDir()
	})
	sources := utils.MapFI(files, func(file os.FileInfo) string {
		return filepath.Join(*inDir, file.Name())
	})

	jobCount := len(sources)
	jobs := make(chan string, jobCount)
	results := make(chan string, jobCount)

	log.Warn().Msgf("Running with %v workers", *workersOpt)
	for i := 0; i < *workersOpt; i++ {
		go utils.TileTrimWorker(jobs, results, *zoomLvl)
	}
	queueSources(sources, jobs)

	progBar := progressbar.NewOptions(len(sources),
    progressbar.OptionSetDescription("Cleaning tiles..."),
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

	for i := 0; i < jobCount; i++ {
		var rst = <-results
		progBar.Add(1)
		log.Debug().Msg(rst)
	}
	close(jobs)
	progBar.Finish()
	log.Warn().Msg("Done with all jobs")

}

func queueSources(sources []string, jobs chan<- string) {
	progBar := progressbar.NewOptions(len(sources),
    progressbar.OptionSetDescription("Preparing tiles for cleaning..."),
    progressbar.OptionSetTheme(progressbar.Theme{
        Saucer:        "=",
        SaucerHead:    ">",
        SaucerPadding: " ",
        BarStart:      "[",
        BarEnd:        "]",
    }),
	)
	for _, source := range sources {
		jobs <- source
		progBar.Add(1)
	}
	progBar.Finish()
}
