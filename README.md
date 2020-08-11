# Open Land Data

## Toolbox

- [tippecanoe](https://github.com/mapbox/tippecanoe)
- ogr2ogr - from [gdal](https://gdal.org/download.html)
- [geojson-polygon-labels](https://github.com/andrewharvey/geojson-polygon-labels)

Note: These will need to be installed and in your PATH

## Dependencies & Compilation

Download all script dependencies by running `go get -d ./...`

Compile with `go build main.go`. This will create an executable called `main`.
Alternatively, you can run `go run main.go` which will compile and run in the same step, but will not leave an executable.

If there are any missing dependencies, the compiler will throw an error `cannot find package "<package-name>"` Use `go get <package-name>` to install them.

## Usage

View script usage by running `./main -h`

To generate all sources, run `./main`

## Sources

This script reads sources from a directory structure of .json files. On generation, it copies the sources directory structure, and each .json becomes a useable generated file. By default, no sources are included in this project. To obtain sources to convert:

- For the last stable set of sources, visit the [json sources repo](https://github.com/Solidsilver/land-sources). This can be dropped into this project at the root level.

- For the very latest sources (still being audited), visit the [sources doc](https://docs.google.com/spreadsheets/d/1fMXYRg1WATDALrtcbyD0U4BKYJWt7hV_ikix1dROlCs/edit?usp=sharing), download as a .csv, and convert to json.

To convert sources from .csv to json, and vise versa, run `./sourcemgr.py`
