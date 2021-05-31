#!/usr/bin/python
from mininet.net import Mininet
from mininet.node import Controller, OVSSwitch, RemoteController, OVSKernelSwitch, IVSSwitch, UserSwitch
from mininet.link import Link, TCLink
from mininet.cli import CLI
from mininet.log import setLogLevel

def topology():

    "Create a network."
    net = Mininet( controller=RemoteController, link=TCLink, switch=OVSKernelSwitch )

    print ("*** Creating nodes")
    h1_ip = '10.0.0.1'
    h2_ip = '10.0.0.2'
    h1 = net.addHost('h1', ip=h1_ip)
    h2 = net.addHost('h2', ip=h2_ip)
    
    c0 = net.addController( 'c0',ip='127.0.0.1',port=6633 )
    
    s1 = net.addSwitch('s1', cls = OVSSwitch, protocols='OpenFlow13')
    

    print ("*** Creating links")
    net.addLink(h1, s1)
    net.addLink(h2, s1)

    print ("*** Starting network")
    net.build()
    c0.start()
    s1.start( [c0] )

    print("*** Test TCP traffic")
    h1.cmd("nc –t –l 9999")
    h2.cmd("nc -t "+h1_ip+" 9999 < _testFile.txt")
    
    print ("*** Running CLI")
    CLI( net )

    print ("*** Stopping network")
    net.stop()

if __name__ == '__main__':
    setLogLevel( 'info' )
    topology()