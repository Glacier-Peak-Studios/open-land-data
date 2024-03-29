package main

import (
	"flag"

	"glacierpeak.app/openland/pkg/runners"
	utils "glacierpeak.app/openland/pkg/utils"
)

func main() {

	flag.Usage = utils.CliUsage("pdf2tiff",
		`This is a tool that runs gdal's 'pdf2tiff' command in a parallelized way. 
		It takes in a directory of PDF files and outputs a direcotry of
		converted TIFF files. You can specify the number of parallel instances of
		this conversion, as well as the DPI of the TIFF files.`)

	workersOpt := flag.Int("t", utils.GetDefaultWorkers(), "The number of concurrent jobs being processed")
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

	utils.CheckRequiredFlags("i", "o")

	utils.SetupLogByLevel(*verboseOpt)

	// log.Warn().Msgf("Searching sources dir: %v", *outDir)
	// sources, _ := utils.WalkMatch(*inDir, "*.pdf")

	// jobCount := len(sources)
	// jobs := make(chan string, jobCount)
	// results := make(chan string, jobCount)

	filterLayers := utils.ReadInFilterList(*filterFile)

	runners.PDF2TIFF(*inDir, *outDir, filterLayers, *dpi, *workersOpt)

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
