package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

func PDF2TiffWorker(jobs <-chan string, results chan<- string, cmd string, constArgs ...string) {
	for job := range jobs {
		// jobSplit := strings.Split(job, " ")
		println("-> Job -", job)
		pdfLayers := GetGeoPDFLayers(job)

		pdfLayers = Filter(pdfLayers, LayerFilter)
		keepLayers := "\"" + strings.Join(pdfLayers[:], ",") + "\""
		args := append(constArgs, "--config", "GDAL_PDF_LAYERS", keepLayers)

		fout := filepath.Join(filepath.Dir(job), StripExt(job)+".tif")
		args = append(args, job, fout)
		// constArgs = append(constArgs, job, fout)

		if !fileExists(fout) {
			// println("Going to run::")
			// argList := strings.Join(args, " ")
			// println(cmd, argList)
			out, err := RunCommand2(false, true, cmd, args...)
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

	return true
}

func OverviewWorker(jobs <-chan string, results chan<- string) {
	for job := range jobs {
		var err error = nil

		curTile, basepath := pathToTile(job)

		img1 := filepath.Join(basepath, curTile.getPath()+".png")
		tRight := curTile.rightTile()
		img2 := filepath.Join(basepath, tRight.getPath()+".png")
		tDown := curTile.downTile()
		img3 := filepath.Join(basepath, tDown.getPath()+".png")
		tDiag := tRight.downTile()
		img4 := filepath.Join(basepath, tDiag.getPath()+".png")

		tOver := curTile.overviewTile()
		imgOut := filepath.Join(basepath, tOver.getPath()+".png")

		err = GenerateOverviewTile(imgOut, img1, img2, img3, img4)

		if err != nil {
			logMsg(results, job, err.Error())
		} else {
			logMsg(results, imgOut, "Job done")
		}

	}
}

func TilesetMergeWorker(jobs <-chan string, results chan<- string, ts1Dir string, ts2Dir string) {
	for job := range jobs {
		curTile, _ := pathToTile(job)
		outImg := job
		ts1File := ts1Dir + "/" + curTile.getPath() + ".png"
		ts2File := ts2Dir + "/" + curTile.getPath() + ".png"
		os.MkdirAll(filepath.Dir(outImg), 0755)
		// if intersect.PointInBBox(curTile.xyPoint()) {
		// 	err := MergeTiles(ts1File, ts2File, outImg)
		// 	if err != nil {
		// 		log.Error("Could not merge images: ", err.Error())
		// 	}
		// } else {
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
			// os.Link(tileCopy, outImg)
			os.Rename(tileCopy, outImg)

			logMsg(results, job, "Copying tile")
		}
	}
	// }
}

func TileTrimWorker(jobs <-chan string, results chan<- string) {
	for job := range jobs {
		println(job)
		basepath := job
		workingPath := filepath.Join(basepath, "18")
		bbx := BBoxFromTileset(workingPath)

		sides := [4]string{"left", "right", "up", "bottom"}

		for _, side := range sides {
			checkLine := bbx.getSideLine(side)
			toRemove := BBx(bbx.Origin(), bbx.Origin())
			if side == "right" || side == "bottom" {
				toRemove = BBx(bbx.Extent(), bbx.Extent())
			}
			isWhite := checkLine.isBBoxWhite(basepath, 18)
			for ; isWhite; isWhite = checkLine.isBBoxWhite(basepath, 18) {
				toRemove, _ = GetBBoxMerge(toRemove, checkLine)
				bbx.ChangeSide(side, -1)
				checkLine = bbx.getSideLine(side)
			}
			removeTilesInBBox(toRemove, basepath, 18)
			// CleanBBoxEdge(checkLine, side, basepath, 18)

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

const fstopoArc = "https://apps.fs.usda.gov/arcx/rest/services/EDW/EDW_FSTopo_01/MapServer/tile"

func VectorMergeWorker(jobs <-chan string, results chan<- string) {
	for job := range jobs {
		var err error = nil
		// print(job)
		// curDir := filepath.Dir(job)
		curTile, basedir := pathToTile(job)

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