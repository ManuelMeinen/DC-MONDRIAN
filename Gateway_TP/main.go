package main

import (
	"fmt"
	"gateway_tp/config"
	"gateway_tp/fetcher"
	"gateway_tp/forwarder"
	"gateway_tp/logger"
	"log"
	"net"
	"time"
	//"gateway_tp/types"
	//"time"
)

func main(){
	config.Init("config/config.json")
	logger.InitLogger()
	defer logger.CloseLogger()
	log.Println("test")
	f := fetcher.NewFetcher(config.TPAddr, 10)
	ip_addr, _, _ := net.ParseCIDR("20.0.1.123/8")
	time.Sleep(1*time.Second)
	fmt.Println("OK")
	zoneID, tpAddr, err := f.GetZoneAndSite(ip_addr)
	if err!=nil{
		fmt.Println("No zone found")
		fmt.Println(err)
	}
	fmt.Println(zoneID)
	fmt.Println(tpAddr)
	zoneID, tpAddr, err = f.GetZoneAndSite(ip_addr)
	if err!=nil{
		fmt.Println("No zone found")
	}
	fmt.Println(zoneID)
	fmt.Println(tpAddr)
	sites := f.GetSites()
	for _, site := range sites{
		fmt.Println(site.Name)
	}
	//fwd := forwarder.NewForwarder(f)
	//defer fwd.Close_conns()
	//fmt.Println(fwd)
	internal_iface := forwarder.NewIface(config.HostName+"-eth0")
	defer internal_iface.Close()
	external_iface := forwarder.NewIface(config.HostName+"-eth1")
	defer external_iface.Close()
	go internal_iface.Process_Packets(external_iface)
	go external_iface.Process_Packets(internal_iface)
	for{
		// loop forever
	}


	//iface := forwarder.NewIface("lo")
	//fmt.Println(iface)
	//iface.Process_Packets()
	//forwarder.Test()
	//for {
//
	//}
}