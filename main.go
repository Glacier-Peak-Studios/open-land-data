// package main

// import (
// 	"flag"
// 	"io/ioutil"
// 	"os"
// 	"path/filepath"

// 	log "github.com/Sirupsen/logrus"

// 	"./utils"
// )

// func main() {

// 	workersOpt := flag.Int("t", 4, "The number of concurrent jobs being processed")
// 	sourceDirOpt := flag.String("src", "./land-sources", "The root directory of the source files")
// 	cleanupOpt := flag.Bool("nc", false, "Don't clean up the zip files and folders in the generated directories")
// 	verboseOpt := flag.Int("v", 1, "Set the verbosity level:\n"+
// 		" 0 - Only prints error messages\n"+
// 		" 1 - Adds run specs and error details\n"+
// 		" 2 - Adds general progress info\n"+
// 		" 3 - Adds debug info and details more detail\n")
// 	startNewOpt := flag.Bool("f", false, "Force generation of all sources, overwriting those existing")
// 	flag.Parse()

// 	switch *verboseOpt {
// 	case 0:
// 		log.SetLevel(log.ErrorLevel)
// 		break
// 	case 1:
// 		log.SetLevel(log.WarnLevel)
// 		break
// 	case 2:
// 		log.SetLevel(log.InfoLevel)
// 		break
// 	case 3:
// 		log.SetLevel(log.DebugLevel)
// 		log.SetReportCaller(true)
// 		break
// 	default:
// 		break
// 	}

// 	if *startNewOpt {
// 		err := os.RemoveAll("./generated")
// 		if err != nil {
// 			log.Warn(err)
// 		}
// 	}

// 	log.Warn("Searching sources dir: ", *sourceDirOpt)
// 	sources, _ := utils.WalkMatch(*sourceDirOpt, "*.json")
// 	jobCount := len(sources)
// 	jobs := make(chan string, jobCount)
// 	results := make(chan string, jobCount)

// 	log.Warn("Sources found: ", jobCount)
// 	log.Warn("Running with ", *workersOpt, " workers")
// 	for i := 0; i < *workersOpt; i++ {
// 		go worker(jobs, results, !(*cleanupOpt))
// 	}
// 	queueSources(*sourceDirOpt, jobs)
// 	close(jobs)

// 	for i := 0; i < jobCount*2; i++ {
// 		var rst = <-results
// 		log.Debug(rst)
// 	}
// 	log.Warn("Done with all jobs")

// }

// func worker(jobs <-chan string, results chan<- string, cleanUp bool) {
// 	for job := range jobs {
// 		files, err := ioutil.ReadDir(job)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		for _, f := range files {
// 			if !f.IsDir() && filepath.Ext(f.Name()) == ".json" {
// 				err := utils.ProcessSource(job, f.Name())
// 				fileFull := job + "/" + f.Name()
// 				log.Info(f.Name(), ": Processing Source")
// 				if err != nil {
// 					logMsg(results, fileFull, " Source failed!")
// 					log.Error(fileFull, ": ", err.Error())
// 				} else {
// 					log.Info(f.Name(), ": Finished processing source")
// 					logMsg(results, f.Name(), " Job done")
// 				}
// 			}

// 		}
// 		if cleanUp {
// 			utils.CleanJob(job)
// 		}
// 	}
// }

// func logMsg(results chan<- string, source, msg string) {
// 	toSend := source + ": " + msg
// 	results <- toSend
// }

// func queueSources(sourceDir string, jobs chan<- string) {
// 	files, err := ioutil.ReadDir(sourceDir)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	count := 0
// 	for _, f := range files {
// 		if f.IsDir() {
// 			count++
// 			queueSources(sourceDir+"/"+f.Name(), jobs)
// 		}
// 	}
// 	if count == 0 {
// 		jobs <- sourceDir
// 	}
// }
