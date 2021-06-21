package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
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

func DirEx(filename string) bool {
	return dirExists(filename)
}

func getFnameOnly(file string) string {
	var filename = filepath.Base(file)
	var extension = filepath.Ext(filename)
	return filename[0 : len(filename)-len(extension)]
}

func StripExt(file string) string {
	var filename = filepath.Base(file)
	var extension = filepath.Ext(filename)
	return filename[0 : len(filename)-len(extension)]
}

func fileToStr(file string) string {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	return string(content)
}

func FtoStr(file string) string {
	return fileToStr(file)
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
	log.Debug().Msgf("Searching dir %v for pattern %v", root, pattern)
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

func WalkRecursive(root string, workers int) []string {
	jobs := make(chan string, 500)
	filesRet := make(chan string, 200)

	var workersDone uint64
	dirCount := uint64(1)
	workersTotal := uint64(workers)

	var tileList []string

	for i := 0; i < workers; i++ {
		go FileFinder(jobs, filesRet, &workersDone, workersTotal, &dirCount)
	}

	jobs <- root

	for tileFile := range filesRet {
		tileList = append(tileList, tileFile)
	}

	return tileList
}

func GetAllTiles2(root string, workers int) []string {
	dirsList, err := ioutil.ReadDir(root)
	if err != nil {
		log.Error().Err(err).Msg("Error reading sources list")
	}

	jobCount := len(dirsList)
	jobs := make(chan string, jobCount)
	filesRet := make(chan string, 200)

	var workersDone uint64
	workersTotal := uint64(workers)

	var tileList []string

	for i := 0; i < workers; i++ {
		go TilesetListWorkerStreamed(jobs, filesRet, &workersDone, workersTotal)
	}

	for _, dir := range dirsList {
		if dir.IsDir() {
			jobs <- filepath.Join(root, dir.Name())
		}
	}
	close(jobs)

	for tileFile := range filesRet {
		tileList = append(tileList, tileFile)
	}

	return tileList

}

func GetAllTilesStreamed(root string, workerCount int, foundTiles chan<- string) {
	dirsList, err := ioutil.ReadDir(root)
	if err != nil {
		log.Error().Err(err).Msg("Error reading sources list")
	}

	searchDirs := make(chan string, 500)
	// filesRet := make(chan string, 200)

	var workersDone uint64
	workersTotal := uint64(workerCount)

	// var tileList []string

	for i := 0; i < workerCount; i++ {
		go TilesetListWorkerStreamed(searchDirs, foundTiles, &workersDone, workersTotal)
	}

	for _, dir := range dirsList {
		if dir.IsDir() {
			xlist, err := ioutil.ReadDir(filepath.Join(root, dir.Name()))
			if err != nil {
				log.Error().Err(err).Msg("Error reading sources list")
			}
			for _, x := range xlist {
				if x.IsDir() {
					searchDirs <- filepath.Join(root, dir.Name(), x.Name())
				}
			}
		}
	}
	close(searchDirs)

	// for tileFile := range filesRet {
	// 	tileList = append(tileList, tileFile)
	// }

	// return tileList

}

func GetAllTiles(root string, workers int) (map[string][]string, []Tile) {
	dirsList, err := ioutil.ReadDir(root)
	if err != nil {
		log.Error().Err(err).Msg("Error reading sources list")
	}

	jobCount := len(dirsList)
	log.Debug().Msgf("Making job channel of length %v", jobCount)
	jobs := make(chan string, jobCount)
	filesRet := make(chan string, 30)

	var workersDone uint64
	workersTotal := uint64(workers)

	// var tileList []string

	for i := 0; i < workers; i++ {
		go TilesetListWorker(jobs, filesRet, &workersDone, workersTotal)
	}

	for _, dir := range dirsList {
		if dir.IsDir() {
			jobs <- filepath.Join(root, dir.Name(), "17")
		}
	}
	log.Debug().Msg("Done queing folder list, closing channel")
	close(jobs)

	m := make(map[string][]string)
	// m := make(map[int][]string)
	var tileList []Tile

	for tileFile := range filesRet {
		tile, base := PathToTile(tileFile)
		tileXY := tile.GetPathXY()
		// tXYInt := tile.GetXYInt()
		tSources := m[tileXY]
		// tSources := m[tXYInt]
		if tSources == nil {
			tileList = append(tileList, tile)
			lenTL := len(tileList)
			if lenTL % 10000000 == 0 {
				log.Debug().Msgf("Length of tileList is now %v", lenTL)
			}
		}
		tSources = append(tSources, base)
		m[tileXY] = tSources
		// m[tXYInt] = tSources
		lenMap := len(m)
		if lenMap % 100000 == 0 {
			log.Debug().Msgf("Length of map is now %v", lenMap)
		}
	}

	return m, tileList

}

func GetAllTiles0(root string, zLvl string, workers int) (map[int][]string, []Tile) {
	dirsList, err := ioutil.ReadDir(root)
	if err != nil {
		log.Error().Err(err).Msg("Error reading sources list")
	}
	gatherTilesBar := progressbar.NewOptions(-1, 
		progressbar.OptionSetDescription("Gathering tiles to merge"),
		progressbar.OptionSetItsString("tiles"),
		progressbar.OptionShowIts(),
		progressbar.OptionSpinnerType(14),	

	)

	

	jobCount := len(dirsList)
	log.Debug().Msgf("Making job channel of length %v", jobCount)
	jobs := make(chan string, jobCount)
	filesRet := make(chan string, 30)

	var workersDone uint64
	workersTotal := uint64(workers)

	for i := 0; i < workers; i++ {
		go TilesetListWorker(jobs, filesRet, &workersDone, workersTotal)
	}

	for _, dir := range dirsList {
		if dir.IsDir() {
			jobs <- filepath.Join(root, dir.Name(), zLvl)
		}
	}
	log.Debug().Msg("Done queing folder list, closing channel")
	close(jobs)

	m := make(map[int][]string)
	var tileList []Tile

	for tileFile := range filesRet {
		tile, base := PathToTile(tileFile)
		savedBase := strings.ReplaceAll(base, root, "")
		tXYInt := tile.GetXYInt()
		tSources := m[tXYInt]
		if tSources == nil {
			tileList = append(tileList, tile)
			lenTL := len(tileList)
			if lenTL % 10000000 == 0 {
				log.Debug().Msgf("Number of tiles is now %v", lenTL)
			}
		}
		tSources = append(tSources, savedBase)
		m[tXYInt] = tSources
		gatherTilesBar.Add(1)
	}

	gatherTilesBar.Finish()
	return m, tileList

}

// CleanJob cleans up folders and .zip files in the target job's directory
func CleanJob(job string) error {
	outdir := strings.Replace(job, "land-sources", "generated", 1)
	log.Info().Msgf("Cleaning job: %v", outdir)
	zipfiles, err := WalkMatch(outdir, "*.zip")
	if err != nil {
		log.Error().Err(err).Msg("")
	}
	for i := 0; i < len(zipfiles); i++ {
		log.Debug().Msg("Removing zipfile: " + zipfiles[i])
		err = os.Remove(zipfiles[i])
		folder := outdir + "/" + getFnameOnly(zipfiles[i])
		log.Debug().Msgf("Removing folder: %v", folder)
		err = os.RemoveAll(folder)
	}
	kmzfiles, err := WalkMatch(outdir, "*.kmz")
	if err != nil {
		log.Print(err)
	}
	for i := 0; i < len(kmzfiles); i++ {
		log.Debug().Msg("Removing kmzfile: " + kmzfiles[i])
		err = os.Remove(kmzfiles[i])
	}
	return err
}

func BBoxFromTileset(path string) (BBox, error) {
	xrange, err := ioutil.ReadDir(path)
	if err != nil {
		log.Error().Msg("Couldn't read source dir")
		return ZeroBBox(), err
	}
	x0 := xrange[0].Name()
	x1 := xrange[len(xrange)-1].Name()

	x0Path := filepath.Join(path, x0)
	x1Path := filepath.Join(path, x1)
	x0ListY, err := ioutil.ReadDir(x0Path)
	x1ListY, err := ioutil.ReadDir(x1Path)
	if err != nil {
		log.Error().Msg("Couldn't read source dir")
		return ZeroBBox(), err
	}
	if len(x0ListY) == 0 || len(x1ListY) == 0 {
		log.Error().Msg("Couldn't read source dir")
		return ZeroBBox(), err
	}
	y0 := strings.Replace(x0ListY[0].Name(), ".png", "", 1)
	y1 := strings.Replace(x1ListY[len(x1ListY)-1].Name(), ".png", "", 1)

	tsOrigin, _ := NewPoint(x0, y0)
	tsExtent, _ := NewPoint(x1, y1)

	return BBx(tsOrigin, tsExtent), nil
}

func CleanBBoxEdge(b BBox, side string, basepath string, zoom int) {
	// var err error = nil
	sideNum := SideToNum(side)
	for ix := b.Origin().X; ix <= b.Extent().X; ix++ {
		for iy := b.Origin().Y; iy <= b.Extent().Y; iy++ {
			tile := Tile{X: ix, Y: iy, Z: zoom}
			imgFile := filepath.Join(basepath, tile.GetPath()+".png")
			CleanTileEdge(imgFile, sideNum)
			// err = os.Remove(imgFile)
		}
	}
}

// func GetTrimBBox(file string, curBBox BBox) BBox {

// }

func GetGeoPDFLayers(file string) []string {
	out, err := RunCommand(true, "gdalinfo", "-mdd", "LAYERS", file)
	log.Err(err).Msg("Quering gdalinfo for layers")
	lines := strings.Split(out, "\n")
	var layers []string
	for _, line := range lines {
		if strings.Contains(line, "LAYER_") {
			layer := strings.Split(line, "=")[1]
			layers = append(layers, layer)
		}
	}
	// println(lines)
	return layers
}

func ReadInFilterList(file string) []string {
	dat, err := ioutil.ReadFile(file)
	if (err != nil) {
		fmt.Print(err)
	}
  return strings.Fields(string(dat))
	// return {""}
}

func removeTilesInBBox(b BBox, basepath string, z int) error {
	var err error = nil
	for ix := b.Origin().X; ix <= b.Extent().X; ix++ {
		for iy := b.Origin().Y; iy <= b.Extent().Y; iy++ {
			tile := Tile{X: ix, Y: iy, Z: z}
			imgFile := filepath.Join(basepath, tile.GetPath()+".png")
			err = os.Remove(imgFile)
			if err != nil {
				log.Error().Err(err).Msg("Failed to remove file")
			}
		}
	}
	return err
}

func IsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
