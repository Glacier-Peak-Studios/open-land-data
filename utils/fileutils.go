package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

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

// WalkMatch gets all files in root with specified pattern
func WalkMatch(root string, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

// CleanJob cleans up folders and .zip files in the target job's directory
func CleanJob(job string) error {
	outdir := strings.Replace(job, "land-sources", "generated", 1)
	fmt.Println("Cleaning job: ", outdir)
	// Remove zipfiles when done with job:
	zipfiles, err := WalkMatch(outdir, "*.zip")
	if err != nil {
		log.Print(err)
	}
	for i := 0; i < len(zipfiles); i++ {
		fmt.Println("Removing zipfile: " + zipfiles[i])
		err = os.Remove(zipfiles[i])
		folder := outdir + "/" + getFnameOnly(zipfiles[i])
		fmt.Println("Removing folder: ", folder)
		err = os.RemoveAll(folder)
		// err = os.Remove(zipfiles[i])
	}
	// Remove kmzfiles when done with job:
	kmzfiles, err := WalkMatch(outdir, "*.kmz")
	if err != nil {
		log.Print(err)
	}
	for i := 0; i < len(kmzfiles); i++ {
		fmt.Println("Removing kmzfile: " + kmzfiles[i])
		err = os.Remove(kmzfiles[i])
	}
	return err
}
