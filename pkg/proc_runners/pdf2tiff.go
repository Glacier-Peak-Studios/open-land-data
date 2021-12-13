package proc_runners

import (
	"github.com/rs/zerolog/log"
	"glacierpeak.app/openland/pkg/utils"
)

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

func queueSources(sources []string, jobs chan<- string) {
	for _, source := range sources {
		jobs <- source
	}
}

type PDF2TIFFExecutor struct {
	inDir        string
	outDir       string
	filterLayers []string
	dpi          string
	workers      int
}

func NewPDF2TIFFExecutor(inDir string, outDir string, filterLayers []string, dpi string, workers int) *PDF2TIFFExecutor {
	return &PDF2TIFFExecutor{inDir, outDir, filterLayers, dpi, workers}
}

func (m *PDF2TIFFExecutor) Value() interface{} {
	return &m
}

// run function to run this executor
func (e *PDF2TIFFExecutor) Run() {
	PDF2TIFF(e.inDir, e.outDir, e.filterLayers, e.dpi, e.workers)
}
