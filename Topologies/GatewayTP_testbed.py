#!/usr/bin/python3
"""
This is the most simple example to showcase Containernet.
"""
from mininet.net import Containernet
from mininet.node import Controller, OVSKernelSwitch, OVSSwitch, RemoteController
from mininet.cli import CLI
from mininet.link import TCLink
from mininet.log import info, setLogLevel
import json
import time
from util.setup import SetupUtil
from util.test import TestUtil
from util.services import ServicesUtil
setLogLevel('info')

setup = SetupUtil()
test = TestUtil()
service = ServicesUtil()

class GatewayTPTestbed:

    def __init__(self):
        pass


    def topology(self):
        "Create a network."
        net = Containernet( controller=RemoteController, link=TCLink, switch=OVSKernelSwitch )
        service.start_Mondrian_Controller()

        info('*** Adding controller\n')
       # c0 = net.addController('c0')

        info('*** Adding hosts and docker containers\n')
        h11 = net.addHost('h11', ip='10.0.0.2')
        h1 = net.addHost('h1', ip='10.0.0.1')
        d1 = net.addDocker('d1', ip='40.0.0.1', dimage="gateway_tp:1.0", volumes=["/home/mmeinen/polybox/code/DC-MONDRIAN:/vol1"])
        d2 = net.addDocker('d2', ip='40.0.0.3', dimage="gateway_tp:1.0", volumes=["/home/mmeinen/polybox/code/DC-MONDRIAN:/vol1"])
        h2 = net.addHost('h2', ip='20.0.0.1')
        h21 = net.addHost('h21', ip='20.0.0.2')
        # Act as a router for controll traffic between Gateway TPs
        hctl = net.addHost('hctl', ip='100.0.0.1')
        info( '*** Add switches\n')
        # This switch is used to connect the key exchange network (100.0.0.0/8)
        s1 = net.addSwitch('s1', cls=OVSKernelSwitch, failMode='standalone')
 
        info('*** Creating links\n')
        net.addLink(h11, h1)
        net.addLink(h1, d1)
        net.addLink(h21, h2)
        net.addLink(h2, d2)
        net.addLink(d1, d2)
        net.addLink(d1, s1)
        net.addLink(d2, s1)
        net.addLink(hctl, s1)

        info('*** Starting network\n')
        net.build()
        info( '*** Starting switches\n')
        net.get('s1').start([])
        

        info('*** Configuring stuff')
        setup.set_up_interface(h11, 'eth0','10.0.0.2', '255.0.0.0')
        setup.set_up_interface(h1, 'eth0','10.0.0.1', '255.0.0.0')
        setup.set_up_interface(h1, 'eth1','30.0.0.2', '255.0.0.0')
        setup.set_up_interface(d1, 'eth0','40.0.0.1', '255.0.0.0')
        setup.set_up_interface(d1, 'eth1','40.0.0.2', '255.0.0.0')
        setup.set_up_interface(d2, 'eth0','40.0.0.3', '255.0.0.0')
        setup.set_up_interface(d2, 'eth1','40.0.0.4', '255.0.0.0')
        setup.set_up_interface(h2, 'eth0','20.0.0.1', '255.0.0.0')
        setup.set_up_interface(h2, 'eth1','30.0.0.3', '255.0.0.0')
        setup.set_up_interface(h21, 'eth0', '20.0.0.2', '255.0.0.0')
        setup.set_up_interface(hctl, 'eth0','100.0.0.1', '255.0.0.0')
        setup.set_up_interface(d1, 'eth2', '100.0.0.2', '255.0.0.0')
        setup.set_up_interface(d2, 'eth2', '100.0.0.3', '255.0.0.0')

        setup.set_up_default_gw(h1, '30.0.0.3')
        setup.set_up_default_gw(h2, '30.0.0.2')

        setup.set_up_default_gw(d1, '100.0.0.1')
        setup.set_up_default_gw(d2, '100.0.0.1')

        setup.set_up_route(h1, '20.0.0.0/8', '30.0.0.3')
        setup.set_up_route(h2, '10.0.0.0/8', '30.0.0.2')

        setup.set_up_route(d1, '40.0.0.0/8', '40.0.0.2')
        setup.set_up_route(d2, '40.0.0.0/8', '40.0.0.3')

        setup.set_up_forwarding(h1)
        setup.set_up_forwarding(h2)
        setup.set_up_forwarding(hctl)

        setup.set_up_inet(d1, 'eth0', '172.17.0.2', '255.255.0.0')
        setup.set_up_inet(d2, 'eth0', '172.17.0.3', '255.255.0.0')

        setup.set_up_default_gw(h11, '10.0.0.1')
        setup.set_up_default_gw(h21, '20.0.0.1')

        info("*** Starting the Gateway TPs")
        setup.prepare_gateway_TP(d1, "30.0.0.2")
        setup.prepare_gateway_TP(d2, "30.0.0.3")
        setup.start_gateway_TP(d1)
        setup.start_gateway_TP(d2)
        self.net = net


    def startCLI(self): 
        print ("*** Running CLI")
        CLI( self.net )

    def stopNet(self):  
        print ("*** Stopping network")
        self.net.stop()
        service.kill_processes()

if __name__=='__main__':
    topo = GatewayTPTestbed()
    topo.topology()
    topo.startCLI()
    topo.stopNet()