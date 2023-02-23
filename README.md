# Open Land Data

These are a collection of scripts and libraries that aid in processing geospatial data.

All runnable go scripts are located at `cmd/<module>/<script>.go`. They can be run by moving your current directory to one of these script modules and running the go script. Ex:

```shell
cd cmd/bulk2tiles
go run bulk2tiles.go -h
```

Miscelanious scripts used to pre-process data are in `misc-scripts/`.
All library code is in `pkg/`

---

## Toolbox

Ensure these tools are installed and in your PATH. Some scripts depend on these.

- [tippecanoe](https://github.com/mapbox/tippecanoe)
- ogr2ogr - from [gdal](https://gdal.org/download.html)
- [geojson-polygon-labels](https://github.com/andrewharvey/geojson-polygon-labels)

## Usage

All go scripts can be passed `-h` to see usage information. Each script can be used on it's own, but they are built to be used in conjunction. Ex: For converting USFS topo PDF's to
a merged tileset, this is the order to run the commands on the dataset:

```sh
pdf2tiff -> bulk2tiles -> cleantiles -> fixbackground -> tilemerge -> tileoverview
```

Most scripts create a copy and then modify the data, so each step is reversible. (Note: One exception to this is the `cleantiles` script, which modifies in place)

## Dependencies & Compilation

Ensure all modules are verified by running `go mod verify`

You can compile a given script with `go build <script-name>.go`. This will create an executable called `<script-name>`.
Alternatively, you can run `go run <script-name>.go` which will compile and run in the same step, but will not leave an executable.

If there are any missing dependencies, the compiler will throw an error `cannot find package "<package-name>"` Use `go get <package-name>` to install them.
