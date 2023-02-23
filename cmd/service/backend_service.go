package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"glacierpeak.app/openland/pkg/proc_mgmt"
	"glacierpeak.app/openland/pkg/proc_runners"
)

func main() {
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// workersOpt := flag.Int("t", 1, "The number of concurrent jobs being processed")
	logFileLoc := flag.String("logfile", "./log.txt", "The root directory of the source files")
	logToFile := flag.Bool("f", false, "Whether or not to log to file")
	verboseOpt := flag.Int("v", 1, "Set the verbosity level:\n"+
		" 0 - Only prints error messages\n"+
		" 1 - Adds run specs and error details\n"+
		" 2 - Adds general progress info\n"+
		" 3 - Adds debug info and details more detail\n")
	port := flag.String("p", "8080", "The port to listen on")
	flag.Parse()

	if *logToFile {
		f, err := os.OpenFile(*logFileLoc, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Error().Msgf("error opening file: %v", err)
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		} else {
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: f})
		}
		defer f.Close()
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	switch *verboseOpt {
	case 0:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		// log.SetLevel(log.ErrorLevel)
	case 1:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
		// log.SetLevel(log.WarnLevel)
	case 2:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case 3:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		// log.SetReportCaller(true)
	default:
		break
	}

	log.Info().Msg("Starting backend service")

	// currentProcesses := make([]Process, 0)

	// port := os.Getenv("OLPORT")

	processManager := proc_mgmt.NewProcessManager()

	// Starts a new Gin instance with no middle-ware
	r := gin.New()
	setupGin(r, processManager)

	go processManager.Start()
	log.Printf("Started process manager")

	mockTaskChain := proc_mgmt.NewOpenlandTaskChain("mockTaskChain")
	toTiffExecutor := proc_runners.NewPDF2TIFFExecutor("./inDir", "./outDir", []string{"filt1", "filt2"}, "700", 4)
	tilOverviewExecutor := proc_runners.NewMassTileMergeExecutor("./inDir", "./outDir", "17", 4)
	mockTask1 := proc_mgmt.NewOpenlandTask("mockTask1", toTiffExecutor)
	mockTask2 := proc_mgmt.NewOpenlandTask("mockTask2", tilOverviewExecutor)

	mockTaskChain.AddTask(mockTask1)
	mockTaskChain.AddTask(mockTask2)
	processManager.Pause()
	processManager.QueueTaskChain(mockTaskChain)
	// processManager.AddTaskChain(mockTaskChain)

	// Listen and serve on defined port
	log.Printf("Listening on port %s", *port)
	r.Run(":" + *port)

}

type HandlerMap struct {
	name     string
	handlers []*proc_mgmt.ProcessExecutable
}

func setupGin(r *gin.Engine, processManager *proc_mgmt.ProcessManager) {
	// Define handlers
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!")
	})
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.POST("/pause", func(c *gin.Context) {
		processManager.Pause()
		c.String(http.StatusOK, "Paused")
	})
	r.POST("/resume", func(c *gin.Context) {
		processManager.Resume()
		c.String(http.StatusOK, "Resumed")
	})
	r.GET("/processQueue", func(c *gin.Context) {
		queue := processManager.GetProcessQueue()
		log.Info().Msgf("Process queue: %v", *queue)
		c.JSON(http.StatusOK, queue)
	})
	r.GET("/queueHandlers", func(c *gin.Context) {
		queue := processManager.GetProcessQueue()
		tasks := make([]*HandlerMap, 0)

		for _, taskChain := range queue.Items() {
			handlers := make([]*proc_mgmt.ProcessExecutable, 0)
			for _, task := range taskChain.Tasks {
				handlerVal := (*task.Handler).Value()
				log.Printf("Handler value: %v", *handlerVal)
				handlers = append(handlers, handlerVal)
			}
			toAdd := &HandlerMap{taskChain.Name, handlers}
			log.Printf("Adding: %v", toAdd)
			tasks = append(tasks, toAdd)
		}
		log.Info().Msgf("Process queue: %v", len(tasks))

		c.JSON(http.StatusOK, tasks)
	})
	r.GET("/process:id", func(c *gin.Context) {
		id := c.Param("id")
		idInt, _ := strconv.Atoi(id)
		process, err := processManager.GetProcess(idInt)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, process)
	})

}
