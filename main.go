package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"./utils"
	"github.com/cavaliercoder/grab"
)

func main() {
	sources, _ := utils.WalkMatch("./land-sources", "*.json")
	jobCount := len(sources)
	fmt.Println("Sources:", jobCount)
	// jobs := make(chan [2]string, jobCount)
	jobs := make(chan string, jobCount)
	results := make(chan string, jobCount)
	workerCount := 8
	for i := 0; i < workerCount; i++ {
		go worker2(jobs, results)
	}
	processDirs2("./land-sources", jobs)
	close(jobs)

	// for result := range results {
	// 	fmt.Println(result)
	// }
	for i := 0; i < jobCount*3; i++ {
		fmt.Println(<-results)
	}

}

func worker(jobs <-chan [2]string, results chan<- string) {
	for job := range jobs {
		sourceDir := string(job[0])
		fName := string(job[1])
		fileFull := sourceDir + "/" + fName
		logMsg(results, fName, "Processing Source")
		err := processSource(sourceDir, fName)
		if err != nil {
			logMsg(results, fileFull, "Source failed!")
			logMsg(results, fileFull, err.Error())
		} else {
			logMsg(results, fName, "Finished processing source")
			logMsg(results, fName, "Job done")
		}
	}
}

func worker2(jobs <-chan string, results chan<- string) {
	for job := range jobs {
		files, err := ioutil.ReadDir(job)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range files {
			if !f.IsDir() && filepath.Ext(f.Name()) == ".json" {
				// processDirs(job+"/"+f.Name(), jobs)
				err := processSource(job, f.Name())
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
		// Remove zipfiles when done with dir:
		// zipfiles, err := utils.WalkMatch(job, "*.zip")
		// for i := 0; i < len(zipfiles); i++ {
		// 	err = os.Remove(zipfiles[i])
		// }
		// if err != nil {
		// 	log.Print(err)
		// }
	}
}

func logMsg(results chan<- string, source, msg string) {
	toSend := source + ":: " + msg
	results <- toSend
}

func processDirs(sourceDir string, jobs chan<- [2]string) {
	files, err := ioutil.ReadDir(sourceDir)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if f.IsDir() {
			processDirs(sourceDir+"/"+f.Name(), jobs)
		} else {
			if filepath.Ext(f.Name()) == ".json" {
				jobToAdd := [2]string{sourceDir, f.Name()}
				jobs <- jobToAdd
				// fmt.Println("Added job:", jobToAdd[0], jobToAdd[1])
			}

		}

	}
}
func processDirs2(sourceDir string, jobs chan<- string) {
	files, err := ioutil.ReadDir(sourceDir)
	if err != nil {
		log.Fatal(err)
	}
	count := 0
	for _, f := range files {
		if f.IsDir() {
			count++
			processDirs2(sourceDir+"/"+f.Name(), jobs)
		}
	}
	if count == 0 {
		jobs <- sourceDir
	}
}

func processSource(dir string, file string) error {
	pathedFile := dir + "/" + file
	sourceJSON := fileToStr(pathedFile)
	sourceURL := getPropFromJSON("url", sourceJSON)
	dlPath := strings.Replace(dir, "land-sources", "generated", 1)
	dlurl, _ := url.Parse(sourceURL)
	dlFile := dlPath + "/" + filepath.Base(dlurl.Path)
	var err error
	// fmt.Println("Checking if dl exists: " + dlFile)
	if !fileExists(dlFile) {
		dlFile, err = utils.DownloadFile(dlPath, sourceURL)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("DL exists: " + dlFile)
	}
	if filepath.Ext(dlFile) == ".zip" {
		if !dirExists(dlPath + "/" + getFnameOnly(dlFile)) {
			_, err := runCommand(true, "unzip", "-j", dlFile, "-d", dlPath+"/"+getFnameOnly(dlFile))
			if err != nil {
				return err
			}
		}
		// if fileExists(dlFile) {
		// 	fmt.Println("Removing zip: " + dlFile)
		// 	err := os.Remove(dlFile)
		// 	if err != nil {
		// 		return err
		// 	}
		// }
	}
	fname := getPropFromJSON("filenameInZip", sourceJSON)
	switch filetype := getPropFromJSON("filetype", sourceJSON); filetype {
	case "geojson":
		return processGeoJSON(dlPath, getFnameOnly(dlFile))
	case "shp":
		return processShp(dlPath+"/"+getFnameOnly(dlFile), fname, getFnameOnly(file))
	default:
		return errors.New("Filetype not supported: " + filetype)
	}
}

func processGeoJSON(path, filename string) error {
	fileWithPath := path + "/" + getFnameOnly(filename)
	geojson := fileWithPath + ".geojson"
	geojsonLabels := fileWithPath + "-labels.geojson"
	mbtiles := fileWithPath + ".mbtiles"
	mbtilesLabels := fileWithPath + "-labels.mbtiles"
	combined := fileWithPath + "-combined.mbtiles"
	fmt.Println("Processing geoJson: " + geojson)
	var err error

	err = generateLabels(geojsonLabels, geojson)
	err = generateMBTiles(mbtiles, geojson)
	err = generateMBTiles(mbtilesLabels, geojsonLabels)
	err = combineMBTiles(combined, mbtiles, mbtilesLabels)
	if err != nil {
		return err
	}
	err = os.Remove(geojsonLabels)
	err = os.Remove(mbtiles)
	err = os.Remove(mbtilesLabels)
	return err
}

func generateLabels(newfile, geojson string) error {
	if !fileExists(geojson) {
		return errors.New("Cannot create label! geojson doesn't exist: " + geojson)
	}
	if !fileExists(newfile) {
		return runAndWriteCommand(newfile, "geojson-polygon-labels", "--label=polylabel", "--include-minzoom=6-11", geojson)
	}
	return nil
}

func generateMBTiles(newfile, geojson string) error {
	if !fileExists(geojson) {
		return errors.New("Cannot create mbtile! geojson doesn't exist: " + geojson)
	}
	if !fileExists(newfile + ".mbtiles") {
		_, err := runCommand(false, "tippecanoe", "-f", "-z11", "-o", newfile, geojson)
		return err
	}
	return nil
}

func combineMBTiles(newfile, mbtiles, mbtilesLabels string) error {
	if !fileExists(mbtiles) {
		return errors.New("Cannot join mbtiles! base mbtile doesn't exist: " + mbtiles)
	}
	if !fileExists(mbtilesLabels) {
		return errors.New("Cannot join mbtiles! labels mbtile doesn't exist: " + mbtilesLabels)
	}
	if !fileExists(newfile + ".mbtiles") {
		_, err := runCommand(false, "tile-join", "-f", "-o", newfile, mbtiles, mbtilesLabels)
		return err
	}
	return nil
}

func processShp(path, filename, fileOutName string) error {
	if filename == "" {
		shapefiles, err := utils.WalkMatch(path, "*.shp")
		if err != nil {
			return err
		}
		if len(shapefiles) > 1 {
			fmt.Println("shapefiles-in-dir: ", shapefiles)
			return errors.New("Multiple shapefiles in zip, none specified in source")
		}
		if len(shapefiles) == 0 {
			return errors.New("No shapefiles in folder: " + path)
		}
		filename = filepath.Base(shapefiles[0])
	}
	basepath := filepath.Dir(path)
	filename = getFnameOnly(filename)
	fileOutName = getFnameOnly(fileOutName)
	fileWithPath := path + "/" + filename
	geojson := basepath + "/" + fileOutName + ".geojson"
	shapefile := fileWithPath + ".shp"
	// fmt.Println(geojson + ", " + shapefile)
	fmt.Println("Processing shapefile: " + shapefile)

	_, err := runCommand(false, "ogr2ogr", "-f", "GeoJSON", "-t_srs", "crs:84", geojson, shapefile)
	if err != nil {
		return err
	}
	err = processGeoJSON(basepath, fileOutName)

	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func dirExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func getFnameOnly(file string) string {
	var filename = filepath.Base(file)
	var extension = filepath.Ext(filename)
	return filename[0 : len(filename)-len(extension)]
}

func downloadFile(path string, url string) string {
	err := os.MkdirAll(path, 0755)
	// Force unchecked certs
	if err != nil {
		log.Print(err)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := grab.NewClient()
	client.HTTPClient = &http.Client{Transport: tr}
	// Make Request
	req, _ := grab.NewRequest(path, url)
	resp := client.Do(req)
	if err := resp.Err(); err != nil {
		log.Print(err)
	} else {
		fmt.Println("Download saved to", resp.Filename)
	}
	return resp.Filename
}

func fileToStr(file string) string {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

func getPropFromJSON(prop string, strJSON string) string {
	var result map[string]interface{}
	json.Unmarshal([]byte(strJSON), &result)
	val, isStr := result[prop].(string)
	if isStr {
		return val
	}
	return ""
}

func runAndWriteCommand(outName, cmdName string, args ...string) error {
	// for i := 0; i < len(args); i++ {
	// 	args[i] = strings.Replace(args[i], " ", "\\ ", -1)
	// }
	// fmt.Println("Running cmd: " + cmdName + " " + strings.Join(args, " ") + " > " + outName)
	cmd := exec.Command(cmdName, args...)

	outfile, err := os.Create(outName)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile

	err = cmd.Start()
	if err != nil {
		fmt.Println("Error running cmd: " + cmdName + " " + strings.Join(args, " ") + " > " + outName)
		return err
	}
	cmd.Wait()
	return nil
}

func runCommand(silent bool, cmd string, args ...string) (string, error) {
	// for i := 0; i < len(args); i++ {
	// 	args[i] = strings.Replace(args[i], " ", "\\ ", -1)
	// }
	// fmt.Println("Running cmd: " + cmd + " " + strings.Join(args, " "))
	out, err := exec.Command(cmd, args...).Output()
	if !silent {
		fmt.Printf("%s\n", out)
	}
	if err != nil {
		fmt.Println("Command unsuccessful: " + cmd + " " + strings.Join(args, " "))
		return "", err
	}
	return string(out), nil
}
