package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// Here are some constants that can be reconfigured to the user's needs
var BASE_PATH string //"/home/mmeinen/polybox/code/DC-MONDRIAN"//TODO: change this if run from somewhere else
var ControllerAddr = "172.17.0.1" // IP-Address of docker0
var ControllerPort = "4433"
var ClientCert = BASE_PATH+"/Gateway_TP/certs/client_cert.pem"
var ClientKey = BASE_PATH+"/Gateway_TP/certs/client_key.pem"
var ServerCert = BASE_PATH+"/Gateway_TP/certs/server_cert.pem"
var ServerKey = BASE_PATH+"/Gateway_TP/certs/server_key.pem"
var HostName string
var LogDir string
var TPAddr string
var KeyLength int = 16
var KeyTTL time.Duration = 24 * time.Hour
var MaxTimeDiff time.Duration = 1 * time.Second
var ServerPort = 9090


type Cfg struct{
	TPAddr string 	`json:"tp_addr"`
	Hostname string `json:"hostname"`
	LogDir string	`json:"log_dir"`
	BasePath string `json:"base_path"`
}

func Init(config_path string){
	jsonFile, err := os.Open(config_path)
	if err!=nil{
		fmt.Println("[1] Failed to read "+config_path)
	}
	data, err := ioutil.ReadAll(jsonFile)
	defer jsonFile.Close()
	if err!=nil{
		fmt.Println("[2] Failed to read "+config_path)
	}
	 // json data
	 var config Cfg
	 // unmarshall it
	 err = json.Unmarshal(data, &config)
	 if err != nil {
		 fmt.Println("error:", err)
	 }
	 // Set it as var such that we don't need to access the struct every time
	 HostName = config.Hostname
	 LogDir = config.LogDir
	 TPAddr = config.TPAddr
	 BASE_PATH = config.BasePath
}