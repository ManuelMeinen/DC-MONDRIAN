package main

import (
	"gateway_tp/config"
	"gateway_tp/forwarder"
	//"gateway_tp/keyman"
	"gateway_tp/logger"
	//"gateway_tp/fetcher"
	//"log"
	//"time"
	//"net"
	//"gateway_tp/types"
	"time"
)

func main() {
	config.Init("config/config.json")

	logger.InitLogger()
	defer logger.CloseLogger()

	fwd := forwarder.NewForwarder()
	fwd.Start()
	defer fwd.Stop()

	for {
		// Sleep for 292 years
    	time.Sleep(time.Duration(1<<63 - 1))
	}

}
