package utils

import (
	"os"
	"fmt"
	"log"
	"net"
	"time"
	"strings"
	//"net/url"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

type Settings struct {
	InModules []map[string] string
	OutModules []map[string] string
	UpdateServer []map[string] string
	Interval time.Duration	// nano seconds
	Hb time.Duration
	Update time.Duration
}

/* Load settings */
func LoadSettings() (Settings, error) {
	var settings Settings
	bytes, err := ioutil.ReadFile("../etc/settings.txt")
	if err != nil {
		return settings, err
	}
	err = json.Unmarshal(bytes, &settings)
	if err != nil {
		return settings, err
	}
	return settings, nil
}

/*
	Get specific IP address (starts with "192.168") and the hostname.
	Return: IP and hostname
*/
func GetLocalInfo() (ip string, hostName string, err error) {
	addrs,err := net.InterfaceAddrs()
    for _, ad := range addrs {
    	if tmp := strings.Split(ad.String(),"/")[0]; strings.HasPrefix(tmp, "192.168") {
    		ip = tmp
    		break
    	}
    }
	hostName, _ = os.Hostname()
	return ip, hostName, err
}

/*
	Do GET reuqest. Returns a slice of byte.
	If the hostHeader string for a module is "" then we use no hostHeader for it.
*/
func ReadRemote(urlString string, hostHeader string) (b []byte, err error) {
	req, _ := http.NewRequest("GET", urlString, nil)
	client := &http.Client{}
	if hostHeader != "" {
		req.Header.Set("Host", hostHeader)
	} 
	res, err := client.Do(req)
	if err != nil {
	    return b, err
	}
	resp, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
	    return b, err
	}
	b = resp
	return b, nil
}

/* Initiate and return a logger by the filename passed in */
func InitLogger(filename string) *log.Logger {
	logfile,err := os.OpenFile(filename, os.O_CREATE | os.O_RDWR | os.O_APPEND, 0666) 
	if err!=nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	logger := log.New(logfile,"\r\n",log.Ldate|log.Ltime|log.Lshortfile)
	return logger
}
