package runners

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
	"glacierpeak.app/openland/pkg/utils"
)

func FixBackground(dir string, out string, workers int, zLvl string) {
	m := make(map[string]bool)
	sources, _ := utils.WalkMatch(dir, "*.png")
	// get directory listing.
	// files, _ :=
	// sources := utils.GetAllTiles2(filepath.Join(dir, zLvl), workers)
	var tileList []utils.Tile

	for _, source := range sources {
		tile, _ := utils.PathToTile(source)
		m[tile.GetPathXY()] = true
		tileList = append(tileList, tile)
		// tileList = utils.AppendSetT(tileList, tile)
	}

	progBar := progressbar.NewOptions(len(tileList),
		progressbar.OptionSetDescription("Fixing tile backgrounds..."),
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

	jobCount := len(tileList)
	jobs := make(chan utils.Tile, jobCount)
	results := make(chan string, jobCount)

	log.Warn().Msgf("Running with %v workers", workers)
	for i := 0; i < workers; i++ {
		go utils.FixBackgroundWorker(jobs, results, m, dir, out)
		// go utils.TilesetMergeWorker2(jobs, results, m, out)
	}
	for _, tile := range tileList {
		jobs <- tile
	}
	for i := 0; i < jobCount; i++ {
		var rst = <-results
		progBar.Add(1)
		log.Debug().Msg(rst)
	}
	close(jobs)
	progBar.Finish()
	log.Warn().Msg("Done with all jobs")
}

type FixBackgroundExecutor struct {
	Dir  string
	Out  string
	ZLvl string
}

func NewFixBackgroundExecutor(dir string, out string, zLvl string) *FixBackgroundExecutor {
	return &FixBackgroundExecutor{dir, out, zLvl}
}

func (fbe *FixBackgroundExecutor) Value() string {
	return fmt.Sprintf("FixBackground: In:%s, Out:%s, ZLvl:%s", fbe.Dir, fbe.Out, fbe.ZLvl)
}

func (fbe *FixBackgroundExecutor) Args() []string {
	return []string{"dir", "out", "zlvl"}
}

func (fbe *FixBackgroundExecutor) Run() {
	FixBackground(fbe.Dir, fbe.Out, 1, fbe.ZLvl)
}
