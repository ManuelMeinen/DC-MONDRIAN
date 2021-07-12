#!/usr/bin/python
from mininet.net import Mininet
from mininet.node import Controller, RemoteController, OVSKernelSwitch, IVSSwitch, UserSwitch
from mininet.link import Link, TCLink
from mininet.cli import CLI
from mininet.log import setLogLevel

def topology():

    "Create a network."
    net = Mininet( controller=RemoteController, link=TCLink, switch=OVSKernelSwitch )

    print ("*** Creating nodes")
    h1_ip = '10.0.0.1'
    h2_ip = '10.0.0.2'
    h3_ip = '20.0.0.1'
    h1_mac = '00:00:00:00:00:01'
    h2_mac = '00:00:00:00:00:02'
    h3_mac = '00:00:00:00:00:03'
    h1 = net.addHost('h1', ip=h1_ip, mac=h1_mac)
    h2 = net.addHost('h2', ip=h2_ip, mac=h2_mac)
    h3 = net.addHost('h3', ip=h3_ip, mac=h3_mac)
    
    c0 = net.addController( 'c0',ip='127.0.0.1',port=6633 )
    
    s1 = net.addSwitch('s1')
    

    print ("*** Creating links")
    net.addLink(h1, s1)
    net.addLink(h2, s1)
    net.addLink(h3, s1)

    print ("*** Starting network")
    net.build()
    c0.start()
    s1.start( [c0] )
    
    print ("*** Running CLI")
    CLI( net )
    
    print ("*** Stopping network")
    net.stop()

if __name__ == '__main__':
    setLogLevel( 'info' )
    topology()