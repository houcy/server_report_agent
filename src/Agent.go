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
	//"os/signal"
	"encoding/json"
)

var settings Settings
var logger *log.Logger

type Settings struct {
	InModules []map[string] string
	OutModules []map[string] string
	Interval int
	Hb int
}

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

func getLocalInfo() (ip string, hostName string) {
	addrs,err := net.InterfaceAddrs()
    if err != nil {
        logger.Println(err.Error())
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

func runInModule(name string, channel chan string) {
	splitted := strings.Split(name, " ")
	cmd := exec.Command(splitted[0], splitted[1:]...)
	buf, err := cmd.Output()
	if err != nil {
		logger.Println(err.Error())
	}
	channel <- string(buf)
}

func inModules() []string {
	numModules := len(settings.InModules)
	var rcvs [16]chan string
	for index, mod := range settings.InModules {
		var prepend string
		var ext string
		modName := mod["name"]
		rcvs[index] = make(chan string)
		switch runtime.GOOS {
			case "windows":
				prepend = "cscript.exe /nologo ../mod/"
				ext = ".vbs"
			default:
				prepend = "/bin/bash ../mod/"
				ext = ".sh"
		}
		command := prepend + modName + ext
		go runInModule(command, rcvs[index])
	}
	var tmp, results [16]string
	for i := 0; i<numModules; i++ {
		tmp[i] = <-rcvs[i]
	}
	for i := 0; i<numModules; i++ {
		results[i] = prepareOutput(settings.InModules[i]["bid"], tmp[i]) 
	}
	return results[:numModules]
}

func prepareOutput(bid string, output string) string {
	timeStamp := time.Now().Unix()
	dateNow := time.Now().Format("20060102")
	ip, hostName := getLocalInfo()
	content := fmt.Sprintf("%s\t%d\t%s\t%s\t%s", dateNow, timeStamp, ip, hostName, output)
	result := fmt.Sprintf("bid=%s&time=%d&content=%s", bid, timeStamp, url.QueryEscape(content))
	return result
}

func outModules(srcString string) {
	for i, _ := range settings.OutModules {
		//fmt.Printf("%s\n", srcString)
		urlString := settings.OutModules[i]["url"] + "?" + srcString
		req, _ := http.NewRequest("GET", urlString, nil)
		client := &http.Client{}
		proxyString := settings.OutModules[i]["proxy"]
		if proxyString != "" {
			proxy, err := url.Parse(proxyString)
			if err != nil {
			    logger.Println(err.Error())
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
		}
		resp, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
		    logger.Println(err.Error())
		}
		fmt.Printf("%s\n", resp)
	}
}

func invokeModules() {
	inModuleReturns := inModules()
	for _, srcString := range inModuleReturns {
		outModules(srcString)
	}
}

func heartbeat() {
	outModules(prepareOutput("1007", "alive"))
	c := time.Tick(30 * time.Second)
	for _ = range c {
		outModules(prepareOutput("1007", "alive"))
	}
}

func main() {
	logfile,err := os.OpenFile("../log/agent.log",os.O_APPEND|os.O_CREATE,0)
	if err!=nil {
		logger.Fatalln(err.Error())
		os.Exit(1)
	}
	/*cc := make(chan os.Signal, 1)
	signal.Notify(cc, os.Interrupt)
	go func(){
	    for _ = range cc {
	        logger.Println("Agent stopped")
	    }
	}()*/
	defer logfile.Close()
	logger = log.New(logfile,"",log.Ldate|log.Ltime|log.Lshortfile)
	logger.Println("Agent started")
	loadSettings()
	logger.Printf("Settings: %v", settings)
	go heartbeat()
	go invokeModules()
	c := time.Tick(300 * time.Second)
	for _ = range c {
		go invokeModules()
	}
}

