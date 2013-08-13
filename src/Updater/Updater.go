package main 

import (
	"os"
	"log"
	"utils"
)

var logger *log.Logger


func main() {
	logfile,err := os.OpenFile("../log/update.log", os.O_CREATE | os.O_RDWR | os.O_APPEND, 0666)
	if err!=nil {
		logger.Fatalln(err.Error())
		os.Exit(1)
	}
	logger = log.New(logfile,"",log.Ldate|log.Ltime|log.Lshortfile)
	defer logfile.Close()
	
	ip, hostname, err := utils.GetLocalInfo()
	if err != nil {
		logger.Fatalln("aaa")
	}
	logger.Println(ip, hostname)
	
}

