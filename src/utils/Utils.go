package utils

import (
	"os"
	"net"
	"time"
	"strings"
	"net/url"
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
	If the proxy string for a module is "" then we use no proxy for it.
*/
func ReadRemote(urlString string, proxyString string) (b []byte, err error) {
	req, _ := http.NewRequest("GET", urlString, nil)
	client := &http.Client{}
	if proxyString != "" {
		proxy, err := url.Parse(proxyString)
		if err != nil {
		    return b, err
		}
		client = &http.Client{
			Transport: &http.Transport {
				Proxy : http.ProxyURL(proxy),
			},
		}
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
