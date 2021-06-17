package config

// Here are some constants that can be reconfigured to the user's needs
var BASE_PATH = "/home/mmeinen/polybox/code/DC-MONDRIAN"//TODO: change this if run from somewhere else
var DbPath = BASE_PATH+"/MONDRIAN_Controller/backendDB/controllerDB.sqlite"
var ControllerAddr = "172.17.0.1"
var ControllerPort = "4433"
var ServerCert = BASE_PATH+"/MONDRIAN_Controller/certs/server_cert.pem"
var ServerKey = BASE_PATH+"/MONDRIAN_Controller/certs/server_key.pem"