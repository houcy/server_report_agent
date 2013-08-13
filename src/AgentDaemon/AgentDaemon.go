package main 

import (
	"os"
	"fmt"
	"log"
	"time"
	"utils"
	"runtime"
	"os/exec"
	"strconv"
	"strings"
	"net/url"
)

var settings utils.Settings
var logger *log.Logger
var pid int

/*
	Check the update server to get update information.
	If the content is not empty, then split the file names splitted by a semicolon.
	File names are like these: mod/cpu_usage.vbs or bin/EccReportAgent
*/
func checkList() []string {
	var list []string
	urlString := settings.UpdateServer[0]["url"] + "/get_list"
	proxyString := settings.UpdateServer[0]["proxy"]
	resp, err := utils.ReadRemote(urlString, proxyString)
	if err != nil {
		logger.Println(err.Error())
	    return list
	}
	list = strings.Split(string(resp), ";")
	return list
}

/*
	If the file name prepend with "../" exists, rename it by appending an
	timestamp for now. And then download the file to the corresponding place.
*/
func downloadAndReplaceFile(filename string) bool {
	theFile := "../" + filename
	if _, err := os.Stat(theFile); err == nil {
		// If file exists.
	    err = os.Rename(theFile, theFile + "." + strconv.FormatInt(time.Now().Unix(), 10))
		if err != nil {
		    logger.Println(err.Error())
		    return false
		}
	}
	f, err := os.OpenFile(theFile, os.O_CREATE|os.O_WRONLY, 0666)
	defer f.Close()
    if err != nil {
	    logger.Println(err.Error())
	    return false
	}
	urlString := settings.UpdateServer[0]["url"] + "/get_file/?name=" + url.QueryEscape(filename)
	proxyString := settings.UpdateServer[0]["proxy"]
    resp, err := utils.ReadRemote(urlString, proxyString)
	if err != nil {
	    logger.Println(err.Error())
	    return false
	}
	f.Write(resp)
	return true
}

func stopAndUpdate() {
	c := time.Tick(300 * time.Second)
	for _ = range c {
		if files :=checkList(); len(files) != 0 {
			kp, err := os.FindProcess(pid)
			if err != nil {
			    logger.Println(err.Error())
			    return
			}
			err = kp.Kill()
			if err != nil {
			    logger.Println(err.Error())
			    return
			}
			for _, filename := range files {
				if !downloadAndReplaceFile(filename) {
					logger.Println("update failed for:", filename)
				}
				time.Sleep(time.Second * 3)
			}
		}
	}
}

func main() {
	logfile,err := os.OpenFile("../log/daemon.log", os.O_CREATE | os.O_RDWR | os.O_APPEND, 0666) 
	if err!=nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	logger := log.New(logfile,"",log.Ldate|log.Ltime|log.Lshortfile)
	defer logfile.Close()
	settings, err = utils.LoadSettings()
	if err != nil {
		logger.Fatalln(err)
		os.Exit(1)
	}
	// Update goroutine
	go stopAndUpdate()
	// Daemon deadloop
	mainProgram := "./EccReportAgent"
	if runtime.GOOS == "windows" {
		mainProgram += ".exe"
	}
	for {
		cmd := exec.Command(mainProgram)
		err := cmd.Start()
		if err != nil {
			logger.Println(err.Error())
			time.Sleep(time.Second * 5)
			continue
		}
		pid = cmd.Process.Pid
		logger.Println("Agent started by the daemon. PID:", pid)
		err = cmd.Wait()
		logger.Println("Agent program exit.", err.Error())
		time.Sleep(time.Second * 100)
	}
}
