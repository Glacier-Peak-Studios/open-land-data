package main

import (
	"errors"
	"flag"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func main() {
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// workersOpt := flag.Int("t", 1, "The number of concurrent jobs being processed")
	logFileLoc := flag.String("src1", "./log.txt", "The root directory of the source files")
	verboseOpt := flag.Int("v", 1, "Set the verbosity level:\n"+
		" 0 - Only prints error messages\n"+
		" 1 - Adds run specs and error details\n"+
		" 2 - Adds general progress info\n"+
		" 3 - Adds debug info and details more detail\n")
	port := flag.String("p", "8080", "The port to listen on")
	flag.Parse()

	f, err := os.OpenFile(*logFileLoc, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Error().Msgf("error opening file: %v", err)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: f})
	}
	defer f.Close()

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

	currentProcesses := make([]Process, 0)

	// port := os.Getenv("OLPORT")

	// Starts a new Gin instance with no middle-ware
	r := gin.New()
	setupGin(r, &currentProcesses)

	// Listen and serve on defined port
	log.Printf("Listening on port %s", port)
	r.Run(":" + *port)

}

func setupGin(r *gin.Engine, currentProcesses *[]Process) {
	// Define handlers
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!")
	})
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.POST("/newProcess", func(c *gin.Context) {
		processName := c.Params.ByName("name")
		var process Process
		c.BindJSON(&process)
		process.ID = len(*currentProcesses)
		process.Status = "running"
		process.Name = processName
		*currentProcesses = append(*currentProcesses, process)

		c.JSON(http.StatusCreated, process)
	})

	r.GET("/process/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		idInt, _ := strconv.Atoi(id)
		process, err := getProcess(idInt, currentProcesses)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "process found",
			"process": process,
		})
	})

	r.POST("/end-process/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		idInt, _ := strconv.Atoi(id)
		err := removeProcess(idInt, currentProcesses)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "process ended",
		})
	})
}

func removeProcess(id int, currentProcesses *[]Process) error {
	for i, process := range *currentProcesses {
		if process.ID == id {
			*currentProcesses = append((*currentProcesses)[:i], (*currentProcesses)[i+1:]...)
			return nil
		}
	}
	return errors.New("process not found")
}

func getProcess(id int, currentProcesses *[]Process) (Process, error) {
	for _, process := range *currentProcesses {
		if process.ID == id {
			return process, nil
		}
	}
	return Process{}, errors.New("process not found")
}

type Process struct {
	ID     int
	Name   string
	Status string
}
