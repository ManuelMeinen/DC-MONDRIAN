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
	
	// Init Sites
	sites = append(sites, &types.Site{"200.0.0.1", "Site 1"})
	sites = append(sites, &types.Site{"200.0.0.2", "Site 2"})
	sites = append(sites, &types.Site{"200.0.0.3", "Site 3"})
	
	// Init Subnets
	subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("10.0.1.0"), Mask: net.IPv4Mask(255, 255, 255, 0)}, 1, "200.0.0.1"})
	subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("10.1.0.0"), Mask: net.IPv4Mask(255, 255, 0, 0)}, 1, "200.0.0.1"})
	subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("10.2.0.0"), Mask: net.IPv4Mask(255, 255, 0, 0)}, 1, "200.0.0.1"})

	subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("20.0.1.0"), Mask: net.IPv4Mask(255, 255, 255, 0)}, 1, "200.0.0.2"})
	subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("20.2.0.0"), Mask: net.IPv4Mask(255, 255, 0, 0)}, 2, "200.0.0.2"})

	subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("30.0.1.0"), Mask: net.IPv4Mask(255, 255, 255, 0)}, 1, "200.0.0.3"})
	subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("30.1.0.0"), Mask: net.IPv4Mask(255, 255, 0, 0)}, 2, "200.0.0.3"})
	subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("30.2.0.0"), Mask: net.IPv4Mask(255, 255, 0, 0)}, 3, "200.0.0.3"})
	// Init Transitions
	// Transition is the type of a policy
	// Transition{PolicyID uint, Src ZoneID, Dest ZoneID, SrcPort uint, DestPort uint, Proto string, Action string}
	transitions = append(transitions, &types.Transition{1, 1, 2, 70, 90, "TCP", "forwarding"})
	transitions = append(transitions, &types.Transition{2, 2, 1, 70, 90, "UDP", "forwarding"})
	transitions = append(transitions, &types.Transition{3, 1, 2, 0, 0, "TCP", "forwarding"})
	transitions = append(transitions, &types.Transition{4, 3, 0, 0, 0, "", "drop"})
	transitions = append(transitions, &types.Transition{5, 1, 2, 80, 100, "TCP", "established"})
	transitions = append(transitions, &types.Transition{6, 1, 2, 80, 0, "TCP", "drop"}) // less specific 
	transitions = append(transitions, &types.Transition{7, 2, 1, 0, 100, "UDP", "established"})
	transitions = append(transitions, &types.Transition{8, 1, 3, 0, 0, "", "established"})

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
