package main 

import (
	"os"
	"time"
	"strconv"
	"runtime"
	"os/exec"
)

func daemon(programName string) {
	for {
		cmd := exec.Command(programName)
		err := cmd.Start()
		if err != nil {
			time.Sleep(time.Second * 5)
			continue
		}
		pid := cmd.Process.Pid
		pid_file, _ := os.OpenFile("../etc/" + programName + ".pid", os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0666) 
		pid_file.WriteString(strconv.Itoa(pid))
		defer pid_file.Close()
		err = cmd.Wait()
		time.Sleep(time.Second * 100)
	}
}

func main() {
	done := make(chan bool, 1)
	mainProgram := "./EccReportAgent"
	updateProgram := "./AgentUpdate"
	if runtime.GOOS == "windows" {
		mainProgram += ".exe"
		updateProgram += ".exe"
	}
	go daemon(mainProgram)
	go daemon(updateProgram)
	<- done
}
