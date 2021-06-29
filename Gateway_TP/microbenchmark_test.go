package main

// go test -v -bench=.
import (
	"fmt"
	"gateway_tp/config"
	"gateway_tp/forwarder"
	"gateway_tp/logger"
    "gateway_tp/mondrian"
	"log"
	"testing"
    "sort"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)
var initialized = false
var fwd *forwarder.Forwarder

func Init(){
    if initialized == false{
        config.Init("config/config.json")
	    logger.InitLogger()
	    defer logger.CloseLogger()
	    fwd = forwarder.NewForwarder()
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
    handle, err = pcap.OpenOffline(config.BASE_PATH+pcapFile)
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
    sizes :=[]int{}
    for key := range mondrianPackets{
        sizes = append(sizes, key)
    }
    sort.Ints(sizes)
    
	for _, size := range sizes {
		b.Run(fmt.Sprintf("packet size %d", size), func(b *testing.B) {
            packet := mondrianPackets[size]
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
    handle, err = pcap.OpenOffline(config.BASE_PATH+pcapFile)
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

    sizes :=[]int{}
    for key := range ipPackets{
        sizes = append(sizes, key)
    }
    sort.Ints(sizes)
    
	for _, size := range sizes {
        packet := ipPackets[size]
		b.Run(fmt.Sprintf("packet size %d", size), func(b *testing.B) {
			b.SetBytes(int64(size))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
                fwd.ToMondrian(packet)  
			}
		})
	}
}
