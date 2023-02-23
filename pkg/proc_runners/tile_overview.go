package proc_runners

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
	"glacierpeak.app/openland/pkg/proc_mgmt"
	"glacierpeak.app/openland/pkg/utils"
)

func CreateOverviewRange(zMax int, zMin int, dir string, workers int) {
	for i := zMax; i > zMin; i-- {
		CreateOverview(filepath.Join(dir, strconv.Itoa(i)), workers)
	}
}

func CreateOverview(dir string, workers int) {
	log.Warn().Msgf("Searching sources dir: %v", dir)
	sources := utils.GetAllTiles2(dir, workers)
	// sources, _ := utils.WalkMatch(dir, "*")
	m := make(map[string]bool)
	var overviews []string

	for _, source := range sources {
		over := utils.OverviewRoot(source)
		tile, _ := utils.PathToTile(over)
		if !m[tile.GetPathXY()] {
			overviews = append(overviews, over)
			m[tile.GetPathXY()] = true
		}

		// tileList = utils.AppendSetT(tileList, tile)
	}

	progBar := progressbar.NewOptions(len(overviews),
		progressbar.OptionSetDescription(fmt.Sprintf("Generating overview for level %s...", dir)),
		progressbar.OptionSetItsString("tiles"),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
	// sources = utils.SetMap(sources, utils.OverviewRoot)

	jobCount := len(overviews)
	jobs := make(chan string, jobCount)
	results := make(chan string, jobCount)

	log.Warn().Msgf("Running with %v workers", workers)
	for i := 0; i < workers; i++ {
		go utils.OverviewWorker(jobs, results)
	}
	go queueSources(overviews, jobs)

	for i := 0; i < jobCount; i++ {
		var rst = <-results
		log.Debug().Msg(rst)
		progBar.Add(1)
	}
	close(jobs)
	progBar.Finish()
	log.Warn().Msg("Done with all jobs")
}

type CreateOverviewRangeExecutor struct {
	ZMax    int
	ZMin    int
	Dir     string
	Workers int
}

func NewCreateOverviewRangeExecutor(zMax int, zMin int, dir string, workers int) *CreateOverviewRangeExecutor {
	return &CreateOverviewRangeExecutor{zMax, zMin, dir, workers}
}

func (e *CreateOverviewRangeExecutor) Value() *proc_mgmt.ProcessExecutable {
	return &proc_mgmt.ProcessExecutable{
		Name: "CreateOverviewRange",
		Args: []string{
			strconv.Itoa(e.ZMax),
			strconv.Itoa(e.ZMin),
			e.Dir,
			strconv.Itoa(e.Workers),
		},
		Run: e.Run,
	}
}

func (e *CreateOverviewRangeExecutor) Run() {
	CreateOverviewRange(e.ZMax, e.ZMin, e.Dir, e.Workers)
}
