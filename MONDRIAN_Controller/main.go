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
	fmt.Println(fmt.Sprintf("Listening at: https://%s%s/", config.ControllerAddr, config.ControllerPort))
	//Note: if we don't specify the controller address then it listens and serves both on docker0 and lo
    //log.Fatal(http.ListenAndServeTLS(config.ControllerAddr+config.ControllerPort, config.ServerCert, config.ServerKey, nil))
	log.Fatal(http.ListenAndServeTLS(config.ControllerPort, config.ServerCert, config.ServerKey, nil))
}

func main() {
	handler.Init()
    startController()
}