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
	log.Println("waiting...")
	time.Sleep(20*time.Second)
	log.Println("... done waiting")

	//TODO only works if invoked the second time ... --> fix the bug
	log.Println("Derive L1 Key for 30.0.0.2")
	key, err := km.DeriveL1Key("30.0.0.2")
	
	if err!=nil{
		log.Println(err)
	}
	log.Println(key)
	log.Println("Derive L2 Key for Zone 1")
	l2_key, err := km.DeriveL2(key.Key, 1)
	if err!=nil{
		log.Println(err)
	}
	log.Println(l2_key)
	
	log.Println("Fetch L1 from Remote")
	key, err = km.FetchL1FromRemote("30.0.0.2")
	
	if err!=nil{
		log.Println(err)
	}
	log.Println("Fetching resulted in:")
	log.Println(key)
	
	log.Println("Derive L2 key with fetched key")
	l2_key, err = km.DeriveL2(key.Key, 1)
	if err!=nil{
		log.Println(err)
	}
	log.Println(l2_key)
	




	for {
		// loop forever
	}

}
