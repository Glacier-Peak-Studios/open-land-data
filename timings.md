## Timings
(10 random pdfs)
go run pdf2tiff.go -i /Users/solidsilver/Code/USFS/timings/pdfs -o  -t 10  3667.25s user 118.80s system 725% cpu 8:41.79 total
Each tiff is 1.01 GB
go run pdf2tiff.go -i /Users/solidsilver/Code/USFS/timings/pdfs -o  -dpi 500   2691.64s user 94.30s system 708% cpu 6:33.08 total
Each tiff is 448.1 MB

gdal2tiles.py --zoom="17" -e --processes=16 --xyz   144.13s user 8.88s system 1340% cpu 11.417 total


## Relative Changes

500/750 ~= 0.667 (DPI Ratio)
6.55/8.69 ~= 0.754 (Time Ratio)

448.1 MB / 1.01 GB ~= 0.433 (Size Ratio)


## Size & Time Bounds


### PDF to TIFF

#### 500 DPI
@448.1 MB -> 4.814 TB

Time 4.16 Days

##### 750 DPI
@1.01 GB -> 10.85 TB

Time 5.53 Days


### TIFF to Tiles

zoom@17 -> 37.42 Hrs

zoom@18 -> 5.524 Days


750 DPI + Zoom 17