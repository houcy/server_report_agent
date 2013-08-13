package utils

import (
	"os"
	"net"
	"time"
	"strings"
	"io/ioutil"
	//"net/http"
	"encoding/json"
)

type Settings struct {
	InModules []map[string] string
	OutModules []map[string] string
	Interval time.Duration	// nano seconds
	Hb time.Duration
}

/* Load settings or die */
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
