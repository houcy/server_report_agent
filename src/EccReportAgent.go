package main 

import (
	"os"
	"fmt"
	"log"
	"net"
	"time"
	"runtime"
	"strings"
	"os/exec"
	"net/url"
	"net/http"
	"io/ioutil"
	"os/signal"
	"encoding/json"
)

var settings Settings
var logger *log.Logger
var stop bool

type Settings struct {
	InModules []map[string] string
	OutModules []map[string] string
	Interval time.Duration	// nano seconds
	Hb time.Duration
}

/* Load settings or die */
func loadSettings() {
	bytes, err := ioutil.ReadFile("../etc/settings.txt")
	if err != nil {
		logger.Fatalln(err.Error())
		os.Exit(1)
	}
	err = json.Unmarshal(bytes, &settings)
	if err != nil {
		logger.Fatalln(err.Error())
		os.Exit(1)
	}
}

/*
	Get specific IP address (starts with "192.168") and the hostname.
	Return: IP and hostname
*/
func getLocalInfo() (ip string, hostName string) {
	addrs,err := net.InterfaceAddrs()
    if err != nil {
        logger.Fatalln(err.Error())
        os.Exit(1)
    }
    for _, ad := range addrs {
    	if tmp := strings.Split(ad.String(),"/")[0]; strings.HasPrefix(tmp, "192.168") {
    		ip = tmp
    		break
    	}
    }
	hostName, _ = os.Hostname()
	return ip, hostName
}

/* Send heartbeat signal to server */
func heartbeat() {
	outModules(prepareOutput("1007", "alive"))
	c := time.Tick(30 * time.Second)
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
				prepend = "cscript.exe /nologo "
				ext = ".vbs"
			default:
				prepend = "/bin/bash "
				ext = ".sh"
		}
		modPathName := modPath + modName + ext
		_, err := os.Stat(modPathName)
		if nil != err && !os.IsExist(err) {
			logger.Println(modPathName, "doesn't exist.")
			continue
		}
		command := prepend + modPathName
		go runInModule(command, rcvs[index])
	}
	var tmp, results [16]string
	for i := 0; i<numModules; i++ {
		// Here we wait until all of the scripts we invoked return,
		// so the output time will be determined by the script that uses the longest time 
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
	ip, hostName := getLocalInfo()
	content := fmt.Sprintf("%s\t%d\t%s\t%s\t%s", dateNow, timeStamp, ip, hostName, output)
	result := fmt.Sprintf("bid=%s&time=%d&content=%s", bid, timeStamp, url.QueryEscape(content))
	return result
}

/*
	Do GET reuqest.
	If the proxy string for a module is "" then we use no proxy for it.
*/
func runOutModule(urlString string, proxyString string) {
	req, _ := http.NewRequest("GET", urlString, nil)
	client := &http.Client{}
	if proxyString != "" {
		proxy, err := url.Parse(proxyString)
		if err != nil {
		    logger.Println(err.Error())
		    return
		}
		client = &http.Client{
			Transport: &http.Transport {
				Proxy : http.ProxyURL(proxy),
			},
		}
	} 
	res, err := client.Do(req)
	if err != nil {
	    logger.Println(err.Error())
	    return
	}
	resp, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
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

func invokeModules() {
	inModuleReturns := inModules()
	for _, srcString := range inModuleReturns {
		outModules(srcString)
	}
	c := time.Tick(300 * time.Second)
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
		logger.Fatalln(err.Error())
		os.Exit(1)
	}
	logger = log.New(logfile,"",log.Ldate|log.Ltime|log.Lshortfile)
	defer logfile.Close()
	// code snippet: capture Ctrl-C signal and handle it
	cc := make(chan os.Signal, 1)
	signal.Notify(cc, os.Interrupt)
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
	loadSettings()
	logger.Printf("Settings: %v", settings)
	go heartbeat()
	go invokeModules()
	<- done
}

