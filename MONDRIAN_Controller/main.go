package main

import (
	"controller/config"
	"controller/db"
	"controller/handler"
	"flag"
	"fmt"
	"log"
	"net/http"

	
)

// This function starts the Controller
func startController() {
	dbPath := flag.String("db", config.DbPath, "Path to the database file")
	addr := flag.String("controllerAddr", config.ControllerAddr, "Address of the Mondrian Controller")
	listen := flag.String("controllerPort", fmt.Sprintf(":%s", config.ControllerPort), "Controller listen port")
	flag.Parse()
	config.DbPath = *dbPath
	config.ControllerAddr = *addr
	config.ControllerPort = *listen
	
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