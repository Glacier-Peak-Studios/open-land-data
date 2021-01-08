package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	log "github.com/Sirupsen/logrus"

	"./utils"
)

const fstopo_arc = "https://apps.fs.usda.gov/arcx/rest/services/EDW/EDW_FSTopo_01/MapServer/tile"

func main() {

	workersOpt := flag.Int("t", 4, "The number of concurrent jobs being processed")
	sourceDirOpt := flag.String("src", "./land-sources", "The root directory of the source files")
	// cleanupOpt := flag.Bool("nc", false, "Don't clean up the zip files and folders in the generated directories")
	verboseOpt := flag.Int("v", 1, "Set the verbosity level:\n"+
		" 0 - Only prints error messages\n"+
		" 1 - Adds run specs and error details\n"+
		" 2 - Adds general progress info\n"+
		" 3 - Adds debug info and details more detail\n")
	startNewOpt := flag.Bool("f", false, "Force generation of all sources, overwriting those existing")
	flag.Parse()

	switch *verboseOpt {
	case 0:
		log.SetLevel(log.ErrorLevel)
		break
	case 1:
		log.SetLevel(log.WarnLevel)
		break
	case 2:
		log.SetLevel(log.InfoLevel)
		break
	case 3:
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
		break
	default:
		break
	}

	if *startNewOpt {
		err := os.RemoveAll("./generated")
		if err != nil {
			log.Warn(err)
		}
	}

	log.Warn("Searching sources dir: ", *sourceDirOpt)
	sources, _ := utils.WalkMatch(*sourceDirOpt, "*.png")
	sources = Filter(sources, isEvenTile)

	jobCount := len(sources)
	jobs := make(chan string, jobCount)
	results := make(chan string, jobCount)

	log.Warn("Sources found: ", jobCount)
	// for _, source := range sources {
	// 	log.Info("Source: ", source)
	// }
	log.Warn("Running with ", *workersOpt, " workers")
	for i := 0; i < *workersOpt; i++ {
		go worker2(jobs, results)
	}
	queueSources(sources, jobs)
	close(jobs)

	for i := 0; i < jobCount*2; i++ {
		var rst = <-results
		log.Debug(rst)
	}
	log.Warn("Done with all jobs")

}

func worker(jobs <-chan string, results chan<- string) {
	for job := range jobs {
		var err error = nil
		// print(job)
		// curDir := filepath.Dir(job)
		y := utils.StripExt(filepath.Base(job))

		fdir := filepath.Dir(job)
		x := filepath.Base(fdir)
		fdir = filepath.Dir(fdir)
		z := filepath.Base(fdir)
		fdir = filepath.Dir(fdir)

		// zxy := z + "/" + x + "/" + y
		zx := z + "/" + x

		// baseFolder := fdir + "/" + zx
		topoFolder := fdir + "-topo/" + zx
		outFolder := fdir + "-merged/" + zx
		baseImg := job
		topoImg := topoFolder + "/" + y + "-topo.png"
		outImg := outFolder + "/" + y + ".png"
		// fmt.Printf("(x, y, z) - (%v, %v, %v)\n", x, y, z)

		// if fileExists(baseImg) {
		// 	os.Remove(job)
		// 	os.Rename(baseImg, job)
		// }

		vecturl := fmt.Sprintf("%v/%v/%v/%v", fstopo_arc, z, y, x)
		// println(vecturl)

		topoDownloaded := true
		if !fileExists(topoImg) {
			dlTopoFolder := topoFolder + "/" + y + "-temp"
			// print("DLLoc:", dlLoc)
			_, err = utils.DownloadFile(dlTopoFolder, vecturl)
			if err != nil {
				log.Error("Failed to download file - ", err.Error())
				os.Remove(dlTopoFolder)
				topoDownloaded = false
				os.Link(baseImg, outImg)
			} else {
				dlTopo := dlTopoFolder + "/" + x
				os.Rename(dlTopo, topoImg)
				os.Remove(dlTopoFolder)
			}
		} else {
			log.Info("Topo tile already downloaded, using cached version")
		}
		if topoDownloaded {

			err = os.MkdirAll(filepath.Dir(outImg), 0755)
			if err != nil {
				log.Error("Failed to create output dir - ", err.Error())
			} else {
				err = utils.CombineImages(baseImg, topoImg, outImg)
			}

		}
		if err != nil {
			logMsg(results, job, err.Error())
		} else {
			logMsg(results, outImg, "Job done")
		}

	}
}

func worker2(jobs <-chan string, results chan<- string) {
	for job := range jobs {
		var err error = nil
		// print(job)
		// curDir := filepath.Dir(job)
		yStr := utils.StripExt(filepath.Base(job))

		fdir := filepath.Dir(job)
		xStr := filepath.Base(fdir)
		fdir = filepath.Dir(fdir)
		zStr := filepath.Base(fdir)
		fdir = filepath.Dir(fdir)
		basepath := fdir

		x, err := strconv.Atoi(xStr)
		y, err := strconv.Atoi(yStr)
		z, err := strconv.Atoi(zStr)
		// yNum, err := strconv.Atoi(yStr)
		img1 := filepath.Join(basepath, zStr, xStr, yStr+".png")
		img2 := filepath.Join(basepath, zStr, strconv.Itoa(x+1), yStr+".png")
		img3 := filepath.Join(basepath, zStr, xStr, strconv.Itoa(y+1)+".png")
		img4 := filepath.Join(basepath, zStr, strconv.Itoa(x+1), strconv.Itoa(y+1)+".png")

		imgOut := filepath.Join(basepath, strconv.Itoa(z-1), strconv.Itoa(x/2), strconv.Itoa(y/2)+".png")

		err = utils.GenerateOverview(imgOut, img1, img2, img3, img4)

		if err != nil {
			logMsg(results, job, err.Error())
		} else {
			logMsg(results, imgOut, "Job done")
		}

	}
}

// func pathBuilder(args ...string) string {
// 	var sb strings.Builder
// 	sb.WriteString(args[0])
// 	for _, val := range args[1:] {
// 		sb.WriteString("/")
// 		sb.WriteString(val)
// 	}
// 	return sb.String()
// }

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
		log.Error("Could not parse string to int")
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
