package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func runAndWriteCommand(outName, cmdName string, args ...string) error {
	// for i := 0; i < len(args); i++ {
	// 	args[i] = strings.Replace(args[i], " ", "\\ ", -1)
	// }
	// fmt.Println("Running cmd: " + cmdName + " " + strings.Join(args, " ") + " > " + outName)
	cmd := exec.Command(cmdName, args...)

	outfile, err := os.Create(outName)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile

	err = cmd.Start()
	if err != nil {
		fmt.Println("Error running cmd: " + cmdName + " " + strings.Join(args, " ") + " > " + outName)
		return err
	}
	cmd.Wait()
	return nil
}

func runCommand(silent bool, cmd string, args ...string) (string, error) {
	// for i := 0; i < len(args); i++ {
	// 	args[i] = strings.Replace(args[i], " ", "\\ ", -1)
	// }
	// fmt.Println("Running cmd: " + cmd + " " + strings.Join(args, " "))
	out, err := exec.Command(cmd, args...).Output()
	if !silent {
		fmt.Printf("%s\n", out)
	}
	if err != nil {
		fmt.Println("Command unsuccessful: " + cmd + " " + strings.Join(args, " "))
		return "", err
	}
	return string(out), nil
}
