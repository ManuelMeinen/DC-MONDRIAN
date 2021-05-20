package main

import (
	"flag"
    "fmt"
    "log"
    "net/http"
	"controller/handler"
	"controller/db"
	"controller/config"
)

// This function starts the Controller
func startController() {
	dbPath := flag.String("db", config.DbPath, "path to the database file")
	listen := flag.String("listen", fmt.Sprintf(":%s", config.ControllerPort), "server listen address")
	flag.Parse()
	fmt.Println(*dbPath)
	db.SetupDB(*dbPath)
	for api, handler := range handler.ApiMap {
        http.HandleFunc(api, handler)
    }
	fmt.Println("*** Controller Ready ***")
	fmt.Println(fmt.Sprintf("Listening at: https://%s:%s/", config.ControllerAddr, config.ControllerPort))
    log.Fatal(http.ListenAndServeTLS(*listen, config.ServerCert, config.ServerKey, nil))
}

func main() {
	handler.Init()
    startController()
}