package kit

import (
	"log"
	"os/exec"
)

func Exec(cmd string, args ...string) {
	c := exec.Command(cmd, args...)
	stdout, err := c.Output()
	if err != nil {
		log.Printf("exec err - %s \n", err.Error())
		return
	}
	if len(stdout) > 0 {
		log.Printf(string(stdout))
	}
}
