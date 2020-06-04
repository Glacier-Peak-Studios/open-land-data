package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"./utils"
)

func main() {
	sources, _ := utils.WalkMatch("./land-sources", "*.json")
	jobCount := len(sources)
	fmt.Println("Sources:", jobCount)
	jobs := make(chan string, jobCount)
	results := make(chan string, jobCount)
	workerCount := 8
	// workerCount := 1
	for i := 0; i < workerCount; i++ {
		go worker(jobs, results)
	}
	queueSources("./land-sources", jobs)
	close(jobs)

	// for result := range results {
	// 	fmt.Println(result)
	// }
	for i := 0; i < jobCount*3; i++ {
		fmt.Println(<-results)
	}

}

func worker(jobs <-chan string, results chan<- string) {
	for job := range jobs {
		files, err := ioutil.ReadDir(job)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range files {
			if !f.IsDir() && filepath.Ext(f.Name()) == ".json" {
				err := utils.ProcessSource(job, f.Name())
				fileFull := job + "/" + f.Name()
				logMsg(results, f.Name(), "Processing Source")
				if err != nil {
					logMsg(results, fileFull, "Source failed!")
					logMsg(results, fileFull, err.Error())
				} else {
					logMsg(results, f.Name(), "Finished processing source")
					logMsg(results, f.Name(), "Job done")
				}
			}

		}
		// utils.CleanJob(job)
	}
}

func logMsg(results chan<- string, source, msg string) {
	toSend := source + ":: " + msg
	results <- toSend
}

func queueSources(sourceDir string, jobs chan<- string) {
	files, err := ioutil.ReadDir(sourceDir)
	if err != nil {
		log.Fatal(err)
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
