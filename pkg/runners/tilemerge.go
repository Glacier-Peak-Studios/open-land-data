package runners

import (
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
	"glacierpeak.app/openland/pkg/utils"
)

func MassTileMerge(setsDir string, out string, zLevel string, workers int) {

	m, tileList := utils.GetAllTiles0(setsDir, zLevel, workers)

	log.Warn().Msg("Done gathering tiles.")

	rsltLen := len(tileList)
	jobCount := 32
	jobs := make(chan utils.Tile, jobCount)
	results := make(chan string, jobCount)
	readChan := make(chan int, 1)

	log.Warn().Msgf("Running with %v workers", workers)
	go resultReaderWorker(results, jobs, rsltLen, readChan)

	mapLock := sync.RWMutex{}

	for i := 0; i < workers; i++ {
		go utils.TilesetMergeWorker0(jobs, results, m, out, setsDir, &mapLock)
	}
	for _, tile := range tileList {
		jobs <- tile
	}

	<-readChan

	log.Warn().Msg("Done with all jobs")

}

func resultReaderWorker(toRead <-chan string, jobs chan utils.Tile, resultCount int, result chan<- int) {

	progBar := progressbar.NewOptions(resultCount,
		progressbar.OptionSetDescription("Merging tiles..."),
		progressbar.OptionSetItsString("tiles"),
		progressbar.OptionShowIts(),
		progressbar.OptionThrottle(1*time.Second),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	for i := 0; i < resultCount; i++ {
		<-toRead
		progBar.Add(1)
	}
	close(jobs)
	progBar.Finish()
	result <- 1

}

func TileMerge(src1 string, src2 string, out string, workers int) {
	src1BBox, _ := utils.BBoxFromTileset(src1 + "/18")
	src2BBox, _ := utils.BBoxFromTileset(src2 + "/18")

	intersect, _ := utils.GetBBoxIntersect(src1BBox, src2BBox)
	intersect.ExpandBy(1)

	var sources []string

	for ix := src1BBox.Origin().X; ix <= src1BBox.Extent().X; ix++ {
		for iy := src1BBox.Origin().Y; iy <= src1BBox.Extent().Y; iy++ {
			sources = append(sources, filepath.Join(out, "18", strconv.Itoa(ix), strconv.Itoa(iy)+".png"))
		}
	}

	for ix := src2BBox.Origin().X; ix <= src2BBox.Extent().X; ix++ {
		for iy := src2BBox.Origin().Y; iy <= src2BBox.Extent().Y; iy++ {
			curP := utils.Point{X: ix, Y: iy}
			if !intersect.PointInBBox(curP) {
				sources = append(sources, filepath.Join(out, "18", strconv.Itoa(ix), strconv.Itoa(iy)+".png"))
			} else {
				println("In bbox: ", ix, "/", iy)
			}
		}
	}

	jobCount := len(sources)
	jobs := make(chan string, jobCount)
	results := make(chan string, jobCount)

	log.Warn().Msgf("Running with %v workers", workers)
	for i := 0; i < workers; i++ {
		go utils.TilesetMergeWorker(jobs, results, src1, src2)
	}
	queueSources(sources, jobs)

	for i := 0; i < jobCount; i++ {
		var rst = <-results
		log.Debug().Msg(rst)
	}
	close(jobs)
	log.Warn().Msg("Done with all jobs")
}

func Filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

// type MassTileMergeExecutor struct {
// 	setsDir string
// 	out     string
// 	zLevel  string
// 	workers int
// }

// func NewMassTileMergeExecutor(setsDir string, out string, zLevel string, workers int) *MassTileMergeExecutor {
// 	return &MassTileMergeExecutor{setsDir, out, zLevel, workers}
// }

// func (m *MassTileMergeExecutor) Run() {
// 	MassTileMerge(m.setsDir, m.out, m.zLevel, m.workers)
// }

// func (m *MassTileMergeExecutor) Value() *management.ProcessExecutable {
// 	return &management.ProcessExecutable{
// 		Name: "MassTileMerge",
// 		Args: []string{m.setsDir, m.out, m.zLevel, strconv.Itoa(m.workers)},
// 		Run:  m.Run,
// 	}
// }
