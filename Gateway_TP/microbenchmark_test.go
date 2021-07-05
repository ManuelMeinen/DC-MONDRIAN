package main

// go test -v -bench=.
import (
	"fmt"
	"gateway_tp/config"
	"gateway_tp/crypto"
	"gateway_tp/forwarder"

	"gateway_tp/fetcher"
	"gateway_tp/keyman"
	"gateway_tp/logger"
	"gateway_tp/mondrian"

	"gateway_tp/api"
	"gateway_tp/types"
	"log"
	"net"
	"sort"
	"testing"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var initialized = false
var fwd *forwarder.Forwarder
var int_iface *forwarder.Iface
var ext_iface *forwarder.Iface

func Init() {
	if initialized == false {
		config.Init("config/config.json")
		logger.InitLogger()
		
		f := &fetcher.Fetcher{
			LocalAddr:        config.TPAddr,
			ControllerAddr:   config.ControllerAddr,
			ControllerPort:   config.ControllerPort,
			Refresh_interval: 100 * 60,
			Conn:             *api.GetConn(config.ControllerAddr, config.ControllerPort),
		}
		var subnets types.Subnets
		var sites types.Sites

		subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("10.0.1.0"), Mask: net.IPv4Mask(255, 255, 255, 0)}, 1, "200.0.0.1"})
		subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("10.1.0.0"), Mask: net.IPv4Mask(255, 255, 0, 0)}, 1, "200.0.0.1"})
		subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("10.2.0.0"), Mask: net.IPv4Mask(255, 255, 0, 0)}, 1, "200.0.0.1"})

		subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("20.0.1.0"), Mask: net.IPv4Mask(255, 255, 255, 0)}, 1, "200.0.0.2"})
		subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("20.2.0.0"), Mask: net.IPv4Mask(255, 255, 0, 0)}, 2, "200.0.0.2"})

		subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("30.0.1.0"), Mask: net.IPv4Mask(255, 255, 255, 0)}, 1, "200.0.0.3"})
		subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("30.1.0.0"), Mask: net.IPv4Mask(255, 255, 0, 0)}, 2, "200.0.0.3"})
		subnets = append(subnets, &types.Subnet{net.IPNet{IP: net.ParseIP("30.2.0.0"), Mask: net.IPv4Mask(255, 255, 0, 0)}, 3, "200.0.0.3"})

		sites = append(sites, &types.Site{"200.0.0.1", "Site 1"})
		sites = append(sites, &types.Site{"200.0.0.2", "Site 2"})
		sites = append(sites, &types.Site{"200.0.0.3", "Site 3"})

		f.SetSubnets(subnets)
		f.SetSites(sites)
		int_iface_name := "lo"
		ext_iface_name := "lo"
		int_iface = forwarder.NewIface(int_iface_name)
		ext_iface = forwarder.NewIface(ext_iface_name)
		km := keyman.NewKeyMan(config.MasterSecret, true)
        egressL1Keys := make(map[string]keyman.KeyPld)
        L1key, err := km.DeriveL1Key("200.0.0.1") //Note: This only works since both have the same master secret
        if err!=nil{
            fmt.Println(err)
        }
        egressL1Keys["200.0.0.2"] = *L1key
		km.SetEgressL1Keys(egressL1Keys)
		fwd = &forwarder.Forwarder{
			Fetcher:            f,
			Internal_interface: int_iface,
			External_interface: ext_iface,
			Km:                 km,
		}
		fmt.Println("forwarder ready")

		// wait such that fetcher is ready
		time.Sleep(10 * time.Second)
	}
	initialized = true
}

func BenchmarkFromMondrian(b *testing.B) {
	var mondrianPackets = make(map[int]gopacket.Packet)

	var (
		pcapFile string = "/Gateway_TP/pcap_files/ingress.pcap"
		handle   *pcap.Handle
		err      error
	)
	Init()

	// Loop through packets in file
	handle, err = pcap.OpenOffline(config.BASE_PATH + pcapFile)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	defer handle.Close()
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
			if mLayer := packet.Layer(mondrian.MondrianLayerType); mLayer != nil {
				mondrianPackets[len(packet.Data())] = packet
			}
		}
	}
	sizes := []int{}
	for key := range mondrianPackets {
		sizes = append(sizes, key)
	}
	sort.Ints(sizes)

	for _, size := range sizes {
		b.Run(fmt.Sprintf("packet size %d", size), func(b *testing.B) {
			packet := mondrianPackets[size]
			//Warmup
			fwd.FromMondrian(packet)
			b.SetBytes(int64(size))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				fwd.FromMondrian(packet)
			}
		})
	}
}

func BenchmarkToMondrian(b *testing.B) {
	var ipPackets = make(map[int]gopacket.Packet)

	var (
		pcapFile string = "/Gateway_TP/pcap_files/egress.pcap"
		handle   *pcap.Handle
		err      error
	)
	Init()
	// Loop through packets in file
	handle, err = pcap.OpenOffline(config.BASE_PATH + pcapFile)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	defer handle.Close()
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
			ipPackets[len(packet.Data())] = packet
		}
	}

	sizes := []int{}
	for key := range ipPackets {
		sizes = append(sizes, key)
	}
	sort.Ints(sizes)

	for _, size := range sizes {
		packet := ipPackets[size]
		//Warmup
		fwd.ToMondrian(packet)
		b.Run(fmt.Sprintf("packet size %d", size), func(b *testing.B) {
			b.SetBytes(int64(size))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				fwd.ToMondrian(packet)
			}
		})
	}
}

func BenchmarkID(b *testing.B) {
	var ipPackets = make(map[int]gopacket.Packet)

	var (
		pcapFile string = "/Gateway_TP/pcap_files/egress.pcap"
		handle   *pcap.Handle
		err      error
	)
	Init()
	// Loop through packets in file
	handle, err = pcap.OpenOffline(config.BASE_PATH + pcapFile)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	defer handle.Close()
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
			ipPackets[len(packet.Data())] = packet
		}
	}

	sizes := []int{}
	for key := range ipPackets {
		sizes = append(sizes, key)
	}
	sort.Ints(sizes)

	for _, size := range sizes {
		packet := ipPackets[size]
		//Warmup
		fwd.ID(packet)
		b.Run(fmt.Sprintf("packet size %d", size), func(b *testing.B) {
			b.SetBytes(int64(size))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				fwd.ID(packet)
			}
		})
	}
}

func BenchmarkEncryption(b *testing.B) {
	var ipPackets = make(map[int]gopacket.Packet)

	var (
		pcapFile string = "/Gateway_TP/pcap_files/egress.pcap"
		handle   *pcap.Handle
		err      error
	)
	Init()
	// Loop through packets in file
	handle, err = pcap.OpenOffline(config.BASE_PATH + pcapFile)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	defer handle.Close()
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
			ipPackets[len(packet.Data())] = packet
		}
	}

	sizes := []int{}
	for key := range ipPackets {
		sizes = append(sizes, key)
	}
	sort.Ints(sizes)

	for _, size := range sizes {
		b.Run(fmt.Sprintf("packet size %d", size), func(b *testing.B) {
			packet := ipPackets[size]

			srcTP, destTP, zone, key, err := fwd.GetMondrianInfoEgress(packet)
			if err != nil {
				fmt.Println(err)
			}

			srcTP_ip, _, _ := net.ParseCIDR(srcTP + "/8")
			destTP_ip, _, _ := net.ParseCIDR(destTP + "/8")

			// Get eth header to restore it afterwards
			var eth layers.Ethernet
			eth_parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth)
			eth_decoded := []gopacket.LayerType{}
			eth_parser.DecodeLayers(packet.Data(), &eth_decoded)
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
			pld := gopacket.Payload(packet.Data())
			buf := gopacket.NewSerializeBuffer()
			opts := gopacket.SerializeOptions{}
			// Create intermediate packet stored in buf
			err = gopacket.SerializeLayers(buf, opts, &eth, &ip, &m, &pld)
			if err != nil {
				fmt.Println(err)
			}

			tmp_pkt := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
			// Get Mondrian Layer from tmp_pkt
			mlayer := tmp_pkt.Layer(mondrian.MondrianLayerType).(*mondrian.MondrianLayer)
			//Warmup
			err = mlayer.Encrypt(key[:])
			if err != nil {
				fmt.Println(err.Error())
			}

			b.SetBytes(int64(size))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Get Mondrian Layer from tmp_pkt
				mlayer := tmp_pkt.Layer(mondrian.MondrianLayerType).(*mondrian.MondrianLayer)
				err = mlayer.Encrypt(key[:])
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		})
	}
}

func BenchmarkDecryption(b *testing.B) {
	var mondrianPackets = make(map[int]gopacket.Packet)

	var (
		pcapFile string = "/Gateway_TP/pcap_files/ingress.pcap"
		handle   *pcap.Handle
		err      error
	)
	Init()

	// Loop through packets in file
	handle, err = pcap.OpenOffline(config.BASE_PATH + pcapFile)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	defer handle.Close()
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
			if mLayer := packet.Layer(mondrian.MondrianLayerType); mLayer != nil {
				mondrianPackets[len(packet.Data())] = packet
			}
		}
	}
	sizes := []int{}
	for key := range mondrianPackets {
		sizes = append(sizes, key)
	}
	sort.Ints(sizes)

	for _, size := range sizes {
		b.Run(fmt.Sprintf("packet size %d", size), func(b *testing.B) {
			packet := mondrianPackets[size]

			// Get the key for decryption
			_, _, _, key, err := fwd.GetMondrianInfoIngress(packet)
			if err != nil {
				fmt.Println(err.Error())
			}
			// Get Mondrian layer
			tmp_packet := gopacket.NewPacket(packet.Data(), layers.LayerTypeEthernet, gopacket.Default)
			mlayer := tmp_packet.Layer(mondrian.MondrianLayerType).(*mondrian.MondrianLayer)

			//Warmup
			// Decrypt/verify the packet
			err = mlayer.Decrypt(key[:])
			if err != nil {
				fmt.Println(err.Error())
			}

			b.SetBytes(int64(size))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Get Mondrian layer
				tmp_packet = gopacket.NewPacket(packet.Data(), layers.LayerTypeEthernet, gopacket.Default)
				mlayer = tmp_packet.Layer(mondrian.MondrianLayerType).(*mondrian.MondrianLayer)
				// Decrypt/verify the packet
				err = mlayer.Decrypt(key[:])
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		})
	}
}

func BenchmarkSendPacket(b *testing.B) {
	var mondrianPackets = make(map[int]gopacket.Packet)

	var (
		pcapFile string = "/Gateway_TP/pcap_files/ingress.pcap"
		handle   *pcap.Handle
		err      error
	)
	Init()

	// Loop through packets in file
	handle, err = pcap.OpenOffline(config.BASE_PATH + pcapFile)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	defer handle.Close()
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
			if mLayer := packet.Layer(mondrian.MondrianLayerType); mLayer != nil {
				mondrianPackets[len(packet.Data())] = packet
			}
		}
	}
	sizes := []int{}
	for key := range mondrianPackets {
		sizes = append(sizes, key)
	}
	sort.Ints(sizes)

	for _, size := range sizes {
		b.Run(fmt.Sprintf("packet size %d", size), func(b *testing.B) {
			packet := mondrianPackets[size]
            //time.Sleep(5*time.Second)
			//Warmup
			ext_iface.Send_Packet(packet.Data())
            
			b.SetBytes(int64(size))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ext_iface.Send_Packet(packet.Data())
			}
		})
	}
}

func BenchmarkIteratePackets(b *testing.B) {
	//var mondrianPackets = make(map[int]gopacket.Packet)

	var (
		pcapFile string = "/Gateway_TP/pcap_files/ingress.pcap"
		handle   *pcap.Handle
		err      error
	)
	Init()

	// Loop through packets in file
	handle, err = pcap.OpenOffline(config.BASE_PATH + pcapFile)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	defer handle.Close()
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	b.Run("packet size varying", func(b *testing.B) {

		//Warmup
		for packet := range packetSource.Packets() {
			ignore(packet)
		}

		//b.SetBytes(int64(size))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for packet := range packetSource.Packets() {
				ignore(packet)
			}
		}
	})

}

func ignore(pkt gopacket.Packet) {

}
