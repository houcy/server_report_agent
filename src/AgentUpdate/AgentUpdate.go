package main

import (
	"os"
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
	If the content is not empty, then split the file names split by a semicolon.
	File names are like these: mod/cpu_usage.vbs or bin/EccReportAgent
*/
func checkList() []string {
	var list []string
	ip, _, _ := utils.GetLocalInfo() 
	urlString := settings.UpdateServer[0]["url"] + "/?action=get_list&ip=" + ip
	hostString := settings.UpdateServer[0]["host"]
	resp, err := utils.ReadRemote(urlString, hostString)
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
func downloadAndReplaceFile(version string, filename string) bool {
	theFile := "../" + filename
	if _, err := os.Stat(theFile); err == nil {
		// If local file exists.
	    err = os.Rename(theFile, theFile + "." + strconv.FormatInt(time.Now().Unix(), 10))
		if err != nil {
		    logger.Println(err.Error())
		    return false
		}
	}
	urlString := settings.UpdateServer[0]["url"] + "/?action=get_file&v=" + version + "&name=" + url.QueryEscape(filename)
	hostHeader := settings.UpdateServer[0]["host"]
    resp, err := utils.ReadRemote(urlString, hostHeader)
	if err != nil {
	    logger.Println(err.Error())
	    return false
	}
	if len(resp) != 0 {
		// empty response means no such file exists, we should do nothing.
		f, err := os.OpenFile(theFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		defer f.Close()
		if err != nil {
			logger.Println(err.Error())
			return false
		}
		f.Write(resp)
		return true
	}
	return false
}

/*
	Stop EccReportAgent by its PID recorded by the daemon if updates exist. 
	Then download these files
*/
func stopAndUpdate() {
	c := time.Tick(time.Duration(settings.Update) * time.Second)
	for _ = range c {
		if stop {
			return
		}
		if files :=checkList(); len(files) != 0 {
			logger.Println(files)
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
			// TODO: here we don't have the permission to kill this process. How to solve this?
			if err != nil {
			    logger.Println(err.Error())
			    return
			}
			logger.Println("Killed PID:", pid_string, " from File", pid_file)
			version := files[0]
			for _, filename := range files[1:] {
				if !downloadAndReplaceFile(filename, version) {
					logger.Println("update failed for:", filename)
				}
				time.Sleep(time.Second * 3)
			}
		}
	}
}

func main() {
	done := make(chan bool, 1)
	logger = utils.InitLogger("../log/update.log")
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
	var err error
	settings, err = utils.LoadSettings()
	if err != nil {
		logger.Fatalln(err)
		os.Exit(1)
	}
	stopAndUpdate()
	<- done
}
