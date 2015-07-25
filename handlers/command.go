package handlers

import (
	"github.com/bulletind/moire/logger"
	"os/exec"
	"strings"
)

var log = logger.Logger

func printCommand(cmd *exec.Cmd) {
	log.Debug("Command", "Executing", strings.Join(cmd.Args, " "))
}

func printError(err error) {
	if err != nil {
		log.Debug("Command", "Error", err.Error())
	}
}

func printOutput(outs []byte) {
	if len(outs) > 0 {
		log.Debug("Command", "Output", outs)
	}
}

func executeRaw(cmd *exec.Cmd) error {
	printCommand(cmd)
	output, err := cmd.CombinedOutput()
	printError(err)
	printOutput(output)

	return err
}
