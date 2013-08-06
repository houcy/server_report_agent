package main 

import (
	"os"
	"fmt"
	"net"
	"time"
	"runtime"
	"strings"
	"os/exec"
	"net/url"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

var settings Settings

type Settings struct {
	InModules []map[string] string
	OutModules []map[string] string
	Interval int
	Hb int
}

func loadSettings() {
	bytes, err := ioutil.ReadFile("../etc/settings.txt")
	if err != nil {
		fmt.Println("error: ", err)
		os.Exit(-1)
	}
	err = json.Unmarshal(bytes, &settings)
	if err != nil {
		fmt.Println("error: ", err)
		os.Exit(-1)
	}
}

func getLocalInfo() (ip string, hostName string) {
	addrs,err := net.InterfaceAddrs()
    if err != nil {
        fmt.Println("error: ", err)
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
		fmt.Println("error: ", err)
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
		res, err := http.Get(settings.OutModules[i]["url"] + "?" + srcString)
		if err != nil {
		    fmt.Println("error: ", err)
		}
		resp, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
		    fmt.Println("error: ", err)
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
		loadSettings()
		outModules(prepareOutput("1007", "alive"))
	}
}

func main() {
	loadSettings()
	go heartbeat()
	go invokeModules()
	c := time.Tick(300 * time.Second)
	for _ = range c {
		go invokeModules()
	}
}

