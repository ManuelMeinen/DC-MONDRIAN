package fetcher

import (
	"fmt"
	"gateway_tp/config"
	"gateway_tp/types"
	"gateway_tp/api"
	"net"
	"time"
	"errors"
	"sync"
)

var subnets types.Subnets = nil
var subnets_lock sync.RWMutex

var sites types.Sites = nil
var sites_lock sync.RWMutex

type Fetcher struct {
	LocalAddr      string
	controllerAddr string
	controllerPort string
	refresh_interval int
	conn         api.Connection
}


func NewFetcher(tpAddr string, refresh_interval int) *Fetcher {
	// NewFetcher returns a fetcher used to fetch stuff from the controller
	f :=  &Fetcher{
			LocalAddr:      tpAddr,
			controllerAddr: config.ControllerAddr,
			controllerPort: config.ControllerPort,
			refresh_interval: refresh_interval,
			conn: *api.GetConn(config.ControllerAddr, config.ControllerPort),	
		}
	f.start()
	return f
}

func (f *Fetcher) start(){
	// Start periodical fetching
	go f.fetch_thread()
}

func (f *Fetcher) fetch_thread(){
	// Non-terminating fetching trhead
	for {
		local_subnets, err1 := f.conn.GetSubnets(f.LocalAddr)
		if err1!=nil{
			fmt.Println("Failed to fetch new subnets from the controller")
		}else{
			subnets_lock.Lock() 
			subnets = local_subnets
			subnets_lock.Unlock()
		}
		local_sites, err2 := f.conn.GetAllSites()
		if err2!=nil{
			fmt.Println("Failed to fetch new sites from the controller")
		}else{
			sites_lock.Lock() 
			sites = local_sites
			sites_lock.Unlock()
		}
		if err1==nil && err2==nil{
			fmt.Print(".") //Indicate success
		}
		time.Sleep(time.Duration(f.refresh_interval) * time.Second)		
	}
}

func (f *Fetcher) GetZoneAndSite(ipAddr net.IP)(uint, string, error){
	// For a given Dest IP-Address find the Zone and the Dest Site
	subnets_lock.RLock()
	defer subnets_lock.RUnlock()
	for _, subnet := range subnets {
		if subnet.IPNet.Contains(ipAddr) {
			// Subnet found 
			return uint(subnet.ZoneID), subnet.TPAddr, nil
		}
	}
	return 0, "", errors.New("ERROR: No Zone found")
}

func (f *Fetcher) GetSites()types.Sites{
	// Return all sites 
	// Note: not sure if they are copied or just referenced... --> only read from them
	sites_lock.RLock()
	defer sites_lock.RUnlock()
	return sites
}



