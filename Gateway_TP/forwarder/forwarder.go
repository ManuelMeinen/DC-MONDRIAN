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

func (fwd *Forwarder)getMondrianInfo(pkt gopacket.Packet)(string, string, uint, []byte, error){
	if ipLayer := pkt.Layer(layers.LayerTypeIPv4); ipLayer != nil {
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

func (i *Iface) Process_Egress_Traffic(other *Iface, fwd *Forwarder) {

	log.Println(logPrefix+"Started processing Egress Traffic")
	packetSource := gopacket.NewPacketSource(i.handle, i.handle.LinkType())
	defer i.Close()
	//var ip4 layers.IPv4
	//egressParser := gopacket.NewDecodingLayerParser(layers.LayerTypeIPv4, &ip4)
	//decodedLayers := []gopacket.LayerType{}
	for packet := range packetSource.Packets() {
		log.Println(i.name + " received packet --> " + other.name)
		log.Println(packet)
		outPacket := fwd.toMondiran(packet)
		other.Send_Packet(outPacket.Data())
	}

}

func (i *Iface) Process_Ingress_Traffic(other *Iface, fwd *Forwarder) {

	log.Println(logPrefix+"Started processing Ingress Traffic")
	packetSource := gopacket.NewPacketSource(i.handle, i.handle.LinkType())
	defer i.Close()
	for packet := range packetSource.Packets() {
		log.Println(i.name + " received packet --> " + other.name)
		log.Println(packet)
		outPacket := fwd.fromMondiran(packet)
		other.Send_Packet(outPacket.Data())
	}
}



func (fwd *Forwarder)toMondiran(pkt gopacket.Packet) gopacket.Packet {
	//TODO: Transform pkt into a Mondrian packet and return it
	srcTP, destTP, zone, key, err := fwd.getMondrianInfo(pkt)
	if err!=nil{
		log.Println(err)
		return pkt
	}
	//log.Println(srcTP)
	//log.Println(destTP)
	//log.Println(zone)
	//log.Println(key)
	srcTP_ip, _, _ := net.ParseCIDR(srcTP+"/8")
	destTP_ip, _, _ := net.ParseCIDR(destTP+"/8")
	// Get eth header to restore it afterwards
	var eth layers.Ethernet
	eth_parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth)
	eth_decoded := []gopacket.LayerType{}
	eth_parser.DecodeLayers(pkt.Data(), &eth_decoded)
	//log.Println(eth)

	ip := layers.IPv4{
		SrcIP:    srcTP_ip,
		DstIP:    destTP_ip,
		Version:  4,
		Length:   103,
		IHL:      5,
		Protocol: 112, // VRRP
	}

	m := mondrian.MondrianLayer{
		Type:      1,
		ZoneID:    uint32(zone),
		TimeStamp: time.Now(),
		Nonce:     crypto.Nonce(),
	}

	pld := gopacket.Payload(pkt.Data())
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{} // See SerializeOptions for more details.
	err = gopacket.SerializeLayers(buf, opts, &eth, &ip, &m, &pld)

	if err != nil {
		log.Println(err)
	}
	var eth_new layers.Ethernet
	var ip_new layers.IPv4
	var m_new mondrian.MondrianLayer
	var pld_new gopacket.Payload

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth_new, &ip_new, &m_new, &pld_new)	
	decoded := []gopacket.LayerType{}
	err = parser.DecodeLayers(buf.Bytes(), &decoded)

	if err != nil {
		log.Println(err)

	}

	out_pkt := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
	//log.Println("############################")
	//log.Println(out_pkt)

	
	mlayer := out_pkt.Layer(mondrian.MondrianLayerType).(*mondrian.MondrianLayer)
	//log.Println(mlayer)
	log.Println("Unencrypted Payload:")
	log.Println(mlayer.LayerPayload())
	err = mlayer.Encrypt(key[:])
	if err != nil {
		log.Println(err)
		return pkt
	}
	log.Println("Encrypted Payload:")
	log.Println(mlayer.LayerPayload())
	//out_pkt = gopacket.NewPacket(mlayer.LayerPayload(), layers.LayerTypeEthernet, gopacket.Default)
	log.Println("Mondiran Packet:")
	log.Println(out_pkt)

	//log.Println(ip)
	//log.Println(m)

	//TODO: When the other side is implemented then change with return out_pkt
	return pkt
}

func (fwd *Forwarder)fromMondiran(pkt gopacket.Packet) gopacket.Packet {
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
	km				   *keyman.KeyMan
}

func NewForwarder() *Forwarder {
	f := fetcher.NewFetcher(config.TPAddr, 10)
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
	log.Println(logPrefix+"Forwarder started")
	go f.internal_interface.Process_Egress_Traffic(f.external_interface, f)
	go f.external_interface.Process_Ingress_Traffic(f.internal_interface, f)
}

func (f *Forwarder) Stop() {
	log.Println(logPrefix+"Forwarder stopped")
	f.internal_interface.Close()
	f.external_interface.Close()
}
