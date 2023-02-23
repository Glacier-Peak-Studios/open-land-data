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

func RunCommand(silent bool, cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).Output()
	if !silent {
		log.Debug().Msgf("%s\n", out)
	}
	if err != nil {
		log.Warn().Msg("Command unsuccessful: " + cmd + " " + strings.Join(args, " "))
		return "", err
	}
	return string(out), nil
}

func RunCommand2(silent bool, captureOut bool, cmd string, args ...string) (string, error) {
	var out []byte
	var err error
	if captureOut {
		out, err = exec.Command(cmd, args...).Output()
	} else {
		err = exec.Command(cmd, args...).Run()
	}

	if !silent {
		log.Debug().Msgf("%s\n", out)
	}
	if err != nil {
		log.Warn().Msg("Command unsuccessful: " + cmd + " " + strings.Join(args, " "))
		return "", err
	}
	return string(out), nil
}
