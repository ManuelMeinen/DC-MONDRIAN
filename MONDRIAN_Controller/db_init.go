package main

import (
	"controller/config"
	"controller/db"
	"controller/types"
	"net"
	_ "github.com/mattn/go-sqlite3"
)


// Use this file to initialize a dummy database
func main() {
	var zones types.Zones
	var sites types.Sites
	var subnets types.Subnets
	var transitions types.Transitions
	
	db, err := db.New(config.DbPath)
	if err != nil {
		panic(err)
	}
	// Init Zones
	zones = append(zones, &types.Zone{1, "Zone 1"})
	zones = append(zones, &types.Zone{2, "Zone 2"})
	zones = append(zones, &types.Zone{3, "Zone 3"})
	zones = append(zones, &types.Zone{4, "Zone 4"})
	// Init Sites
	sites = append(sites, &types.Site{"1.2.3.4", "Site 1"})
	sites = append(sites, &types.Site{"2.3.4.5", "Site 2"})
	sites = append(sites, &types.Site{"3.4.5.6", "Site 3"})
	sites = append(sites, &types.Site{"4.5.6.7", "Site 4"})
	// Init Subnets
	subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("192.168.0.1"), Mask: net.IPv4Mask(255, 255, 255, 255)}, 1, "1.2.3.4"})
	subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("192.168.2.0"), Mask: net.IPv4Mask(255, 255, 255, 0)}, 2, "2.3.4.5"})
	subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("192.3.0.0"), Mask: net.IPv4Mask(255, 255, 0, 0)}, 3, "3.4.5.6"})
	subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("4.0.0.0"), Mask: net.IPv4Mask(255, 0, 0, 0)}, 4, "4.5.6.7"})
	// Init Transitions
	transitions = append(transitions, &types.Transition{1, 1, 2, 80, 100, "http", "allow"})
	transitions = append(transitions, &types.Transition{2, 2, 1, 80, 100, "ftp", "drop"})
	transitions = append(transitions, &types.Transition{3, 1, 2, 0, 0, "http", "allow"})
	transitions = append(transitions, &types.Transition{4, 3, 4, 100, 0, "", "allow"})
	transitions = append(transitions, &types.Transition{5, 1, 0, 80, 100, "http", "allow"})
	transitions = append(transitions, &types.Transition{6, 0, 2, 80, 100, "http", "allow"})

	// Insert stuff
	err = db.InsertZones(zones)
	if err != nil {
		panic(err)
	}
	err = db.InsertSites(sites)
	if err != nil {
		panic(err)
	}
	err = db.InsertSubnets(subnets)
	if err != nil {
		panic(err)
	}
	err = db.InsertTransitions(transitions)
	if err != nil {
		panic(err)
	}
}
