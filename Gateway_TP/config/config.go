package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Here are some constants that can be reconfigured to the user's needs
var BASE_PATH = "/home/mmeinen/polybox/code/DC-MONDRIAN"//TODO: change this if run from somewhere else
var ControllerAddr = "172.17.0.1" // IP-Address of docker0
var ControllerPort = "4433"
var ClientCert = BASE_PATH+"/Gateway_TP/certs/client_cert.pem"
var ClientKey = BASE_PATH+"/Gateway_TP/certs/client_key.pem"
var HostName string
var LogDir string
var TPAddr string

type Cfg struct{
	TPAddr string 	`json:"tp_addr"`
	Hostname string `json:"hostname"`
	LogDir string	`json:"log_dir"`
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
}