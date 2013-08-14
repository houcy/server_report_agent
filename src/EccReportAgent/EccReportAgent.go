package main 

import (
	"os"
	"fmt"
	"log"
	"time"
	"runtime"
	"strings"
	"os/exec"
	"net/url"
	"os/signal"
	"utils"
)

var settings utils.Settings
var logger *log.Logger
var stop bool

/* Send heartbeat signal to server */
func heartbeat() {
	outModules(prepareOutput("1007", "alive"))
	c := time.Tick(settings.Hb)
	for _ = range c {
		outModules(prepareOutput("1007", "alive"))
	}
}

func runInModule(name string, channel chan string) {
	splitted := strings.Split(name, " ")
	cmd := exec.Command(splitted[0], splitted[1:]...)
	buf, err := cmd.Output()
	if err != nil {
		channel <- string(buf) + "error"
	}
	channel <- string(buf)
}

/*
	Invoke scripts according to the OS type
	Return: A slice of strings, which includes contents returned by the scripts.
*/
func inModules() []string {
	numModules := len(settings.InModules)
	var rcvs [16]chan string
	for index, mod := range settings.InModules {
		var prepend string
		var ext string
		modPath := "../mod/"
		modName := mod["name"]
		rcvs[index] = make(chan string)
		switch runtime.GOOS {
			case "windows":
				if mod["windows"] != "1" {
					continue 
				}
				prepend = "cscript.exe /nologo "
				ext = ".vbs"
			case "linux":
				if mod["linux"] != "1" { 
					continue 
				}
				prepend = "/bin/bash "
				ext = ".sh"
		}
		modPathName := modPath + modName + ext
		_, err := os.Stat(modPathName)
		if nil != err && !os.IsExist(err) {
			logger.Println(modPathName, "doesn't exist. Ignored.")
			continue
		}
		command := prepend + modPathName
		go runInModule(command, rcvs[index])
	}
	var tmp, results [16]string
	for i := 0; i<numModules; i++ {
		// Here we wait until all of the scripts we invoked return,
		// so the output time will be determined by the script that costs the longest time 
		tmp[i] = <-rcvs[i]
	}
	for i := 0; i<numModules; i++ {
		results[i] = prepareOutput(settings.InModules[i]["bid"], tmp[i]) 
	}
	return results[:numModules]
}

/*
	Format the output string to a standard one
	Param: bid, of the monitoring content; output, original content string.
	Return: Formatted string
*/
func prepareOutput(bid string, output string) string {
	
	timeStamp := time.Now().Unix()
	dateNow := time.Now().Format("20060102")
	ip, hostName, err := utils.GetLocalInfo()
	if err != nil {
		logger.Println(err.Error())
	}
	content := fmt.Sprintf("%s\t%d\t%s\t%s\t%s", dateNow, timeStamp, ip, hostName, output)
	result := fmt.Sprintf("bid=%s&time=%d&content=%s", bid, timeStamp, url.QueryEscape(content))
	return result
}

/*
	Call ReadRemote and print for debug purpose
*/
func runOutModule(urlString string, proxyString string) {
	resp, err := utils.ReadRemote(urlString, proxyString)
	if err != nil {
	    logger.Println(err.Error())
	    return
	}
	fmt.Printf("%s\n", resp)
}

/*
	Output the final content to all the output modules configured in the setting file,
*/
func outModules(srcString string) {
	if stop {
		return
	}
	for i, _ := range settings.OutModules {
		if strings.HasSuffix(srcString, "error") {
			logger.Printf("%s\n", srcString)
			continue
		}
		urlString := settings.OutModules[i]["url"] + "?" + srcString
		proxyString := settings.OutModules[i]["proxy"]
		go runOutModule(urlString, proxyString)
	}
}

/*
	Invoke input modules and output modules in sequence.
	It is a deadloop and repeat itself in every 300 seconds.
*/
func invokeModules() {
	inModuleReturns := inModules()
	for _, srcString := range inModuleReturns {
		outModules(srcString)
	}
	c := time.Tick(settings.Interval)
	for _ = range c {
		inModuleReturns := inModules()
		for _, srcString := range inModuleReturns {
			outModules(srcString)
		}
	}
}

func main() {
	done := make(chan bool, 1)
	logfile,err := os.OpenFile("../log/agent.log", os.O_CREATE | os.O_RDWR | os.O_APPEND, 0666) 
	if err!=nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	logger := log.New(logfile,"",log.Ldate|log.Ltime|log.Lshortfile)
	defer logfile.Close()
	// code snippet: capture Ctrl-C signal and handle it
	cc := make(chan os.Signal, 1)
	signal.Notify(cc, os.Interrupt, os.Kill)
	go func(){
	    for _ = range cc {
	    	stop = true
	    	time.Sleep(time.Second)
	        logger.Println("Agent stopped")
	        done <- true
	        os.Exit(0)
	    }
	}()
	logger.Println("Agent started")
	settings, err = utils.LoadSettings()
	if err != nil {
		logger.Fatalln(err.Error())
		os.Exit(1)
	}
	logger.Printf("Settings: %v", settings)
	go heartbeat()
	go invokeModules()
	<- done
}
