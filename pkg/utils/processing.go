package utils

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// ProcessSource formats source data specified by a .json
func ProcessSource(dir string, file string) error {
	// if file == "public-hunting-lands.json" {
	// 	fmt.Println("Time to debug!")
	// }
	pathedFile := dir + "/" + file
	sourceJSON := FileToStr(pathedFile)
	sourceURL := getPropFromJSON("url", sourceJSON)
	dlPath := strings.Replace(dir, "land-sources", "generated", 1)
	dlurl, _ := url.Parse(sourceURL)
	dlFile := dlPath + "/" + filepath.Base(dlurl.Path)
	var err error
	if !fileExists(dlFile) {
		dlFile, err = DownloadFile(dlPath, sourceURL)
		if err != nil {
			return err
		}
	} else {
		log.Debug().Msg("DL exists: " + dlFile)
	}
	if filepath.Ext(dlFile) == ".zip" {
		dlPath = dlPath + "/" + getFnameOnly(dlFile)
		if !DirExists(dlPath) {
			_, err := RunCommand(true, "unzip", "-j", dlFile, "-d", dlPath)
			if err != nil {
				return err
			}
		}
	}
	fname := getPropFromJSON("filenameInZip", sourceJSON)
	if fname != "" && !fileExists(dlPath+"/"+fname) {
		return errors.New("filenameInZip not found in archive: " + fname)
	}
	switch filetype := getPropFromJSON("filetype", sourceJSON); filetype {
	case "geojson":
		return processGeoJSON(dlPath, getFnameOnly(dlFile), getFnameOnly(file))
	case "shp":
		return processShp(dlPath, fname, getFnameOnly(file))
	case "kml":
		return processKml(dlPath, fname, getFnameOnly(file))
	case "kmz":
		return processKmz(dlPath, fname, getFnameOnly(file))
	default:
		return errors.New("Filetype not supported: " + filetype)
	}
}

func processGeoJSON(path, filename string, fnameOut ...string) error {
	fileWithPath := path + "/" + getFnameOnly(filename)
	geojson := fileWithPath + ".geojson"
	if !fileExists(geojson) {
		return errors.New("Cannot process geojson! file doesn't exist: " + geojson)
	}
	if len(fnameOut) > 0 {
		fileWithPath = path + "/" + getFnameOnly(fnameOut[0])
		os.Rename(geojson, fileWithPath+".geojson")
		geojson = fileWithPath + ".geojson"
	}

	geojsonLabels := fileWithPath + "-labels.geojson"
	mbtiles := fileWithPath + ".mbtiles"
	mbtilesLabels := fileWithPath + "-labels.mbtiles"
	log.Debug().Msg("Processing geoJson: " + geojson)
	var err error

	err = generateLabels(geojsonLabels, geojson)
	if err != nil {
		return err
	}
	err = generateMBTiles(mbtiles, geojson)
	if err != nil {
		return err
	}
	err = generateMBTiles(mbtilesLabels, geojsonLabels)
	if err != nil {
		return err
	}
	const combine = false
	if combine {
		combined := fileWithPath + "-combined.mbtiles"
		err = combineMBTiles(combined, mbtiles, mbtilesLabels)
		if err != nil {
			return err
		}
		if fileExists(geojsonLabels) {
			err = os.Remove(geojsonLabels)
		}
		if fileExists(mbtiles) {
			err = os.Remove(mbtiles)
		}
		if fileExists(mbtilesLabels) {
			err = os.Remove(mbtilesLabels)
		}
	}
	return err
}

func processShp(path, filename, fileOutName string) error {
	if filename == "" {
		shapefiles, err := WalkMatch(path, "*.shp")
		if err != nil {
			return err
		}
		if len(shapefiles) > 1 {
			log.Debug().Msgf("shapefiles-in-dir: %v", shapefiles)
			return errors.New("multiple shapefiles in zip, none specified in source")
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
	var err error
	log.Debug().Msg("Processing shapefile: " + shapefile)
	if !fileExists(geojson) {
		if !fileExists(shapefile) {
			return errors.New("Cannot convert shp to geojson - shp doesn't exist: " + shapefile)
		}
		_, err := RunCommand(false, "ogr2ogr", "-f", "GeoJSON", "-t_srs", "crs:84", geojson, shapefile)
		if err != nil {
			return err
		}
	}
	err = processGeoJSON(basepath, fileOutName)

	return err
}

func processKml(path, filename, fileOutName string) error {
	if filename == "" {
		kmlfiles, err := WalkMatch(path, "*.kml")
		if err != nil {
			return err
		}
		if len(kmlfiles) > 1 {
			log.Debug().Msgf("kmlfiles-in-dir: %v", kmlfiles)
			return errors.New("multiple kmlfiles in zip, none specified in source")
		}
		if len(kmlfiles) == 0 {
			return errors.New("No kmlfiles in folder: " + path)
		}
		filename = filepath.Base(kmlfiles[0])
	}
	basepath := filepath.Dir(path)
	filename = getFnameOnly(filename)
	fileOutName = getFnameOnly(fileOutName)
	fileWithPath := path + "/" + filename
	geojson := basepath + "/" + fileOutName + ".geojson"
	kmlfile := fileWithPath + ".kml"
	var err error
	log.Debug().Msg("Processing kmlfile: " + kmlfile)
	if !fileExists(geojson) {
		if !fileExists(kmlfile) {
			return errors.New("Cannot convert kml to geojson - kml doesn't exist: " + kmlfile)
		}
		_, err := RunCommand(false, "ogr2ogr", "-f", "GeoJSON", "-t_srs", "crs:84", geojson, kmlfile)
		if err != nil {
			return err
		}
	}
	err = processGeoJSON(basepath, fileOutName)

	return err
}

func processKmz(path, filename, fileOutName string) error {
	if filename == "" {
		kmzfiles, err := WalkMatch(path, "*.kmz")
		if err != nil {
			return err
		}
		if len(kmzfiles) > 1 {
			log.Debug().Msgf("kmzfiles-in-dir: %v", kmzfiles)
			return errors.New("multiple kmzfiles in zip, none specified in source")
		}
		if len(kmzfiles) == 0 {
			return errors.New("No kmzfiles in folder: " + path)
		}
		filename = filepath.Base(kmzfiles[0])
	}
	folderPath := path + "/" + getFnameOnly(filename)
	if !DirExists(folderPath) {
		_, err := RunCommand(true, "unzip", "-j", path+"/"+filename, "-d", folderPath)
		if err != nil {
			return err
		}
	}

	return processKml(folderPath, "", fileOutName)
}

func generateLabels(newfile, geojson string) error {
	if !fileExists(newfile) {
		if !fileExists(geojson) {
			return errors.New("Cannot create label! geojson doesn't exist: " + geojson)
		}
		return runAndWriteCommand(newfile, "geojson-polygon-labels", "--label=polylabel", "--include-minzoom=1-11", geojson)
	}
	log.Info().Msg(newfile + " already exists, skipping. Use -f to regenerate")
	return nil
}

func generateMBTiles(newfile, geojson string) error {
	if !fileExists(newfile) {
		if !fileExists(geojson) {
			return errors.New("Cannot create mbtile! geojson doesn't exist: " + geojson)
		}
		_, err := RunCommand(false, "tippecanoe", "-f", "-z11", "-o", newfile, geojson)
		return err
	}
	log.Info().Msg(newfile + " already exists, skipping. Use -f to regenerate")
	return nil
}

func combineMBTiles(newfile, mbtiles, mbtilesLabels string) error {
	if !fileExists(newfile) {
		if !fileExists(mbtiles) {
			return errors.New("Cannot join mbtiles! base mbtile doesn't exist: " + mbtiles)
		}
		if !fileExists(mbtilesLabels) {
			return errors.New("Cannot join mbtiles! labels mbtile doesn't exist: " + mbtilesLabels)
		}
		_, err := RunCommand(false, "tile-join", "-f", "-o", newfile, mbtiles, mbtilesLabels)
		return err
	}
	log.Info().Msg(newfile + " already exists, skipping. Use -f to regenerate")
	return nil
}
