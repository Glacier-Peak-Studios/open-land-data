package utils

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/rs/zerolog/log"
)

func PDF2TiffWorker(jobs <-chan string, results chan<- string, outDir string, cmd string, constArgs ...string) {
	for job := range jobs {
		// jobSplit := strings.Split(job, " ")
		println("-> Job -", job)
		pdfLayers := GetGeoPDFLayers(job)

		pdfLayers = Filter(pdfLayers, RemoveLayer)
		rmLayers := strings.Join(pdfLayers[:], ",")
		args := append(constArgs, "--config", "GDAL_PDF_LAYERS_OFF", rmLayers)

		fout := filepath.Join(outDir, StripExt(job)+".tif")
		args = append(args, job, fout)
		// constArgs = append(constArgs, job, fout)
		println("Going to run::")
		argList := strings.Join(args, " ")
		println(cmd, argList)
		if !fileExists(fout) {
			out, err := RunCommand2(true, false, cmd, args...)
			log.Err(err).Msg(out)
		} else {
			log.Info().Msg("File exists, skipping")
		}
		logMsg(results, job, "Done")
		// time.Sleep(10 * time.Second)
	}

}

func LayerFilter(layer string) bool {
	if strings.HasPrefix(layer, "Quadrangle.Neatline") {
		return false
	}
	if strings.HasPrefix(layer, "Quadrangle.2_5") {
		return false
	}
	if strings.HasPrefix(layer, "Quadrangle_Ext") {
		return false
	}
	if strings.HasPrefix(layer, "Adjacent") {
		return false
	}
	if strings.HasPrefix(layer, "Other") {
		return false
	}
	if strings.HasPrefix(layer, "Quadrangle.UTM") {
		return false
	}
	if strings.HasPrefix(layer, "Ownership") {
		return false
	}

	return true
}

func RemoveLayer(layer string) bool {
	return !LayerFilter(layer)
}

func OverviewWorker(jobs <-chan string, results chan<- string) {
	for job := range jobs {
		var err error = nil

		curTile, basepath := PathToTile(job)

		img1 := filepath.Join(basepath, curTile.getPath())
		tRight := curTile.rightTile()
		img2 := filepath.Join(basepath, tRight.getPath())
		tDown := curTile.downTile()
		img3 := filepath.Join(basepath, tDown.getPath())
		tDiag := tRight.downTile()
		img4 := filepath.Join(basepath, tDiag.getPath())

		tOver := curTile.overviewTile()
		imgOut := filepath.Join(basepath, tOver.getPath())

		if !fileExists(imgOut) {

			err = GenerateOverviewTile(imgOut, img1, img2, img3, img4)

			if err != nil {
				logMsg(results, job, err.Error())
			} else {
				logMsg(results, imgOut, "Job done")
			}
		} else {
			logMsg(results, imgOut, "Out img already exists, skipping. Done.")
		}

	}
}

func TilesetMergeWorker(jobs <-chan string, results chan<- string, ts1Dir string, ts2Dir string) {
	for job := range jobs {
		curTile, _ := PathToTile(job)
		outImg := job
		ts1File := ts1Dir + "/" + curTile.getPath() + ".png"
		ts2File := ts2Dir + "/" + curTile.getPath() + ".png"
		os.MkdirAll(filepath.Dir(outImg), 0755)

		ts1Ex := fileExists(ts1File)
		ts2Ex := fileExists(ts2File)

		if ts1Ex && ts2Ex {
			MergeTiles(ts1File, ts2File, outImg)
			logMsg(results, job, "Merging tiles")
		} else {
			tileCopy := ""
			if ts1Ex {
				tileCopy = ts1File
			} else {
				tileCopy = ts2File
			}
			os.Link(tileCopy, outImg)
			// os.Rename(tileCopy, outImg)

			logMsg(results, job, "Copying tile")
		}
	}
}

func TilesetMergeWorker2(jobs <-chan Tile, results chan<- string, locations map[string][]string, outDir string) {
	for job := range jobs {
		curTile := job
		outImg := filepath.Join(outDir, curTile.getPath()) + ".png"
		if !fileExists(outImg) {

			os.MkdirAll(filepath.Dir(outImg), 0755)

			tileLocs := locations[curTile.GetPathXY()]

			vsf := make([]string, 0)
			for _, v := range tileLocs {
				vF := appendTileToBase(v, curTile) + ".png"
				vsf = append(vsf, vF)
			}
			tileLocs = vsf

			var err error = nil
			if len(tileLocs) > 1 {
				err = MergeNTiles(tileLocs, outImg)
			} else if len(tileLocs) == 1 {
				err = os.Link(tileLocs[0], outImg)
			}
			if err != nil {
				log.Error().Err(err).Msgf("Error creating output tile: %v", outImg)
			}
			logMsg(results, outImg, "Done.")
		} else {
			logMsg(results, outImg, "Already exists, done.")
		}
	}
}

func FixBackgroundWorker(jobs <-chan Tile, results chan<- string, validTiles map[string]bool, inDir string, outDir string) {
	for job := range jobs {
		curTile := job
		imgInPath := filepath.Join(inDir, curTile.getPath()+".png")
		imgOutPath := filepath.Join(outDir, curTile.getPath())
		if !fileExists(imgOutPath) {
			os.MkdirAll(filepath.Dir(imgOutPath), 0755)

			// tileLocs := validTiles[curTile.GetPathXY()]
			surround := make([][]bool, 3)
			for i := range surround {
				surround[i] = make([]bool, 3)
			}
			for x := -1; x < 2; x++ {
				for y := -1; y < 2; y++ {
					tmpTile := MakeTile(curTile.x+x, curTile.y+y, curTile.z)
					surround[y+1][x+1] = validTiles[tmpTile.GetPathXY()]
				}
			}
			imgIn, err := DecodePNGFromPath(imgInPath)
			if err != nil {
				log.Error().Err(err).Msg("Failed to get image")
				logMsg(results, imgOutPath, "Failed")
			} else {
				missingEdges := getMissingEdges(surround)
				var imgOut image.Image
				var bgRects []image.Rectangle

				if len(missingEdges) == 0 {
					missingCorner := getMissingCorners(surround)
					if missingCorner > -1 {
						// println(imgInPath)
						bgRects, _ = GetCoverageRectCorner(imgIn, missingCorner)
					} else {
						bgRects = append(bgRects, image.Rect(0, 0, 256, 256))
					}
				} else {
					rect, _ := GetCoverageRectSide(imgIn, missingEdges[0])
					if len(missingEdges) > 1 {
						rect2, _ := GetCoverageRectSide(imgIn, missingEdges[1])
						rect = rect.Intersect(rect2)
					}
					bgRects = append(bgRects, rect)
				}

				imgOut = ImgOverRects(imgIn, bgRects)

				err := EncodePNGToPath(imgOutPath, imgOut)
				if err != nil {
					log.Error().Err(err).Msgf("Error creating output tile: %v", imgOutPath)
				}
				logMsg(results, imgOutPath, "Done")
			}
		} else {
			logMsg(results, imgOutPath, "Out img already exists. Done")
		}
	}
}

func getMissingEdges(surround [][]bool) []int {
	// var missingEdges []Point
	var missingEdges []int
	// vals :=
	for _, x := range []int{-1, 1} {
		if !surround[1][x+1] {
			missingEdges = append(missingEdges, coordToSideNum(x, 0))
		}
	}
	for _, y := range []int{-1, 1} {
		if !surround[y+1][1] {
			missingEdges = append(missingEdges, coordToSideNum(0, y))
		}
	}
	return missingEdges
}

func coordToSideNum(x, y int) int {
	return (x+1)/2 + (x+2+(y+1)/2)*AbsInt(y)
}

func getMissingCorners(surround [][]bool) int {
	// var missingCorners []Point
	// var missingCorners []int
	// vals :=
	for _, x := range []int{-1, 1} {
		for _, y := range []int{-1, 1} {
			if !surround[y+1][x+1] {
				return coordToCornerNum(x, y)
				// missingCorners = append(missingCorners, coordToCornerNum(x, y))
			}
		}
	}
	// return missingCorners
	return -1
}

func coordToCornerNum(x, y int) int {
	return (x+y+2)/4*AbsInt(x+y)/2 + (5+x)/2*(1-x*y)/2
}

func appendTileToBase(base string, tile Tile) string {
	return filepath.Join(base, tile.getPath())
}

func TileTrimWorker(jobs <-chan string, results chan<- string, zoom int) {
	for job := range jobs {
		println(job)
		basepath := job
		workingPath := filepath.Join(basepath, strconv.Itoa(zoom))
		bbx, err := BBoxFromTileset(workingPath)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create bbox")
			break
		}

		sides := [4]string{"left", "right", "top", "bottom"}

		for _, side := range sides {
			checkLine := bbx.getSideLine(side)
			toRemove := BBx(bbx.Origin(), bbx.Origin())
			if side == "right" || side == "bottom" {
				toRemove = BBx(bbx.Extent(), bbx.Extent())
			}
			// if side == "top" {
			// 	print("top")
			// }
			isWhite := checkLine.isBBoxWhite(basepath, zoom)
			for ; isWhite; isWhite = checkLine.isBBoxWhite(basepath, zoom) {
				toRemove, _ = GetBBoxMerge(toRemove, checkLine)
				bbx.ChangeSide(side, -1)
				checkLine = bbx.getSideLine(side)
			}
			removeTilesInBBox(toRemove, basepath, zoom)
			// CleanBBoxEdge(checkLine, side, basepath, zoom)

		}
		lvlDirs, _ := ioutil.ReadDir(workingPath)
		for _, dir := range lvlDirs {
			path := filepath.Join(workingPath, dir.Name())
			empty, _ := IsEmpty(path)
			if empty {
				os.Remove(path)
			}
		}
		logMsg(results, job, "- Done")
	}
}

func TilesetListWorker(jobs <-chan string, results chan<- string, workersDone *uint64, workerCount uint64) {
	// defer wg.Done()
	for job := range jobs {
		dirWithZ := job
		xList, err := ioutil.ReadDir(dirWithZ)
		if err != nil {
			log.Error().Err(err).Msgf("Could not read z dir: %v", dirWithZ)
		} else {
			// var tileList []string
			for _, xDir := range xList {
				if xDir.IsDir() {
					tiles, err := ioutil.ReadDir(filepath.Join(dirWithZ, xDir.Name()))
					if err != nil {
						log.Error().Err(err).Msgf("Could not read x dir: %v", dirWithZ)
					} else {
						for _, tile := range tiles {
							fname := tile.Name()
							if filepath.Base(fname) != ".DS_Store" {
								fullFilePath := filepath.Join(dirWithZ, xDir.Name(), fname)
								// println("Adding file to results", fullFilePath)
								results <- fullFilePath
							}
							// tileList = append(tileList, tile.Name())
						}
					}
				}
			}
		}
	}
	atomic.AddUint64(workersDone, 1)
	if atomic.LoadUint64(workersDone) == workerCount {
		close(results)
	}
}

func TilesetListWorker2(jobs <-chan string, results chan<- string, workersDone *uint64, workerCount uint64) {
	// defer wg.Done()
	for job := range jobs {
		yList, err := ioutil.ReadDir(job)
		if err != nil {
			log.Error().Err(err).Msgf("Could not read z dir: %v", job)
		} else {
			// var tileList []string
			for _, tile := range yList {
				if !tile.IsDir() {
					fname := tile.Name()
					if filepath.Base(fname) != ".DS_Store" {
						fullFilePath := filepath.Join(job, fname)
						// println("Adding file to results", fullFilePath)
						results <- fullFilePath
					}

				}
			}
		}
	}
	atomic.AddUint64(workersDone, 1)
	if atomic.LoadUint64(workersDone) == workerCount {
		close(results)
	}
}

const fstopoArc = "https://apps.fs.usda.gov/arcx/rest/services/EDW/EDW_FSTopo_01/MapServer/tile"

func VectorMergeWorker(jobs <-chan string, results chan<- string) {
	for job := range jobs {
		var err error = nil
		// print(job)
		// curDir := filepath.Dir(job)
		curTile, basedir := PathToTile(job)

		y := StripExt(filepath.Base(job))

		// fdir := filepath.Dir(job)
		// x := filepath.Base(fdir)
		// fdir = filepath.Dir(fdir)
		// z := filepath.Base(fdir)
		// fdir = filepath.Dir(fdir)

		// zxy := z + "/" + x + "/" + y
		zx := curTile.getPathZX()

		// baseFolder := fdir + "/" + zx
		topoFolder := basedir + "-topo/" + zx
		outFolder := basedir + "-merged/" + zx
		baseImg := job
		topoImg := topoFolder + "/" + y + "-topo.png"
		outImg := outFolder + "/" + y + ".png"
		// fmt.Printf("(x, y, z) - (%v, %v, %v)\n", x, y, z)

		// if fileExists(baseImg) {
		// 	os.Remove(job)
		// 	os.Rename(baseImg, job)
		// }

		vecturl := fmt.Sprintf("%v/%v/%v/%v", fstopoArc, curTile.z, curTile.y, curTile.x)
		// println(vecturl)

		topoDownloaded := true
		if !fileExists(topoImg) {
			dlTopoFolder := topoFolder + "/" + y + "-temp"
			// print("DLLoc:", dlLoc)
			_, err = DownloadFile(dlTopoFolder, vecturl)
			if err != nil {
				log.Error().Msgf("Failed to download file - %v", err.Error())
				os.Remove(dlTopoFolder)
				topoDownloaded = false
				os.Link(baseImg, outImg)
			} else {
				dlTopo := dlTopoFolder + "/" + strconv.Itoa(curTile.x)
				os.Rename(dlTopo, topoImg)
				os.Remove(dlTopoFolder)
			}
		} else {
			log.Info().Msg("Topo tile already downloaded, using cached version")
		}
		if topoDownloaded {

			err = os.MkdirAll(filepath.Dir(outImg), 0755)
			if err != nil {
				log.Error().Msgf("Failed to create output dir - %v", err.Error())
			} else {
				err = CombineImages(baseImg, topoImg, outImg)
			}

		}
		if err != nil {
			logMsg(results, job, err.Error())
		} else {
			logMsg(results, outImg, "Job done")
		}

	}
}

func logMsg(results chan<- string, source, msg string) {
	toSend := source + ": " + msg
	results <- toSend
}
