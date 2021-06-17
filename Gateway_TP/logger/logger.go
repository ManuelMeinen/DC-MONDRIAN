package logger

import (
	"fmt"
	"gateway_tp/config"
	"log"
	"os"
)

var file *os.File

func InitLogger(){
	// log into a file
	err := os.Remove(config.LogDir+config.HostName+".log")
	if err!=nil{
	 fmt.Println("Log file wasn't deleted")
	}
	file, err = os.OpenFile(config.LogDir+config.HostName+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
	 log.Fatal(err)
	}
 	log.SetOutput(file)
}

func CloseLogger(){
	file.Close()
}