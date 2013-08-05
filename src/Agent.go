package main 

import (
	"fmt"
	"time"
	"runtime"
	"strings"
	"os/exec"
	"io/ioutil"
	"encoding/json"
)

type Settings struct {
	In_modules []string
	Out_modules []string
	Interval int
	Hb int
}

func loadSettings() Settings {
	bytes, err := ioutil.ReadFile("../etc/settings.txt")
	if err != nil {
		fmt.Println("error: ", err)
	}
	var settings Settings
	err = json.Unmarshal(bytes, &settings)
	if err != nil {
		fmt.Println("error: ", err)
	}
	return settings
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

func invokeInModules(settings Settings) []string {
	numModules := len(settings.In_modules)
	var rcvs [16]chan string
	for index, modName := range settings.In_modules {
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
		fmt.Println(index, ":", modName)
		go runInModule(command, rcvs[index])
	}
	var results [16]string
	for i := 0; i<numModules; i++ {
		results[i] = (<-rcvs[i])
	}
	return results[:]
}

func main() {
	settings := loadSettings()
	t := time.Now()
	fmt.Println(t.Unix())
	for i := 0; i < 2; i++ {
		results := invokeInModules(settings)
		for _, r := range results {
			fmt.Println(r)
		}
	}
	t = time.Now()
	fmt.Println(t.Unix())
}

