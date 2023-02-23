package utils

import (
	"os"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

func runAndWriteCommand(outName, cmdName string, args ...string) error {
	cmd := exec.Command(cmdName, args...)
	outfile, err := os.Create(outName)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile

	err = cmd.Start()
	if err != nil {
		log.Warn().Msg("Error running cmd: " + cmdName + " " + strings.Join(args, " ") + " > " + outName)
		return err
	}
	cmd.Wait()
	return nil
}

type CmdOpts struct {
	Silent     bool
	CaptureOut bool
}

// RunCommandDefault runs the given command with all opts set to false
func RunCommandDefault(cmd string, args ...string) (string, error) {
	return RunCommand(CmdOpts{}, cmd, args...)
}

// RunCommand runs the given command with the given options
func RunCommand(opts CmdOpts, cmd string, args ...string) (string, error) {
	var out []byte
	var err error
	if opts.CaptureOut {
		out, err = exec.Command(cmd, args...).Output()
	} else {
		err = exec.Command(cmd, args...).Run()
	}

	if !opts.Silent {
		log.Debug().Msgf("%s\n", out)
	}
	if err != nil {
		log.Warn().Msg("Command unsuccessful: " + cmd + " " + strings.Join(args, " "))
		return "", err
	}
	return string(out), nil
}

// CommandRunner takes in options for running a command and
// returns a function to run a given command with those options
func CommandRunner(opts CmdOpts) func(cmd string, args ...string) (string, error) {
	return func(cmd string, args ...string) (string, error) {
		return RunCommand(opts, cmd, args...)
	}
}
