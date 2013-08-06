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

type Settings struct {
	Modules []string
	Bids []int
	Interval int
	Hb int
}

func loadSettings() (Settings, error) {
	bytes, err := ioutil.ReadFile("../etc/settings.txt")
	if err != nil {
		fmt.Println("error: ", err)
	}
	var settings Settings
	err = json.Unmarshal(bytes, &settings)
	if err != nil {
		fmt.Println("error: ", err)
	}
	return settings, err
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

func inModules(settings Settings) []string {
	numModules := len(settings.Modules)
	var rcvs [16]chan string
	for index, modName := range settings.Modules {
		var prepend string
		var ext string
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
		results[i] = prepareOutput(settings.Bids[i], tmp[i]) 
	}
	return results[:]
}

func prepareOutput(bid int, output string) string {
	timeStamp := time.Now().Unix()
	dateNow := time.Now().Format("20060102")
	ip, hostName := getLocalInfo()
	content := fmt.Sprintf("%s\t%d\t%s\t%s\t%s", dateNow, timeStamp, ip, hostName, output)
	result := fmt.Sprintf("bid=%d&time=%d&content=%s", bid, timeStamp, url.QueryEscape(content))
	return result
}

func outModule(srcString string) {
	fmt.Println(srcString)
	res, err := http.Get("http://sh.ecc.com/data/server_view/api/data_collector.php?" + srcString)
	if err != nil {
	    fmt.Println("error: ", err)
	}
	resp, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
	    fmt.Println("error: ", err)
	}
	fmt.Printf("%s", resp)
}

func invokeModules(settings Settings) {
	inModuleReturns := inModules(settings)
	for _, srcString := range inModuleReturns {
		outModule(srcString)
	}
}

func heartbeat(settings Settings) {
	c := time.Tick(30 * time.Second)
	for _ = range c {
		outModule(prepareOutput(1007, "alive"))
	}
}

func main() {
	settings, err := loadSettings()
	if err != nil {
		fmt.Println("error: Cannot load settings.")
	}
	fmt.Println(time.Now().Unix())
	go heartbeat(settings)
	go invokeModules(settings)
	c := time.Tick(300 * time.Second)
	for _ = range c {
		fmt.Println(time.Now().Unix())
		go invokeModules(settings)
	}
	
}

