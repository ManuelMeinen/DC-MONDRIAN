package main

import (
	"gateway_tp/config"
	"gateway_tp/forwarder"
	"gateway_tp/logger"
	//"log"
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
	for {
		// loop forever
	}

}
