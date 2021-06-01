#!/usr/bin/python
from mininet.net import Mininet
from mininet.node import Controller, RemoteController, OVSKernelSwitch, IVSSwitch, UserSwitch
from mininet.link import Link, TCLink
from mininet.cli import CLI
from mininet.log import setLogLevel
import time
from util.setup import SetupUtil
from util.test import TestUtil

setup = SetupUtil()
test = TestUtil()

class EndpointTPTestbed:

    def __init__(self):
        pass

    def topology(self):
        "Create a network."
        net = Mininet( controller=RemoteController, link=TCLink, switch=OVSKernelSwitch )

        print ("*** Creating nodes")
        h1_ip = '10.0.0.2'
        h2_ip = '30.0.0.1'
        h3_ip = '30.0.0.2'
        h4_ip = '20.0.0.2'

        h1_mac = '00:00:00:00:00:01'
        h2_mac = '00:00:00:00:00:02'
        h3_mac = '00:00:00:00:00:03'
        h4_mac = '00:00:00:00:00:04'

        h1 = net.addHost('h1', ip=h1_ip, mac=h1_mac)
        h2 = net.addHost('h2', ip=h2_ip, mac=h2_mac)
        h3 = net.addHost('h3', ip=h3_ip, mac=h3_mac)
        h4 = net.addHost('h4', ip=h4_ip, mac=h4_mac)

        print ("*** Creating links")
        # Create intra domain links first --> intra domain links at eth0
        net.addLink(h1, h2)
        net.addLink(h4, h3)
        # Create inter domain links next --> inter domain links at eth1
        net.addLink(h2, h3)

        print ("*** Starting network")
        net.build()
        # Configure intra domain links 
        # Site 1
        setup.set_up_interface(host=h1, if_name='eth0', ip_addr='10.0.0.2', net_mask='255.0.0.0')
        setup.set_up_interface(host=h2, if_name='eth0', ip_addr='10.0.0.1', net_mask='255.0.0.0')
        # Site 2
        setup.set_up_interface(host=h4, if_name='eth0', ip_addr='20.0.0.2', net_mask='255.0.0.0')
        setup.set_up_interface(host=h3, if_name='eth0', ip_addr='20.0.0.1', net_mask='255.0.0.0')

        # Configure inter domain links 
        # Site 1
        setup.set_up_interface(host=h2, if_name='eth1', ip_addr='30.0.0.1', net_mask='255.0.0.0')
        # Site 2
        setup.set_up_interface(host=h3, if_name='eth1', ip_addr='30.0.0.2', net_mask='255.0.0.0')

        # Set up IP forwarding on the Gateway TPs
        setup.set_up_forwarding(h2)  
        setup.set_up_forwarding(h3)

        # Set up default gateways
        setup.set_up_default_gw(h1, '10.0.0.1')
        setup.set_up_default_gw(h2, '30.0.0.2')
        setup.set_up_default_gw(h3, '30.0.0.1')
        setup.set_up_default_gw(h4, '20.0.0.1') 

        setup.set_up_route(host=h2, dest='20.0.0.0/8', via='30.0.0.2')
        setup.set_up_route(host=h3, dest='10.0.0.0/8', via='30.0.0.1')

        self.net =  net


        #test.test_udp(h1, h4)
        #test.test_udp(h4, h1)
        #test.test_tcp(h1, h4)
        #test.test_tcp(h4, h1)
        #test.test_icmp(h1, h4)
        #test.test_icmp(h4, h1)
    def startCLI(self): 
        print ("*** Running CLI")
        CLI( self.net )

    def stopNet(self):  
        print ("*** Stopping network")
        self.net.stop()


if __name__ == '__main__':
    setLogLevel( 'info' )
    topo = EndpointTPTestbed()
    topo.topology()
    topo.startCLI()
    topo.stopNet()