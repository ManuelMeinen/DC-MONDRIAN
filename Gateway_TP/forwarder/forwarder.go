package forwarder

import (
	"errors"
	"gateway_tp/config"
	"gateway_tp/fetcher"
	"gateway_tp/keyman"
	"gateway_tp/mondrian"
	"gateway_tp/crypto"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"log"
	"time"
	"net"
)

var logPrefix = "[Forwarder] "

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

func (fwd *Forwarder)GetMondrianInfoEgress(pkt gopacket.Packet)(string, string, uint, []byte, error){
	/*
	Fetch the info used to transform an IPv4 packet into a Mondrian packet
	return localTP, remoteTP, zone, key, err
	*/
	if ipLayer := pkt.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		log.Println(logPrefix+"IP layer detected")
		ipv4, _ := ipLayer.(*layers.IPv4)
		src_ip :=ipv4.SrcIP
		dest_ip := ipv4.DstIP
		zone, remoteTP, err := fwd.fetcher.GetZoneAndSite(dest_ip)
		if err!=nil{
			log.Println(logPrefix+"ERROR: Finding Zone and Site failed")
			log.Println(err)
			return "", "", 0, nil, err
		}
		_, localTP, err := fwd.fetcher.GetZoneAndSite(src_ip)
		if err!=nil{
			log.Println(logPrefix+"ERROR: Finding Zone and Site failed")
			log.Println(err)
			return "", "", 0, nil, err
		}
		key, err := fwd.km.GetKey(localTP, remoteTP, uint32(zone))
		if err!=nil{
			log.Println(logPrefix+"ERROR: Getting the key failed")
			return "", "", 0, nil, err
		}
		return localTP, remoteTP, zone, key, nil
	}else{
		return "", "", 0, nil, errors.New("No IPv4 Layer decoded")
	}
}

func (fwd *Forwarder)GetMondrianInfoIngress(pkt gopacket.Packet)(string, string, uint, []byte, error){
	/*
	Parse and fetch the info used to transform a Mondrian packet into an IPv4 packet
	return localTP, remoteTP, zone, key, err
	*/
	if ipLayer := pkt.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		log.Println(logPrefix+"IP layer detected")
		ipv4, _ := ipLayer.(*layers.IPv4)
		src_ip :=ipv4.SrcIP
		dest_ip := ipv4.DstIP
		remoteTP := src_ip.String()
		localTP := dest_ip.String()
		var zone uint32
		if mLayer := pkt.Layer(mondrian.MondrianLayerType); mLayer != nil {
			log.Println(logPrefix+"Mondrian layer detected")
			mondrian := mLayer.(*mondrian.MondrianLayer)
			zone = mondrian.ZoneID
		}else{
			return "", "", 0, nil, errors.New("No Mondrian Layer decoded")
		}
		key, err := fwd.km.GetKey(remoteTP, localTP, zone)
		if err!=nil{
			log.Println(logPrefix+"ERROR: Getting the key failed")
			return "", "", 0, nil, err
		}
		return localTP, remoteTP, uint(zone), key, nil
	}else{
		return "", "", 0, nil, errors.New("No IPv4 Layer decoded")
	}
}

func (i *Iface) Process_Egress_Traffic(other *Iface, fwd *Forwarder) {
	/*
	Process packets from inside the DC (ariving on interface i) 
	going into the internet (via interface other)
	*/
	log.Println(logPrefix+"Started processing Egress Traffic")
	packetSource := gopacket.NewPacketSource(i.handle, i.handle.LinkType())
	defer i.Close()
	for packet := range packetSource.Packets() {
		//go func() {
			log.Println(logPrefix+i.name + " received packet --> " + other.name)
			log.Println(packet)
			outPacket := fwd.ToMondrian(packet)
			other.Send_Packet(outPacket.Data())
		//}()	
	}

}

func (i *Iface) Process_Ingress_Traffic(other *Iface, fwd *Forwarder) {
	/*
	Process packets from the internet (ariving on interface i) 
	going into the DC (via interface other)
	*/
	log.Println(logPrefix+"Started processing Ingress Traffic")
	packetSource := gopacket.NewPacketSource(i.handle, i.handle.LinkType())
	defer i.Close()
	for packet := range packetSource.Packets() {
		//go func(){
			log.Println(logPrefix+i.name + " received packet --> " + other.name)
			log.Println(packet)
			outPacket := fwd.FromMondrian(packet)
			other.Send_Packet(outPacket.Data())
		//}()
	}
}

func (fwd *Forwarder) ID(pkt gopacket.Packet) gopacket.Packet{
	/*
	Just return the packet that was passed in 
	(Used to compare the overhead of transforming packets)
	*/
	return pkt
}


func (fwd *Forwarder)ToMondrian(pkt gopacket.Packet) gopacket.Packet {
	/*
	Convert an IPv4 packet into a Mondrian packet
	Note: If it's not an IPv4 Packet then just return pkt (like this ARP still works)
	*/
	srcTP, destTP, zone, key, err := fwd.GetMondrianInfoEgress(pkt)
	if err!=nil{
		log.Println(logPrefix+err.Error())
		return pkt
	}
	srcTP_ip, _, _ := net.ParseCIDR(srcTP+"/8")
	destTP_ip, _, _ := net.ParseCIDR(destTP+"/8")

	// Get eth header to restore it afterwards
	var eth layers.Ethernet
	eth_parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth)
	eth_decoded := []gopacket.LayerType{}
	eth_parser.DecodeLayers(pkt.Data(), &eth_decoded)
	// Create outer IPv4 header
	ip := layers.IPv4{
		SrcIP:    srcTP_ip,
		DstIP:    destTP_ip,
		Version:  4,
		IHL:      5,
		Protocol: 112, // VRRP
	}
	// Create Mondrian header
	m := mondrian.MondrianLayer{
		Type:      1,
		ZoneID:    uint32(zone),
		TimeStamp: time.Now(),
		Nonce:     crypto.Nonce(),
	}
	// Payload = original IPv4 Packet (for now unencrypted)
	pld := gopacket.Payload(pkt.Data())
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}
	// Create intermediate packet stored in buf
	err = gopacket.SerializeLayers(buf, opts, &eth, &ip, &m, &pld)
	if err != nil {
		log.Println(logPrefix+err.Error())
	}
	tmp_pkt := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
	// Get Mondrian Layer from tmp_pkt
	mlayer := tmp_pkt.Layer(mondrian.MondrianLayerType).(*mondrian.MondrianLayer)
	// Transfor Mondrian Layer into encrypted/authenticated state
	err = mlayer.Encrypt(key[:])
	if err != nil {
		log.Println(logPrefix+err.Error())
		return pkt
	}
	// Create final (encrypted) Mondrian out-packet
	pld = gopacket.Payload(mlayer.LayerPayload())
	buf = gopacket.NewSerializeBuffer()
	opts = gopacket.SerializeOptions{}
	err = gopacket.SerializeLayers(buf, opts, &eth, &ip, mlayer, &pld)
	out_pkt := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
	// Return Mondrian packet
	return out_pkt
}

func (fwd *Forwarder)FromMondrian(pkt gopacket.Packet) gopacket.Packet {
	/*
	Convert a Mondrian packet into an IPv4 packet
	Note: If it's not a Mondrian Packet then just return pkt (like this ARP still works)
	*/
	// Get the key for decryption
	_, _, _, key, err := fwd.GetMondrianInfoIngress(pkt)
	if err!=nil{
		log.Println(logPrefix+err.Error())
		return pkt
	}
	// Get Mondrian layer
	mlayer := pkt.Layer(mondrian.MondrianLayerType).(*mondrian.MondrianLayer)
	// Decrypt/verify the packet
	err = mlayer.Decrypt(key[:])
	if err != nil {
		log.Println(logPrefix+err.Error())
		return pkt
	}
	// Return the Mondrian packet's payload as IPv4 packet
	out_pkt := gopacket.NewPacket(mlayer.LayerPayload(), layers.LayerTypeEthernet, gopacket.Default)
	return out_pkt
}

func (i *Iface) Send_Packet(pkt []byte) {
	/*
	Send a packet on a wire via interface i
	*/
	i.handle.WritePacketData(pkt)
}

type Forwarder struct {
	fetcher            *fetcher.Fetcher
	internal_interface *Iface
	external_interface *Iface
	km				   *keyman.KeyMan
}

func NewForwarder() *Forwarder {
	f := fetcher.NewFetcher(config.TPAddr, 60*60) //Refresh Interval = 1h (in reality this will be smaller)
	int_iface := NewIface(config.HostName + "-eth0")
	ext_iface := NewIface(config.HostName + "-eth1")
	km := keyman.NewKeyMan(config.MasterSecret)
	fwd := &Forwarder{
		fetcher:            f,
		internal_interface: int_iface,
		external_interface: ext_iface,
		km:					km,
	}
	return fwd
}

func (f *Forwarder) Start() {
	/*
	Start gorutines processing ingress and egress traffic
	*/
	log.Println(logPrefix+"Forwarder started")
	go f.internal_interface.Process_Egress_Traffic(f.external_interface, f)
	go f.external_interface.Process_Ingress_Traffic(f.internal_interface, f)
}

func (f *Forwarder) Stop() {
	/*
	Close the interfaces
	*/
	log.Println(logPrefix+"Forwarder stopped")
	f.internal_interface.Close()
	f.external_interface.Close()
}
