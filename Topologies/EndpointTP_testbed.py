#!/usr/bin/python
from mininet.net import Mininet
from mininet.node import Controller, OVSSwitch, RemoteController, OVSKernelSwitch, IVSSwitch, UserSwitch
from mininet.link import Link, TCLink
from mininet.cli import CLI
from mininet.log import setLogLevel
import time
from util.setup import SetupUtil
from util.test import TestUtil
from util.services import ServicesUtil

setup = SetupUtil()
test = TestUtil()
service = ServicesUtil()

class EndpointTPTestbed:

    def __init__(self):
        self.switches = []
        self.controllers = []
        self.gatewayTPs = []
        self.hosts = []

    topo = {
        'Site 1':{
            'tpAddr':'30.0.0.1',
            'ip_range':'10.0.0.0/8',
            'eTP_Port':6633, 
            'default_gw_name':'h1',
            'default_gw':'10.0.0.1',
            'site_switch':'s1',
            'site_controller':'c1',
            'Hosts':{
                'h11':{
                    'ip':'10.0.1.2'
                },
                'h12':{
                    'ip':'10.1.0.3'
                },
                'h13':{
                    'ip':'10.2.0.4'
                }
            }
        },
        'Site 2':{
            'tpAddr':'30.0.0.2',
            'ip_range':'20.0.0.0/8',
            'eTP_Port':6634,
            'default_gw_name':'h2',
            'default_gw':'20.0.0.1',
            'site_switch':'s2',
            'site_controller':'c2',
            'Hosts':{
                'h21':{
                    'ip':'20.0.1.2'
                },
                'h22':{
                    'ip':'20.2.0.3'
                }
            }
        }
    }

    def topology(self):
        "Create a network."
        net = Mininet( controller=RemoteController, link=TCLink, switch=OVSKernelSwitch )
        service.start_Mondrian_Controller()

        for site in self.topo:
            site_info = self.topo[site]
            # Start Endpoint TP
            service.start_Endpoint_TP(tpAddr=site_info['tpAddr'], endpointTPPort=site_info['eTP_Port'])
            self.controllers.append((net.addController( site_info['site_controller'],ip='localhost',port=site_info['eTP_Port']),site))
            # Start site Switch
            self.switches.append((net.addSwitch(site_info['site_switch'], cls = OVSSwitch, protocols='OpenFlow13'),site))
            # Start Gateway TP and connect to site Switch
            self.gatewayTPs.append((net.addHost(site_info['default_gw_name'], ip=site_info['default_gw']),site))
            net.addLink(self.gatewayTPs[-1][0], self.switches[-1][0])
            # Start Hosts and connect to site Switch
            hosts_info = site_info['Hosts']
            for host in hosts_info:
                self.hosts.append((net.addHost(host, ip=hosts_info[host]['ip']),site))
                net.addLink(self.hosts[-1][0], self.switches[-1][0])
        # Create full mesh between Gateway TPs
        for i, g0 in zip(range(len(self.gatewayTPs)), self.gatewayTPs):
            for j, g1 in zip(range(len(self.gatewayTPs)), self.gatewayTPs):
                if j>i:
                    net.addLink(g0[0], g1[0])
        print ("*** Starting network")
        net.build()

        print("*** Config the hosts")
        for site in self.topo:
            site_info = self.topo[site]
            print("*** Config hosts of "+str(site))
            for h in self.hosts:
                if h[1]==site:
                    setup.set_up_interface(h[0], if_name='eth0', ip_addr=site_info['Hosts'][h[0].name]['ip'], net_mask='255.0.0.0')
                    setup.set_up_default_gw(h[0], gw=site_info['default_gw'])
            print("*** Config Gateway TP of "+str(site))
            for g in self.gatewayTPs:
                if g[1]==site:
                    setup.set_up_interface(g[0], if_name='eth0', ip_addr=site_info['default_gw'], net_mask='255.0.0.0')
                    setup.set_up_interface(g[0], if_name='eth1', ip_addr=site_info['tpAddr'], net_mask='255.0.0.0')
                    setup.set_up_forwarding(g[0]) 
                    # Add routes to other Sites
                    for other_site in self.topo:
                        if other_site != site:
                            setup.set_up_route(host=g[0], dest=self.topo[other_site]['ip_range'], via=self.topo[other_site]['tpAddr'])
    
        print("*** Start Controllers")
        for c in self.controllers:
            c[0].start()
        print("*** Map Switches to Controllers")
        for s in self.switches:
            s[0].start([c[0] for c in self.controllers if c[1]==s[1]])
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
        service.kill_processes()


if __name__ == '__main__':
    setLogLevel( 'info' )
    topo = EndpointTPTestbed()
    topo.topology()
    # TODO: Run some tests
    topo.startCLI()
    topo.stopNet()