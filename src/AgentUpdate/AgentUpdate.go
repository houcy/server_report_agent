package main

import (
	"os"
	"fmt"
	"log"
	"time"
	"utils"
	"runtime"
	"strconv"
	"strings"
	"net/url"
	"io/ioutil"
	"os/signal"
)

var settings utils.Settings
var logger *log.Logger
var stop bool

/*
	Check the update server to get update information.
	If the content is not empty, then split the file names splitted by a semicolon.
	File names are like these: mod/cpu_usage.vbs or bin/EccReportAgent
*/
func checkList() []string {
	var list []string
	urlString := settings.UpdateServer[0]["url"] + "/?action=get_list"
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
	urlString := settings.UpdateServer[0]["url"] + "/?action=get_file&name=" + url.QueryEscape(filename)
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
	c := time.Tick(settings.Update)
	for _ = range c {
		if stop {
			return
		}
		if files :=checkList(); len(files) != 0 {
			ext := ""
			if runtime.GOOS == "windows" {
				ext = ".exe"
			}
			pid_file := "../etc/EccReportAgent" + ext + ".pid"
			pid_byte, _ := ioutil.ReadFile(pid_file)
			pid_string := string(pid_byte)
			pid, _ := strconv.Atoi(pid_string)
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
			logger.Println("Killed PID:", pid_string, " from File", pid_file)
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
	done := make(chan bool, 1)
	logfile,err := os.OpenFile("../log/update.log", os.O_CREATE | os.O_APPEND, 0666) 
	if err!=nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	logger := log.New(logfile,"",log.Ldate|log.Ltime|log.Lshortfile)
	defer logfile.Close()
	cc := make(chan os.Signal, 1)
	signal.Notify(cc, os.Interrupt, os.Kill)
	go func(){
	    for _ = range cc {
	    	stop = true
	    	time.Sleep(time.Second)
	        logger.Println("Updater stopped")
	        done <- true
	        os.Exit(0)
	    }
	}()
	logger.Println("Updater started")
	settings, err = utils.LoadSettings()
	if err != nil {
		logger.Fatalln(err)
		os.Exit(1)
	}
	stopAndUpdate()
	<- done
}
