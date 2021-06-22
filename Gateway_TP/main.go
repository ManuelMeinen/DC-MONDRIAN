package main

import (
	"gateway_tp/config"
	"gateway_tp/forwarder"
	"gateway_tp/keyman"
	"gateway_tp/logger"
	"log"
	"time"
	//"gateway_tp/types"
	//"time"
)

func main() {
	config.Init("config/config.json")
	logger.InitLogger()
	defer logger.CloseLogger()

	// Test out some stuff
	//f := fetcher.NewFetcher(config.TPAddr, 10)
	//ip_addr, _, _ := net.ParseCIDR("20.0.1.123/8")
	//time.Sleep(1*time.Second)
	//fmt.Println("OK")
	//zoneID, tpAddr, err := f.GetZoneAndSite(ip_addr)
	//if err!=nil{
	//	fmt.Println("No zone found")
	//	fmt.Println(err)
	//}
	//fmt.Println(zoneID)
	//fmt.Println(tpAddr)
	//zoneID, tpAddr, err = f.GetZoneAndSite(ip_addr)
	//if err!=nil{
	//	fmt.Println("No zone found")
	//}
	//fmt.Println(zoneID)
	//fmt.Println(tpAddr)
	//sites := f.GetSites()
	//for _, site := range sites{
	//	fmt.Println(site.Name)
	//}

	fwd := forwarder.NewForwarder()
	fwd.Start()
	defer fwd.Stop()

	km := keyman.NewKeyMan([]byte("master_secret"))
	log.Println("waiting for 20 sec so the other Gateway TPs are started too before running tests...")
	time.Sleep(20*time.Second)
	log.Println("... done waiting")

	
	finalkey, err := km.GetKey("30.0.0.2", "30.0.0.3", 1)
	if err!=nil{
		log.Println(err)
	}
	log.Println(finalkey)

	finalkey, err = km.GetKey("30.0.0.2", "30.0.0.3", 1)
	if err!=nil{
		log.Println(err)
	}
	log.Println(finalkey)




	for {
		// loop forever
	}

}
