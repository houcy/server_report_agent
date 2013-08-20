package main 

import (
	"net/http"
	"net/url"
	"strings"
	"strconv"
	"utils"
	"fmt"
	"log"
	"os"
	"io"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var logger *log.Logger

func dealRequest(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	action := r.FormValue("action")
    switch action {
    	case "get_list":
    		ip := r.FormValue("ip")
    		getList(ip, w)
    	case "get_file":
    		filename := r.FormValue("name")
			version := r.FormValue("v")
    		getFile(filename, version, w)
    }
    
}

func getList(ip string, w http.ResponseWriter) {
	db,err := sql.Open("mysql", "root:@tcp(localhost:3306)/agentserver?charset=utf8")
	if err != nil {
		logger.Println("cannot open database")
		fmt.Fprintf(w, "")
		return
	}
	rows,err := db.Query("select v,files from version where v=(select v from machine where ip=? and done=0)", ip)
	if err != nil {
		logger.Println("cannot query table version")
		fmt.Fprintf(w, "")
		return
	}
	_, err = db.Exec("update machine set done=1 where ip=? and done=0", ip)
	if err != nil {
		logger.Println("cannot update table machine")
		fmt.Fprintf(w, "")
		return
	}
	for rows.Next() {
		var v,files string
		err = rows.Scan(&v, &files)
		if err == nil {
			fmt.Fprintf(w, "%s;%s", v, files)
			break
		}
	}
	db.Close()
	return
}

func getFile(filename string, version string, w http.ResponseWriter) {
	name, _ := url.QueryUnescape(filename)
	_, err := strconv.Atoi(version)
	if strings.Contains(name, "..") || err!=nil {
		logger.Println("invalid query.")
		fmt.Fprintf(w, "")
		return
	}
	outFile := "../up/" + version + "/" + name
	f, err := os.Open(outFile)
	if nil != err && !os.IsExist(err) {
		logger.Println(err.Error())
		fmt.Fprintf(w, "")
		return
	}
	defer f.Close()
	io.Copy(w, f)
}

func main() {
	os.Mkdir("../log/", 0666)
	logger = utils.InitLogger("../log/server.log")
	http.HandleFunc("/", dealRequest)
	err := http.ListenAndServe(":9090", nil)
    if err != nil {
        logger.Fatalln("ListenAndServe: ", err)
    }
}

