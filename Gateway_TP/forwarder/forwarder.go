package forwarder

import (
	"gateway_tp/config"
	"gateway_tp/fetcher"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"

	"log"
	"time"
)

type Iface struct {
	name    string
	snaplen int32
	promisc bool
	timeout time.Duration
	handle  *pcap.Handle
}

func NewIface(name string) *Iface {
	snaplen := int32(65535)
	promisc := true
	timeout := -1 * time.Second
	handle, err := pcap.OpenLive(name, snaplen, promisc, timeout)
	if err != nil {
		log.Println(err)
	}
	i := &Iface{
		name:    name,
		snaplen: snaplen,
		promisc: promisc,
		timeout: timeout,
		handle:  handle,
	}
	return i
}

func (i *Iface) Close() {
	i.handle.Close()
}

func (i *Iface) Process_Egress_Traffic(other *Iface) {

	log.Println("[Forwarder] Started processing Egress Traffic")
	packetSource := gopacket.NewPacketSource(i.handle, i.handle.LinkType())
	defer i.Close()
	for packet := range packetSource.Packets() {
		log.Println(i.name + " received packet --> " + other.name)
		log.Println(packet)
		outPacket := toMonfiran(packet)
		other.Send_Packet(outPacket.Data())
	}

}

func (i *Iface) Process_Ingress_Traffic(other *Iface) {

	log.Println("[Forwarder] Started processing Ingress Traffic")
	packetSource := gopacket.NewPacketSource(i.handle, i.handle.LinkType())
	defer i.Close()
	for packet := range packetSource.Packets() {
		log.Println(i.name + " received packet --> " + other.name)
		log.Println(packet)
		outPacket := fromMonfiran(packet)
		other.Send_Packet(outPacket.Data())
	}
}

func toMonfiran(pkt gopacket.Packet) gopacket.Packet {
	//TODO: Transform pkt into a Mondrian packet and return it
	return pkt
}

func fromMonfiran(pkt gopacket.Packet) gopacket.Packet {
	//TODO: Transform pkt from a Mondrian packet into an IP packet and return it
	return pkt
}

func (i *Iface) Send_Packet(pkt []byte) {
	i.handle.WritePacketData(pkt)
}

type Forwarder struct {
	fetcher            *fetcher.Fetcher
	internal_interface *Iface
	external_interface *Iface
}

func NewForwarder() *Forwarder {
	f := fetcher.NewFetcher(config.TPAddr, 10)
	int_iface := NewIface(config.HostName + "-eth0")
	ext_iface := NewIface(config.HostName + "-eth1")
	fwd := &Forwarder{
		fetcher:            f,
		internal_interface: int_iface,
		external_interface: ext_iface,
	}
	return fwd
}

func (f *Forwarder) Start() {
	log.Println("[Forwarder] Forwarder started")
	go f.internal_interface.Process_Egress_Traffic(f.external_interface)
	go f.external_interface.Process_Ingress_Traffic(f.internal_interface)
}

func (f *Forwarder) Stop() {
	log.Println("[Forwarder] Forwarder stopped")
	f.internal_interface.Close()
	f.external_interface.Close()
}
