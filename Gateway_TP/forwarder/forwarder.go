package forwarder

import (
	//"gateway_tp/config"
	"gateway_tp/fetcher"
	//"gateway_tp/chain"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	

	//"github.com/google/gopacket/layers"
	"fmt"
	"time"
	"log"
)

//var (
//	iface   string = "lo" //config.HostName+"-eth0"
//	snaplen int32  = 65535
//	promisc bool   = true
//	err     error
//	timeout time.Duration = -1 * time.Second
//	handle  *pcap.Handle
//)

type Iface struct{
	name string
	snaplen int32
	promisc bool
	timeout time.Duration
	handle *pcap.Handle
}

func NewIface(name string)*Iface{
	snaplen := int32(65535)
	promisc := true
	timeout := -1 * time.Second
	handle, err := pcap.OpenLive(name, snaplen, promisc, timeout)
	if err != nil{
		log.Println(err)
	}
	i :=  &Iface{
		name:      	name,
		snaplen: 	snaplen,
		promisc:	promisc,
		timeout: 	timeout,
		handle: 	handle,			
	}
	return i
}

func (i *Iface) Close(){
	i.handle.Close()
}

func (i *Iface) Process_Packets(other *Iface){
	
	log.Println("Interface Ok")
	packetSource := gopacket.NewPacketSource(i.handle, i.handle.LinkType())
	defer i.Close()
	for packet := range packetSource.Packets() {
		log.Println(i.name+" received packet --> "+other.name)
		log.Println(packet)
		other.Send_Packet(packet.Data())
	}
		
}

func (i *Iface) Send_Packet(pkt []byte){
	i.handle.WritePacketData(pkt)
}

type Forwarder struct{
	fetcher fetcher.Fetcher
	site_conns map[string]*net.UDPConn
}

func NewForwarder(fetcher *fetcher.Fetcher)*Forwarder{
	var site_conns map[string]*net.UDPConn
	var conn *net.UDPConn
	site_conns = make(map[string]*net.UDPConn)
	sites := fetcher.GetSites()
	dest_port := "1234" // TODO: put in some config file
	for _, site := range sites{
		if string(fetcher.LocalAddr) != string(site.TPAddr){
			udpAddr, err := net.ResolveUDPAddr("udp4", site.TPAddr+":"+dest_port)
			if err != nil{
				fmt.Println("ERROR: Unable to resolve UDP Address to remote site")
			}
			conn, err = net.DialUDP("udp", nil, udpAddr)
			site_conns[site.TPAddr]=conn	
		}	
	}
	f :=  &Forwarder{
		fetcher:      *fetcher,
		site_conns: site_conns,
			
	}

	return f
}

func (f *Forwarder) Close_conns(){
	for _, conn := range f.site_conns{
		conn.Close()
		fmt.Println("Conn closed")
	} 
}

//func Test() {
//	handle, err = pcap.OpenLive(iface, snaplen, promisc, timeout)
//	if err != nil {
//		fmt.Println("Ooops")
//		fmt.Println(err)
//	}
//	defer handle.Close()
//	//packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
//	//for packet := range packetSource.Packets() {
//	//	handle_packet(packet)
//	//	//break
//	//}
//	var data []byte = make([]byte, 4)
//	for{
//		handle.WritePacketData(data)
//	}
//	
//}

func handle_packet(packet gopacket.Packet) {
	fmt.Println(packet)

}
