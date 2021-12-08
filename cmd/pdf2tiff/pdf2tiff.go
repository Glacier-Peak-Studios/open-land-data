package main

import (
	"flag"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	utils "glacierpeak.app/openland/pkg"
)

func main() {

	workersOpt := flag.Int("t", 4, "The number of concurrent jobs being processed")
	outDir := flag.String("o", "./", "Folder to output the tiff files")
	inDir := flag.String("i", "./", "Folder with the pdf files")
	filterFile := flag.String("f", "./keepLayers.txt", "File containing layers to include when converting")
	dpi := flag.String("dpi", "750", "DPI of the output tiffs")
	verboseOpt := flag.Int("v", 1, "Set the verbosity level:\n"+
		" 0 - Only prints error messages\n"+
		" 1 - Adds run specs and error details\n"+
		" 2 - Adds general progress info\n"+
		" 3 - Adds debug info and details more detail\n")
	flag.Parse()

	setupLogByLevel(*verboseOpt)

	// log.Warn().Msgf("Searching sources dir: %v", *outDir)
	// sources, _ := utils.WalkMatch(*inDir, "*.pdf")

	// jobCount := len(sources)
	// jobs := make(chan string, jobCount)
	// results := make(chan string, jobCount)

	filterLayers := utils.ReadInFilterList(*filterFile)

	PDF2TIFF(*inDir, *outDir, filterLayers, *dpi, *workersOpt)

	// // rmlayers := "\"" + utils.FtoStr("rmlayers.txt") + "\""

	// log.Warn().Msgf("Running with %v workers", *workersOpt)

	// // defArgs := {"GPWestFSTopo.pdf -D 750 -r \"$(cat rmlayers.txt)\" -t EPSG:3857"}
	// for i := 0; i < *workersOpt; i++ {
	// 	// go utils.CommandRunner(jobs, results, "./convert-geopdf.py", "-D", "750", "-r", "\"$(cat rmlayers.txt)\"", "-t", "EPSG:3857")
	// 	go utils.PDF2TiffWorker(jobs, results, filterLayers, *outDir, "gdalwarp", "-co", "TILED=YES", "-co", "TFW=YES", "-t_srs", "EPSG:3857", "-r", "near", "-overwrite", "-dstnodata", "255", "--config", "GDAL_PDF_DPI", *dpi)
	// }
	// queueSources(sources, jobs)

	// for i := 0; i < jobCount; i++ {
	// 	var rst = <-results
	// 	log.Debug().Msg(rst)
	// }
	// close(jobs)
	// log.Warn().Msg("Done with all jobs")

}

func PDF2TIFF(rootDir string, outDir string, filterLayers []string, dpi string, workers int) {
	sources, _ := utils.WalkMatch(rootDir, "*.pdf")

	jobCount := len(sources)
	jobs := make(chan string, jobCount)
	results := make(chan string, jobCount)

	// filterLayers := utils.ReadInFilterList()

	// rmlayers := "\"" + utils.FtoStr("rmlayers.txt") + "\""

	log.Warn().Msgf("Running with %v workers", workers)

	// defArgs := {"GPWestFSTopo.pdf -D 750 -r \"$(cat rmlayers.txt)\" -t EPSG:3857"}
	for i := 0; i < workers; i++ {
		// go utils.CommandRunner(jobs, results, "./convert-geopdf.py", "-D", "750", "-r", "\"$(cat rmlayers.txt)\"", "-t", "EPSG:3857")
		go utils.PDF2TiffWorker(jobs, results, filterLayers, outDir, "gdalwarp", "-co", "TILED=YES", "-co", "TFW=YES", "-t_srs", "EPSG:3857", "-r", "near", "-overwrite", "-dstnodata", "255", "--config", "GDAL_PDF_DPI", dpi)
	}
	queueSources(sources, jobs)

	for i := 0; i < jobCount; i++ {
		var rst = <-results
		log.Debug().Msg(rst)
	}
	close(jobs)
	log.Warn().Msg("Done with all jobs")
}

func setupLogByLevel(level int) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	switch level {
	case 0:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case 1:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case 2:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case 3:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	default:
		break
	}
}

func queueSources(sources []string, jobs chan<- string) {
	for _, source := range sources {
		jobs <- source
	}
}
