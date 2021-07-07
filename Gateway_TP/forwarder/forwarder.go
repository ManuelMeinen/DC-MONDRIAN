package forwarder

import (
	"encoding/binary"
	"errors"
	"gateway_tp/config"
	"gateway_tp/crypto"
	"gateway_tp/fetcher"
	"gateway_tp/keyman"
	"gateway_tp/mondrian"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"log"
	"net"
	"time"
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
		//log.Println(logPrefix+"IP layer detected")
		ipv4, _ := ipLayer.(*layers.IPv4)
		src_ip :=ipv4.SrcIP
		dest_ip := ipv4.DstIP
		zone, remoteTP, err := fwd.Fetcher.GetZoneAndSite(dest_ip)
		if err!=nil{
			log.Println(logPrefix+"ERROR: Finding Zone and Site failed")
			log.Println(err)
			return "", "", 0, nil, err
		}
		_, localTP, err := fwd.Fetcher.GetZoneAndSite(src_ip)
		if err!=nil{
			log.Println(logPrefix+"ERROR: Finding Zone and Site failed")
			log.Println(err)
			return "", "", 0, nil, err
		}
		key, err := fwd.Km.GetKey(localTP, remoteTP, uint32(zone))
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
		//log.Println(logPrefix+"IP layer detected")
		ipv4, _ := ipLayer.(*layers.IPv4)
		src_ip :=ipv4.SrcIP
		dest_ip := ipv4.DstIP
		remoteTP := src_ip.String()
		localTP := dest_ip.String()
		var zone uint32
		if mLayer := pkt.Layer(mondrian.MondrianLayerType); mLayer != nil {
			//log.Println(logPrefix+"Mondrian layer detected")
			mondrian := mLayer.(*mondrian.MondrianLayer)
			zone = mondrian.ZoneID
		}else{
			return "", "", 0, nil, errors.New("No Mondrian Layer decoded")
		}
		key, err := fwd.Km.GetKey(remoteTP, localTP, zone)
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
			//log.Println(logPrefix+i.name + " received packet --> " + other.name)
			//log.Println(packet)
			//outPacket := fwd.ID(packet)
			//TODO combine these two functions to save on packet creation
			outPacket := fwd.ToMondrian(packet)
			go other.Send_Packet(outPacket.Data())
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
			//log.Println(logPrefix+i.name + " received packet --> " + other.name)
			//log.Println(packet)
			//outPacket := fwd.ID(packet)
			//TODO combine these two functions to save on packet creation
			outPacket := fwd.FromMondrian(packet)
			go other.Send_Packet(outPacket.Data())
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

func (fwd *Forwarder)ToMondrianAndSend(pkt gopacket.Packet, outIface *Iface){
	/*
	Convert an IPv4 packet into a Mondrian packet
	Note: If it's not an IPv4 Packet then just return pkt (like this ARP still works)
	*/
	

	// Get eth header to restore it afterwards
	var eth layers.Ethernet
	var ip_internal layers.IPv4
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip_internal)
	decoded := []gopacket.LayerType{}
	parser.DecodeLayers(pkt.Data(), &decoded)
	
	if len(decoded) < 2{
		//log.Println(logPrefix+"No IPv4 header detected")
		go outIface.Send_Packet(pkt.Data())
	}
	zone, remoteTP, err := fwd.Fetcher.GetZoneAndSite(ip_internal.DstIP)
	if err!=nil{
		log.Println(logPrefix+"ERROR: Finding Zone and Site failed")
		log.Println(err)
		go outIface.Send_Packet(pkt.Data())
	}
	_, localTP, err := fwd.Fetcher.GetZoneAndSite(ip_internal.SrcIP)
	if err!=nil{
		log.Println(logPrefix+"ERROR: Finding Zone and Site failed")
		log.Println(err)
		go outIface.Send_Packet(pkt.Data())
	}
	key, err := fwd.Km.GetKey(localTP, remoteTP, uint32(zone))
	if err!=nil{
		log.Println(logPrefix+"ERROR: Getting the key failed")
		go outIface.Send_Packet(pkt.Data())
	}
	srcTP_ip, _, _ := net.ParseCIDR(localTP+"/8")
	destTP_ip, _, _ := net.ParseCIDR(remoteTP+"/8")
	// Create outer IPv4 header
	ip := layers.IPv4{
		SrcIP:    srcTP_ip,
		DstIP:    destTP_ip,
		Version:  4,
		IHL:      5,
		Protocol: 112, // VRRP
	}
	//Test out inlining Encryption
	var ad []byte = make([]byte, 20)
	ad[0] = 1
	dummy := make([]byte, 4)
	binary.LittleEndian.PutUint32(dummy, uint32(zone))
	copy(ad[1:4], dummy)
	timestamp:=time.Now()
	binary.LittleEndian.PutUint32(ad[4:8], uint32(timestamp.Unix()))
	nonce := crypto.Nonce()
	copy(ad[8:20], nonce)

	aead, err := crypto.NewAEAD(key[:])
	if err != nil {
		log.Println(logPrefix+err.Error())
		go outIface.Send_Packet(pkt.Data())
	}
	data_len := len(pkt.Data())
	res := aead.Seal(nil, nonce, pkt.Data(), ad)
	
	m := mondrian.MondrianLayer{
		Type:      1,
		ZoneID:    uint32(zone),
		TimeStamp: timestamp,
		Nonce:     nonce,
	}
	m.Payload = res[:data_len]
	m.MAC = res[data_len:]

	// Create final (encrypted) Mondrian out-packet
	pld := gopacket.Payload(m.LayerPayload())
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}
	err = gopacket.SerializeLayers(buf, opts, &eth, &ip, &m, &pld)
	//out_pkt := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
	go outIface.Send_Packet(buf.Bytes())
}

func (fwd *Forwarder)FromMondrianAndSend(pkt gopacket.Packet, outIface *Iface){
	var eth layers.Ethernet
	var ip layers.IPv4
	var m mondrian.MondrianLayer
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip, &m)
	decoded := []gopacket.LayerType{}
	parser.DecodeLayers(pkt.Data(), &decoded)
	
	if len(decoded) < 3{
		//log.Println(logPrefix+"No Mondrian header detected")
		go outIface.Send_Packet(pkt.Data())
	}

	src_ip := ip.SrcIP
	dest_ip := ip.DstIP
	remoteTP := src_ip.String()
	localTP := dest_ip.String()
	zone := m.ZoneID

	key, err := fwd.Km.GetKey(remoteTP, localTP, zone)
	if err!=nil{
		log.Println(logPrefix+err.Error())
		go outIface.Send_Packet(pkt.Data())
	}

	aead, err := crypto.NewAEAD(key[:])
	if err != nil {
		log.Println(logPrefix+err.Error())
		go outIface.Send_Packet(pkt.Data())
	}
	buf := append(m.Payload, m.MAC...)
	m.Payload, err = aead.Open(nil, m.Nonce, buf, m.Contents[:20])
	if err != nil {
		log.Println(logPrefix+err.Error())
		go outIface.Send_Packet(pkt.Data())
	}
	
	// Return the Mondrian packet's payload as IPv4 packet
	//out_pkt := gopacket.NewPacket(m.LayerPayload(), layers.LayerTypeEthernet, gopacket.Default)
	go outIface.Send_Packet(m.LayerPayload())
}

func (fwd *Forwarder)ToMondrian(pkt gopacket.Packet) gopacket.Packet {
	/*
	Convert an IPv4 packet into a Mondrian packet
	Note: If it's not an IPv4 Packet then just return pkt (like this ARP still works)
	*/
	

	// Get eth header to restore it afterwards
	var eth layers.Ethernet
	var ip_internal layers.IPv4
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip_internal)
	decoded := []gopacket.LayerType{}
	parser.DecodeLayers(pkt.Data(), &decoded)
	
	if len(decoded) < 2{
		//log.Println(logPrefix+"No IPv4 header detected")
		return pkt
	}
	zone, remoteTP, err := fwd.Fetcher.GetZoneAndSite(ip_internal.DstIP)
	if err!=nil{
		log.Println(logPrefix+"ERROR: Finding Zone and Site failed")
		log.Println(err)
		return pkt
	}
	_, localTP, err := fwd.Fetcher.GetZoneAndSite(ip_internal.SrcIP)
	if err!=nil{
		log.Println(logPrefix+"ERROR: Finding Zone and Site failed")
		log.Println(err)
		return pkt
	}
	key, err := fwd.Km.GetKey(localTP, remoteTP, uint32(zone))
	if err!=nil{
		log.Println(logPrefix+"ERROR: Getting the key failed")
		return pkt
	}
	srcTP_ip, _, _ := net.ParseCIDR(localTP+"/8")
	destTP_ip, _, _ := net.ParseCIDR(remoteTP+"/8")
	// Create outer IPv4 header
	ip := layers.IPv4{
		SrcIP:    srcTP_ip,
		DstIP:    destTP_ip,
		Version:  4,
		IHL:      5,
		Protocol: 112, // VRRP
	}
	//Test out inlining Encryption
	var ad []byte = make([]byte, 20)
	ad[0] = 1
	dummy := make([]byte, 4)
	binary.LittleEndian.PutUint32(dummy, uint32(zone))
	copy(ad[1:4], dummy)
	timestamp:=time.Now()
	binary.LittleEndian.PutUint32(ad[4:8], uint32(timestamp.Unix()))
	nonce := crypto.Nonce()
	copy(ad[8:20], nonce)

	aead, err := crypto.NewAEAD(key[:])
	if err != nil {
		log.Println(logPrefix+err.Error())
		return pkt
	}
	data_len := len(pkt.Data())
	res := aead.Seal(nil, nonce, pkt.Data(), ad)
	
	m := mondrian.MondrianLayer{
		Type:      1,
		ZoneID:    uint32(zone),
		TimeStamp: timestamp,
		Nonce:     nonce,
	}
	m.Payload = res[:data_len]
	m.MAC = res[data_len:]

	// Create final (encrypted) Mondrian out-packet
	pld := gopacket.Payload(m.LayerPayload())
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}
	err = gopacket.SerializeLayers(buf, opts, &eth, &ip, &m, &pld)
	out_pkt := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
	//End test



	//// Create Mondrian header
	//m := mondrian.MondrianLayer{
	//	Type:      1,
	//	ZoneID:    uint32(zone),
	//	TimeStamp: time.Now(),
	//	Nonce:     crypto.Nonce(),
	//}
	//// Payload = original IPv4 Packet (for now unencrypted)
	//pld := gopacket.Payload(pkt.Data())
	//buf := gopacket.NewSerializeBuffer()
	//opts := gopacket.SerializeOptions{}
	//// Create intermediate packet stored in buf
	//err = gopacket.SerializeLayers(buf, opts, &eth, &ip, &m, &pld)
	//if err != nil {
	//	log.Println(logPrefix+err.Error())
	//}
	//tmp_pkt := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
	//// Get Mondrian Layer from tmp_pkt
	//mlayer := tmp_pkt.Layer(mondrian.MondrianLayerType).(*mondrian.MondrianLayer)
	//// Transfor Mondrian Layer into encrypted/authenticated state
	//err = mlayer.Encrypt(key[:])
	//if err != nil {
	//	log.Println(logPrefix+err.Error())
	//	return pkt
	//}
	//// Create final (encrypted) Mondrian out-packet
	//pld = gopacket.Payload(mlayer.LayerPayload())
	//buf = gopacket.NewSerializeBuffer()
	//opts = gopacket.SerializeOptions{}
	//err = gopacket.SerializeLayers(buf, opts, &eth, &ip, mlayer, &pld)
	//out_pkt := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
	// Return Mondrian packet
	return out_pkt
}


func (fwd *Forwarder)ToMondrianOld(pkt gopacket.Packet) gopacket.Packet {
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
	var eth layers.Ethernet
	var ip layers.IPv4
	var m mondrian.MondrianLayer
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip, &m)
	decoded := []gopacket.LayerType{}
	parser.DecodeLayers(pkt.Data(), &decoded)
	
	if len(decoded) < 3{
		//log.Println(logPrefix+"No Mondrian header detected")
		return pkt
	}

	src_ip := ip.SrcIP
	dest_ip := ip.DstIP
	remoteTP := src_ip.String()
	localTP := dest_ip.String()
	zone := m.ZoneID

	key, err := fwd.Km.GetKey(remoteTP, localTP, zone)
	if err!=nil{
		log.Println(logPrefix+err.Error())
		return pkt
	}

	aead, err := crypto.NewAEAD(key[:])
	if err != nil {
		log.Println(logPrefix+err.Error())
		return pkt
	}
	buf := append(m.Payload, m.MAC...)
	m.Payload, err = aead.Open(nil, m.Nonce, buf, m.Contents[:20])
	if err != nil {
		log.Println(logPrefix+err.Error())
		return pkt
	}
	
	//// Get Mondrian layer
	//mlayer := pkt.Layer(mondrian.MondrianLayerType).(*mondrian.MondrianLayer)
	//// Decrypt/verify the packet
	//err = mlayer.Decrypt(key[:])
	//if err != nil {
	//	log.Println(logPrefix+err.Error())
	//	return pkt
	//}
	// Return the Mondrian packet's payload as IPv4 packet
	out_pkt := gopacket.NewPacket(m.LayerPayload(), layers.LayerTypeEthernet, gopacket.Default)
	return out_pkt
}

func (fwd *Forwarder)FromMondrianOld(pkt gopacket.Packet) gopacket.Packet {
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
	Fetcher            *fetcher.Fetcher
	Internal_interface *Iface
	External_interface *Iface
	Km				   *keyman.KeyMan
}

func NewForwarder() *Forwarder {
	f := fetcher.NewFetcher(config.TPAddr, 60*60) //Refresh Interval = 1h (in reality this will be smaller)
	int_iface := NewIface(config.HostName + "-eth0")
	ext_iface := NewIface(config.HostName + "-eth1")
	km := keyman.NewKeyMan(config.MasterSecret, false)
	fwd := &Forwarder{
		Fetcher:            f,
		Internal_interface: int_iface,
		External_interface: ext_iface,
		Km:					km,
	}
	return fwd
}

func (f *Forwarder) Start() {
	/*
	Start gorutines processing ingress and egress traffic
	*/
	log.Println(logPrefix+"Forwarder started")
	go f.Internal_interface.Process_Egress_Traffic(f.External_interface, f)
	go f.External_interface.Process_Ingress_Traffic(f.Internal_interface, f)
}

func (f *Forwarder) Stop() {
	/*
	Close the interfaces
	*/
	log.Println(logPrefix+"Forwarder stopped")
	f.Internal_interface.Close()
	f.External_interface.Close()
}
