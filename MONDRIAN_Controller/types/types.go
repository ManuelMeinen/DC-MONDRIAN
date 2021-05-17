package types

import (
	"encoding/json"
	"net"
)

// ZoneID represents 24bit zone identifiers
// TODO: possible check to make sure IDs are always in the range [0, 1<<24-1]
type ZoneID uint

// Zone denotes a network zone
type Zone struct {
	ID   ZoneID
	Name string
	//	Subnets []*Subnet
}

// Site denotes a branch site of the network
type Site struct {
	TPAddr string
	Name   string
}

// Subnet is an IP subnet that is located behind a TP
type Subnet struct {
	IPNet  net.IPNet
	ZoneID ZoneID
	TPAddr string
}

// Transition is the type of a policy
type Transition struct {
	PolicyID uint
	Src ZoneID
	Dest ZoneID
	SrcPort uint
	DestPort uint
	Proto string
	Action string
}

func (s Subnet) MarshalJSON() ([]byte, error) {
	dummy := struct {
		CIDR   string
		ZoneID ZoneID
		TPAddr string
	}{
		CIDR:   s.IPNet.String(),
		ZoneID: s.ZoneID,
		TPAddr: s.TPAddr,
	}

	return json.Marshal(dummy)
}

func (s *Subnet) UnmarshalJSON(b []byte) error {
	dummy := struct {
		CIDR   string
		ZoneID ZoneID
		TPAddr string
	}{}
	err := json.Unmarshal(b, &dummy)
	if err != nil {
		return err
	}
	_, net, err := net.ParseCIDR(dummy.CIDR)
	if err != nil {
		return err
	}
	s.IPNet = *net
	s.ZoneID = dummy.ZoneID
	s.TPAddr = dummy.TPAddr
	return nil
}

// Network implements the RangerEntry interface for use with github.com/yl2chen/cidranger
func (s *Subnet) Network() net.IPNet {
	return s.IPNet
}

// Zones is a list of zones
type Zones []*Zone

// Sites is a list of sites
type Sites []*Site

// Transitions is a list of transitions
type Transitions []*Transition

// Subnets is a list of IP subnets
type Subnets []*Subnet



