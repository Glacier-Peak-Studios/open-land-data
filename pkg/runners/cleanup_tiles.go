package runners

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
	"glacierpeak.app/openland/pkg/utils"
)

func CleanupTiles(inDir string, zoomLvl int, workers int) {
	files, _ := ioutil.ReadDir(inDir)
	files = utils.Filter(files, func(file os.FileInfo) bool {
		return file.IsDir()
	})
	sources := utils.Map(files, func(file os.FileInfo) string {
		return filepath.Join(inDir, file.Name())
	})

	jobCount := len(sources)
	jobs := make(chan string, jobCount)
	results := make(chan string, jobCount)

	log.Warn().Msgf("Running with %v workers", workers)
	for i := 0; i < workers; i++ {
		go utils.TileCleanupWorker(jobs, results, zoomLvl)
	}
	queueCTSources(sources, jobs)

	progBar := progressbar.NewOptions(len(sources),
		progressbar.OptionSetDescription("Cleaning tiles..."),
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

	for i := 0; i < jobCount; i++ {
		var rst = <-results
		progBar.Add(1)
		log.Debug().Msg(rst)
	}
	close(jobs)
	progBar.Finish()
	log.Warn().Msg("Done with all jobs")
}

func queueCTSources(sources []string, jobs chan<- string) {
	progBar := progressbar.NewOptions(len(sources),
		progressbar.OptionSetDescription("Preparing tiles for cleaning..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
	for _, source := range sources {
		jobs <- source
		progBar.Add(1)
	}
	progBar.Finish()
}

// type CleanTilesExecutor struct {
// 	inDir   string
// 	zoomLvl int
// 	workers int
// }

// func NewCleanTilesExecutor(inDir string, zoomLvl int, workers int) *CleanTilesExecutor {
// 	return &CleanTilesExecutor{
// 		inDir:   inDir,
// 		zoomLvl: zoomLvl,
// 		workers: workers,
// 	}
// }

// func (cte *CleanTilesExecutor) Value() interface{} {
// 	return &cte
// }

// func (cte *CleanTilesExecutor) Run() {
// 	CleanupTiles(cte.inDir, cte.zoomLvl, cte.workers)
// }
