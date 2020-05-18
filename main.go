package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"./utils"
	"github.com/cavaliercoder/grab"
)

// type DataSource struct {
// 	name          string
// 	url           string
// 	filenameInZip string
// 	species       []string
// 	attribution   string
// }

func main() {
	jobs := make(chan [2]string, 1000)
	results := make(chan string, 1000)
	jobCount := 8
	for i := 0; i < jobCount; i++ {
		go worker(jobs, results)
	}
	processDirs("./land-sources", jobs)
	close(jobs)

	for result := range results {
		fmt.Println(result)
	}

}

func worker(jobs <-chan [2]string, results chan<- string) {
	for job := range jobs {
		result := ""
		sourceDir := string(job[0])
		fName := string(job[1])
		fileFull := sourceDir + "/" + fName
		result += "\n-------------------\n"
		result += "Processing Source: " + fName + "\n"
		err := processSource(sourceDir, fName)
		if err != nil {
			result += "Source failed: " + fileFull + "\n"
			log.Print(err)
		} else {
			result += "Finished processing source: " + fName + "\n"
		}
		results <- result
	}
}

func processDirs(sourceDir string, jobs chan<- [2]string) {
	// results :=
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
				fmt.Println("Added job:", jobToAdd[0], jobToAdd[1])
			}

		}

	}
}

func processSource(dir string, file string) error {
	pathedFile := dir + "/" + file
	sourceJSON := fileToStr(pathedFile)
	sourceURL := getPropFromJSON("url", sourceJSON)
	dlPath := strings.Replace(dir, "land-sources", "generated", 1)
	dlFile, err := utils.DownloadFile(dlPath, sourceURL)
	if err != nil {
		return err
	}
	if filepath.Ext(dlFile) == ".zip" {
		if !dirExists(dlPath + "/" + getFnameOnly(dlFile)) {
			runCommand(true, "unzip", dlFile, "-d", dlPath+"/"+getFnameOnly(dlFile))
		}
		err := os.Remove(dlFile)
		if err != nil {
			return err
		}
	}
	fname := getPropFromJSON("filenameInZip", sourceJSON)
	switch filetype := getPropFromJSON("filetype", sourceJSON); filetype {
	case "geojson":
		processGeoJSON(dlFile, fname)
	case "shp":
		processShp(dlPath+"/"+getFnameOnly(dlFile), fname)
	}

	return nil
}

func processGeoJSON(path string, filename string) error {
	fmt.Println("Processing geoJson: " + path + ", " + filename)
	return nil
}

func processShp(path string, filename string) error {
	if filename == "" {
		shapefiles, err := utils.WalkMatch(path, "*.shp")
		if err != nil {
			return err
		}
		if len(shapefiles) > 1 {
			return errors.New("Multiple shapefiles in zip + none specified")
		}
		filename = filepath.Base(shapefiles[0])
	}
	// basepath := filepath.Dir(path)
	filename = getFnameOnly(filename)
	fileWithPath := path + "/" + filename
	geojson := fileWithPath + ".geojson"
	shapefile := fileWithPath + ".shp"
	labels := fileWithPath + "-labels.geojson"
	// fmt.Println(geojson + ", " + shapefile)

	_, err := runCommand(true, "ogr2ogr", "-f", "GeoJSON", "-t_srs", "crs:84", geojson, shapefile)
	if err != nil {
		return err
	}
	runAndWriteCommand(labels, "geojson-polygon-labels", "--label=polylabel", "--include-minzoom=6-11", geojson)
	_, err = runCommand(false, "tippecanoe", "-z11", "-o", fileWithPath+".mbtiles", geojson)
	if err != nil {
		return err
	}
	_, err = runCommand(false, "tippecanoe", "-z11", "-o", fileWithPath+"-labels.mbtiles", labels)
	if err != nil {
		return err
	}
	_, err = runCommand(false, "tile-join", "-o", fileWithPath+"-combined.mbtiles", fileWithPath+".mbtiles", fileWithPath+"-labels.mbtiles")
	if err != nil {
		return err
	}

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

func runAndWriteCommand(outName, cmdName string, args ...string) {
	fmt.Println("Running cmd: " + cmdName + " " + strings.Join(args, " ") + " > " + outName)
	cmd := exec.Command(cmdName, args...)

	outfile, err := os.Create(outName)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile

	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	cmd.Wait()
}

func runCommand(silent bool, cmd string, args ...string) (string, error) {
	// fmt.Println("Running cmd: " + cmd + " " + strings.Join(args, " "))
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		fmt.Println("Could not run cmd: " + cmd + " " + strings.Join(args, " "))
		return "", err
	}
	if !silent {
		fmt.Printf("%s\n", out)
	}
	return string(out), nil
}
