package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	utils "glacierpeak.app/openland/pkg/utils"
)

func main() {
	flag.Usage = utils.CliUsage("land_sources",
		`This tool reads sources from a directory structure of .json files.
		On generation, it copies the sources directory structure, and each .json becomes a useable generated file.
		By default, no sources are included in this project. To obtain sources to convert:
		[(1) For the last stable set of sources, visit the json sources repo. This can be dropped into this project at the root level.
		(2) For the very latest sources (still being audited), visit the sources doc, download as a .csv, and convert to json.]
		
		To convert sources from .csv to json, and vise versa, run ./sourcemgr.py`)

	workersOpt := flag.Int("t", utils.GetDefaultWorkers(), "The number of concurrent jobs being processed")
	sourceDirOpt := flag.String("src", "./land-sources", "The root directory of the source files")
	cleanupOpt := flag.Bool("nc", false, "Don't clean up the zip files and folders in the generated directories")
	verboseOpt := flag.Int("v", 1, "Set the verbosity level:\n"+
		" 0 - Only prints error messages\n"+
		" 1 - Adds run specs and error details\n"+
		" 2 - Adds general progress info\n"+
		" 3 - Adds debug info and details more detail\n")
	startNewOpt := flag.Bool("f", false, "Force generation of all sources, overwriting those existing")
	flag.Parse()

	utils.SetupLogByLevel(*verboseOpt)

	if *startNewOpt {
		err := os.RemoveAll("./generated")
		if err != nil {
			log.Warn().Err(err).Msg("")
		}
	}

	log.Warn().Msgf("Searching sources dir: %v", *sourceDirOpt)
	sources, _ := utils.WalkMatch(*sourceDirOpt, "*.json")
	jobCount := len(sources)
	jobs := make(chan string, jobCount)
	results := make(chan string, jobCount)

	log.Warn().Msgf("Sources found: %v", jobCount)
	log.Warn().Msgf("Running with %v workers", *workersOpt)
	for i := 0; i < *workersOpt; i++ {
		go worker(jobs, results, !(*cleanupOpt))
	}
	queueSources(*sourceDirOpt, jobs)
	close(jobs)

	for i := 0; i < jobCount*2; i++ {
		var rst = <-results
		log.Debug().Msg(rst)
	}
	log.Warn().Msg("Done with all jobs")

}

func worker(jobs <-chan string, results chan<- string, cleanUp bool) {
	for job := range jobs {
		files, err := os.ReadDir(job)
		if err != nil {
			log.Fatal().Err(err).Msg("")
		}
		for _, f := range files {
			if !f.IsDir() && filepath.Ext(f.Name()) == ".json" {
				err := utils.ProcessSource(job, f.Name())
				fileFull := job + "/" + f.Name()
				log.Info().Msg(f.Name() + ": Processing Source")
				if err != nil {
					logMsg(results, fileFull, " Source failed!")
					log.Error().Msg(fileFull + ": " + err.Error())
				} else {
					log.Info().Msg(f.Name() + ": Finished processing source")
					logMsg(results, f.Name(), " Job done")
				}
			}

		}
		if cleanUp {
			utils.CleanJob(job)
		}
	}
}

func logMsg(results chan<- string, source, msg string) {
	toSend := source + ": " + msg
	results <- toSend
}

func queueSources(sourceDir string, jobs chan<- string) {
	files, err := os.ReadDir(sourceDir)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	count := 0
	for _, f := range files {
		if f.IsDir() {
			count++
			queueSources(sourceDir+"/"+f.Name(), jobs)
		}
	}
	if count == 0 {
		jobs <- sourceDir
	}
}
