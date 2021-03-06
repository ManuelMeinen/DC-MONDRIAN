#!/usr/bin/python3
from mininet.net import Mininet
from mininet.node import Controller, OVSSwitch, RemoteController, OVSKernelSwitch, IVSSwitch, UserSwitch
from mininet.link import Link, TCLink
from mininet.cli import CLI
from mininet.log import setLogLevel
import time
import threading
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

    def test_intra_zone(self):
        '''
        Test if all connections work for intra zone traffic (Zone 1)
        both intra and inter domain for the protocols TCP, UDP and ICMP
        '''
        test.prefix = "[Intra Zone Test] "
        host_dict = self.get_host_dict()
        success = True
        print("*** Intra Zone Test started")
        # ICMP
        success = success and test.test_icmp(host_dict['h11'], host_dict['h12'])
        success = success and test.test_icmp(host_dict['h11'], host_dict['h21'])
        success = success and test.test_icmp(host_dict['h12'], host_dict['h11'])
        success = success and test.test_icmp(host_dict['h12'], host_dict['h21'])
        success = success and test.test_icmp(host_dict['h21'], host_dict['h11'])
        success = success and test.test_icmp(host_dict['h21'], host_dict['h12'])
        # TCP
        success = success and test.test_tcp(host_dict['h11'], host_dict['h12'])
        success = success and test.test_tcp(host_dict['h11'], host_dict['h21'])
        success = success and test.test_tcp(host_dict['h12'], host_dict['h11'])
        success = success and test.test_tcp(host_dict['h12'], host_dict['h21'])
        success = success and test.test_tcp(host_dict['h21'], host_dict['h11'])
        success = success and test.test_tcp(host_dict['h21'], host_dict['h12'])
        # UDP
        success = success and test.test_udp(host_dict['h11'], host_dict['h12'])
        success = success and test.test_udp(host_dict['h11'], host_dict['h21'])
        success = success and test.test_udp(host_dict['h12'], host_dict['h11'])
        success = success and test.test_udp(host_dict['h12'], host_dict['h21'])
        success = success and test.test_udp(host_dict['h21'], host_dict['h11'])
        success = success and test.test_udp(host_dict['h21'], host_dict['h12'])
        
        if success:
            print("*** Intra Zone Test passed")
        else: 
            print("*** Intra Zone Test failed")
        test.prefix = ""
    
    def test_inter_zone(self):
        '''
        Test if connections work for inter zone traffic for which
        there is a policy allowing that kind of traffic and that
        the don't if the policy disallows it.
        '''
        test.prefix = "[Inter Zone Test] "
        host_dict = self.get_host_dict()
        success = True
        print("*** Inter Zone Test started")
        #"PolicyID": 1, "Src": 1, "Dest": 2,"SrcPort": 70, "DestPort": 90, "Proto": "TCP", "Action": "forwarding"
        success = success and not(test.test_tcp(src=host_dict['h11'], dest=host_dict['h13'], srcPort=70, destPort=90)) # Fail because for TCP established would be needed
        #"PolicyID": 2, "Src": 2, "Dest": 1, "SrcPort": 70, "DestPort": 90, "Proto": "UDP", "Action": "forwarding"
        success = success and test.test_udp(src=host_dict['h13'], dest=host_dict['h11'], srcPort=70, destPort=90) # OK
        #"PolicyID": 3, "Src": 1, "Dest": 2, "SrcPort": 0, "DestPort": 0, "Proto": "TCP", "Action": "forwarding"
        success = success and not(test.test_tcp(src=host_dict['h11'], dest=host_dict['h13'])) # Fail because for TCP established would be needed
        #"PolicyID": 4, "Src": 3, "Dest": 0, "SrcPort": 0, "DestPort": 0, "Proto": "", "Action": "drop"
        success = success and not(test.test_udp(src=host_dict['h22'], dest=host_dict['h13'])) # Fail because of drop action
        success = success and not(test.test_udp(src=host_dict['h22'], dest=host_dict['h11'])) # Fail because of drop action
        #"PolicyID": 5, "Src": 1, "Dest": 2, "SrcPort": 80, "DestPort": 100, "Proto": "TCP", "Action": "established"
        success = success and test.test_tcp(src=host_dict['h11'], dest=host_dict['h13'], srcPort=80, destPort=100)  #OK 
        success = success and test.test_tcp(src=host_dict['h13'], dest=host_dict['h11'], srcPort=100, destPort=80)  #OK 
        #"PolicyID": 6, "Src": 1, "Dest": 2, "SrcPort": 80, "DestPort": 0, "Proto": "TCP", "Action": "drop"
        success = success and not(test.test_tcp(src=host_dict['h11'], dest=host_dict['h13'], srcPort=80)) # Fail because we drop
        #"PolicyID": 7, "Src": 2, "Dest": 1, "SrcPort": 0, "DestPort": 100, "Proto": "UDP", "Action": "established"
        success = success and test.test_udp(src=host_dict['h13'], dest=host_dict['h12'], srcPort=123, destPort=100)
        success = success and test.test_udp(src=host_dict['h12'], dest=host_dict['h13'], srcPort=100, destPort=123)
        #"PolicyID": 8, "Src": 1, "Dest": 3, "SrcPort": 0, "DestPort": 0, "Proto": "", "Action": "established"
        success = success and test.test_icmp(src=host_dict['h11'], dest=host_dict['h22'])
        success = success and test.test_icmp(src=host_dict['h22'], dest=host_dict['h11'])

        if success:
            print("*** Inter Zone Test passed")
        else: 
            print("*** Inter Zone Test failed")
        test.prefix = ""
        
    def test(self):
        '''
        just test only one thing
        '''
        test.prefix = "[Test] "
        host_dict = self.get_host_dict()
       
          #"PolicyID": 5, "Src": 1, "Dest": 2, "SrcPort": 80, "DestPort": 100, "Proto": "TCP", "Action": "established"
        success = test.test_tcp(src=host_dict['h11'], dest=host_dict['h13'], srcPort=80, destPort=100)  #OK 
        success = test.test_tcp(src=host_dict['h13'], dest=host_dict['h11'], srcPort=100, destPort=80)  #OK 

        test.prefix = ""

    def traffic_generator(self):
        '''
        generate some traffic for the packet-in benchmarking
        '''
        test.prefix = "[ICMP Traffic Generator] "
        host_dict = self.get_host_dict()
        while True:
            t = time.time()
            for i in range(9):
                # 90% intra-zone
                _ = self.net.ping([host_dict['h11'], host_dict['h12'], host_dict['h21']])
                time.sleep(1)
            for i in range(1):
                # 10% mix
                _ = self.net.ping([host_dict['h11'], host_dict['h12'], host_dict['h13'], host_dict['h21'], host_dict['h22']])
            print("*** Time for this round was: "+str(time.time()-t)+"s")

    def traffic_generator_intra_zone(self):
        '''
        generate intra zone traffic for benchmarking packet-in messages
        '''
        test.prefix = "[ICMP Traffic Generator] "
        host_dict = self.get_host_dict()
        while True:
            t = time.time()            
            _ = self.net.ping([host_dict['h11'], host_dict['h12'], host_dict['h21']])
            time.sleep(2)
    
    def traffic_generator_inter_zone(self):
        '''
        generate inter zone traffic for benchmarking packet-in messages
        '''
        test.prefix = "[ICMP Traffic Generator] "
        host_dict = self.get_host_dict()
        while True:
            t = time.time()  
            _ = self.net.ping([host_dict['h11'], host_dict['h12'], host_dict['h13'], host_dict['h21'], host_dict['h22']])
            time.sleep(1)
    
    
    def get_host_dict(self):
        host_dict = {}
        for host in self.hosts:
            host_dict[host[0].name] = host[0]
        return host_dict


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
    #Make sure that everything is ready
    time.sleep(3)
    try:
        #topo.test_intra_zone()
        #topo.test_inter_zone()
        #topo.test()
        #t1 = threading.Thread(target=topo.traffic_generator_inter_zone)
        #t1.daemon = True
        #
        #t2 = threading.Thread(target=topo.traffic_generator_intra_zone)
        #t2.daemon = True
        #t1.start()
        #t2.start()
        #while True:
        #    pass
        topo.traffic_generator_intra_zone()
        topo.startCLI()
    finally:
        topo.stopNet()