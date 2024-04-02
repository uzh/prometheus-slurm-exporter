package main

import (
	"log"
	"os/exec"
	"io"
)

// Tests whether the utility is available from the command line (i.e. in the PATH)
func UtilityAvailable(utility string) bool {
	cmd := exec.Command(utility, "--version")
	err := cmd.Run()
	return err == nil
}


// Execute the sinfo command and return its output
func Execute(command string, arguments []string) []byte {
	cmd := exec.Command(command, arguments...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	out, _ := io.ReadAll(stdout)
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	return out
}

