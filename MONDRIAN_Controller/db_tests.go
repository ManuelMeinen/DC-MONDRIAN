package main

import (
	"controller/db"
	"controller/types"
	"encoding/json"
	"fmt"
	"net"
	_ "github.com/mattn/go-sqlite3"
)

var zones types.Zones
var sites types.Sites
var subnets types.Subnets
var transitions types.Transitions

func main() {
	db, err := db.New(":memory:")
	if err != nil {
		panic(err)
	}
	init_vars()
	err = test_json()
	err = test_reads(db)
	err = test_insertions(db)
	err = test_reads(db)
	err = test_deletes(db)
	err = test_reads(db)

}

// Initialize global variables
func init_vars() {

	// Init Zones
	zones = append(zones, &types.Zone{10, "Zone 10"})
	zones = append(zones, &types.Zone{20, "Zone 20"})
	// Init Sites
	sites = append(sites, &types.Site{"1.2.3.4", "Main DC"})
	sites = append(sites, &types.Site{"2.3.4.5", "Main DC 2"})
	// Init Subnets
	subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("1.1.1.1"), Mask: net.IPv4Mask(255, 255, 255, 0)}, 10, "1.2.3.4"})
	subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("1.1.1.2"), Mask: net.IPv4Mask(255, 255, 255, 0)}, 20, "2.3.4.5"})
	// Init Transitions
	transitions = append(transitions, &types.Transition{10, 10, 20, 0, 0, "", "allow"})
	transitions = append(transitions, &types.Transition{20, 10, 20, 0, 0, "", "drop"})
}

// Test to marshal the structs into Json
func test_json() error {
	// Marshal Zones
	fmt.Println((zones))
	b, err := json.Marshal(zones)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))

	// Marshal Sites
	fmt.Println((sites))
	b, err = json.Marshal(sites)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))

	// Marshal Subnets
	fmt.Println((subnets))
	b, err = json.Marshal(subnets)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))

	// Marshal Transitions
	fmt.Println((transitions))
	b, err = json.Marshal(transitions)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))

	fmt.Println("Json Test Done")
	return nil
}

// Test insertions into the Backend
func test_insertions(db *db.Backend) error {
	// Insert Zones
	err := db.InsertZones(zones)
	if err != nil {
		panic(err)
	}

	// Insert Sites
	err = db.InsertSites(sites)
	if err != nil {
		panic(err)
	}

	// Insert Subnets
	err = db.InsertSubnets(subnets)
	if err != nil {
		panic(err)
	}

	// Insert Transitions
	err = db.InsertTransitions(transitions)
	if err != nil {
		panic(err)
	}
	fmt.Println("Insertions Done")
	return nil
}

// Test insertions into the Backend
func test_reads(db *db.Backend) error {

	// Read Zones
	zones, err := db.GetAllZones()
	if err != nil {
		panic(err)
	} else {
		fmt.Println("GetAllZones")
		fmt.Println(zones)
	}

	// Read Sites
	sites, err := db.GetAllSites()
	if err != nil {
		panic(err)
	} else {
		fmt.Println("GetAllSites")
		fmt.Println(sites)
	}

	// Read Subnets
	subnets, err := db.GetAllSubnets()
	if err != nil {
		panic(err)
	} else {
		fmt.Println("GetAllSubnets")
		fmt.Println(subnets)
	}

	// Read some Subnets
	subnets, err = db.GetSubnets("2.3.4.5")
	if err != nil {
		panic(err)
	} else {
		fmt.Println("GetSubnets")
		fmt.Println(subnets)
	}

	// Read Transitions
	transitions, err := db.GetAllTransitions()
	if err != nil {
		panic(err)
	} else {
		fmt.Println("GetAllTransitions")
		fmt.Println(transitions)
	}

	// Read some Transitions
	transitions, err = db.GetTransitions("1.2.3.4")
	if err != nil {
		panic(err)
	} else {
		fmt.Println("GetTransitions")
		fmt.Println(transitions)
	}
	fmt.Println("Reads Done")
	return nil
}

// Test deletes from the Backend
func test_deletes(db *db.Backend) error {

	// Delete Zones
	err := db.DeleteZones(zones)
	if err != nil {
		panic(err)
	}

	// Delete Sites
	err = db.DeleteSites(sites)
	if err != nil {
		panic(err)
	}

	// Delete Subnets
	err = db.DeleteSubnets(subnets)
	if err != nil {
		panic(err)
	}

	// Delete Transitions
	err = db.DeleteTransitions(transitions)
	if err != nil {
		panic(err)
	}
	fmt.Println("Delete Done")
	return nil
}
