package main

import (
	"controller/api_test/api"
	"controller/types"
	"fmt"
	"net"
)

// This file shows how an admin can use the client side API to manage the database
func main() {
	api.Conn.Client = api.StartInsecureClient()

	sites := api.Conn.GetAllSites()
	fmt.Println(sites)
	for _, site := range sites {
		fmt.Println(site.Name)
	}

	zones := api.Conn.GetAllZones()
	fmt.Println(zones)
	for _, zone := range zones {
		fmt.Println(zone.Name)
	}

	subnets := api.Conn.GetAllSubnets()
	fmt.Println(subnets)
	for _, subnet := range subnets {
		fmt.Println(subnet.ZoneID)
	}

	transitions := api.Conn.GetAllTransitions()
	fmt.Println(transitions)
	for _, transition := range transitions {
		fmt.Println(transition.Action)
	}

	subnets = api.Conn.GetSubnets(sites[0].TPAddr)
	fmt.Println(subnets)
	for _, subnet := range subnets {
		fmt.Println(subnet.ZoneID)
	}

	transitions = api.Conn.GetTransitions(sites[0].TPAddr)
	fmt.Println(transitions)
	for _, transition := range transitions {
		fmt.Println(transition.Action)
	}

	// Test insertions
	var test_zones types.Zones
	var test_sites types.Sites
	var test_subnets types.Subnets
	var test_transitions types.Transitions
	// Init Zones
	test_zones = append(test_zones, &types.Zone{10, "Zone 10"})
	test_zones = append(test_zones, &types.Zone{20, "Zone 20"})
	// Init Sites
	test_sites = append(test_sites, &types.Site{"10.10.10.10", "Main DC"})
	test_sites = append(test_sites, &types.Site{"20.20.20.20", "Main DC 2"})
	// Init Subnets
	test_subnets = append(test_subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("10.10.10.0"), Mask: net.IPv4Mask(255, 255, 255, 0)}, 10, "10.10.10.10"})
	test_subnets = append(test_subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("20.20.20.0"), Mask: net.IPv4Mask(255, 255, 255, 0)}, 20, "20.20.20.20"})
	// Init Transitions
	test_transitions = append(test_transitions, &types.Transition{10, 10, 20, 0, 0, "", "allow"})
	test_transitions = append(test_transitions, &types.Transition{20, 10, 20, 0, 0, "", "drop"})

	api.Conn.InsertSites(test_sites)
	api.Conn.InsertZones(test_zones)
	api.Conn.InsertSubnets(test_subnets)
	api.Conn.InsertTransitions(test_transitions)
 	// Test deletions
	api.Conn.DeleteZones(test_zones)
	api.Conn.DeleteSites(test_sites)
	api.Conn.DeleteSubnets(test_subnets)
	api.Conn.DeleteTransitions(test_transitions)
}