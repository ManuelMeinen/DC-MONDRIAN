#!/usr/bin/python3
from mininet.net import Containernet
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

class MondrianTestbed:

    def __init__(self):
        self.switches = []
        self.controllers = []
        self.gatewayTPs = []
        self.defaultGaetways = []
        self.hosts = []
        self.internet_switch = None
        self.key_net_switch = None
        self.net = None

    topo = {
        'Site 1':{
            'tpAddr':'200.0.0.1',
            'key_net_addr':'100.0.0.1',
            'ip_range':'10.0.0.0/8',
            'eTP_Port':6633, 
            'gateway_tp_name':'d1',
            'default_gw_name':'h1',
            'default_gw':'10.0.0.1',
            'site_switch':'s1',
            'site_controller':'c1',
            'gw_ctl_addr':'172.17.0.10',
            'gw_ctl_mask':'255.255.0.0',
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
            'tpAddr':'200.0.0.2',
            'key_net_addr':'100.0.0.2',
            'ip_range':'20.0.0.0/8',
            'eTP_Port':6634,
            'gateway_tp_name':'d2',
            'default_gw_name':'h2',
            'default_gw':'20.0.0.1',
            'site_switch':'s2',
            'site_controller':'c2',
            'gw_ctl_addr':'172.17.0.20',
            'gw_ctl_mask':'255.255.0.0',
            'Hosts':{
                'h21':{
                    'ip':'20.0.1.2'
                },
                'h22':{
                    'ip':'20.2.0.3'
                },
                'h23':{
                    'ip':'20.2.0.4'
                }
            }
        },
        'Site 3':{
            'tpAddr':'200.0.0.3',
            'key_net_addr':'100.0.0.3',
            'ip_range':'30.0.0.0/8',
            'eTP_Port':6635,
            'gateway_tp_name':'d3',
            'default_gw_name':'h3',
            'default_gw':'30.0.0.1',
            'site_switch':'s3',
            'site_controller':'c3',
            'gw_ctl_addr':'172.17.0.30',
            'gw_ctl_mask':'255.255.0.0',
            'Hosts':{
                'h31':{
                    'ip':'30.0.1.2'
                },
                'h32':{
                    'ip':'30.1.0.3'
                },
                'h33':{
                    'ip':'30.2.0.3'
                }
            }
        }  
    }

    def topology(self):
        "Create a network."
        net = Containernet( controller=RemoteController, link=TCLink, switch=OVSKernelSwitch )
        service.start_Mondrian_Controller()
        self.internet_switch = net.addSwitch('s200', cls=OVSKernelSwitch, failMode='standalone')
        self.key_net_switch =  net.addSwitch('s100', cls=OVSKernelSwitch, failMode='standalone')

        for site in self.topo:
            site_info = self.topo[site]
            # Start Endpoint TP
            service.start_Endpoint_TP(tpAddr=site_info['tpAddr'], endpointTPPort=site_info['eTP_Port'])
            self.controllers.append((net.addController( site_info['site_controller'],ip='localhost',port=site_info['eTP_Port']),site))
            # Start site Switch
            self.switches.append((net.addSwitch(site_info['site_switch'], cls = OVSSwitch, protocols='OpenFlow13'),site))
            # Start Default Gateway and connect to Site Switch
            self.defaultGaetways.append((net.addHost(site_info['default_gw_name'], ip=site_info['default_gw']),site))
            net.addLink(self.defaultGaetways[-1][0], self.switches[-1][0])
            # Start Hosts and connect to site Switch
            hosts_info = site_info['Hosts']
            for host in hosts_info:
                self.hosts.append((net.addHost(host, ip=hosts_info[host]['ip']),site))
                net.addLink(self.hosts[-1][0], self.switches[-1][0])
            # Start Gateway TP and connect Default Gateway -- Gateway TP, Gateway TP -- Internet Switch, Gateway TP -- Key Net Switch
            self.gatewayTPs.append((net.addDocker(site_info['gateway_tp_name'], ip='1.1.1.1', dimage="gateway_tp:1.0", volumes=["/home/mmeinen/polybox/code/DC-MONDRIAN:/vol1"]),site))
            net.addLink(self.defaultGaetways[-1][0], self.gatewayTPs[-1][0])
            net.addLink(self.gatewayTPs[-1][0], self.internet_switch)
            net.addLink(self.gatewayTPs[-1][0], self.key_net_switch)
        
        print('*** Starting network')
        net.build()
        print( '*** Starting standalone switches')
        self.internet_switch.start([])
        self.key_net_switch.start([])

        print("*** Config the hosts")
        for site in self.topo:
            site_info = self.topo[site]
            print("*** Config hosts of "+str(site))
            for h in self.hosts:
                if h[1]==site:
                    setup.set_up_interface(h[0], if_name='eth0', ip_addr=site_info['Hosts'][h[0].name]['ip'], net_mask='255.0.0.0')
                    setup.set_up_default_gw(h[0], gw=site_info['default_gw'])
            print("*** Config Default Gateway of "+str(site))
            for g in self.defaultGaetways:
                if g[1]==site:
                    setup.set_up_interface(g[0], if_name='eth0', ip_addr=site_info['default_gw'], net_mask='255.0.0.0')
                    setup.set_up_interface(g[0], if_name='eth1', ip_addr=site_info['tpAddr'], net_mask='255.0.0.0')
                    setup.set_up_forwarding(g[0]) 
                    # Add routes to other Sites
                    for other_site in self.topo:
                        if other_site != site:
                            setup.set_up_route(host=g[0], dest=self.topo[other_site]['ip_range'], via=self.topo[other_site]['tpAddr'])
            print("*** Config Gateway TP of "+str(site))
            for g in self.gatewayTPs:
                if g[1]==site:
                    setup.set_up_interface(g[0], if_name='eth0', ip_addr='1.1.1.1', net_mask='255.0.0.0')
                    setup.set_up_interface(g[0], if_name='eth1', ip_addr='1.1.1.2', net_mask='255.0.0.0')
                    setup.set_up_interface(g[0], if_name='eth2', ip_addr=site_info['key_net_addr'], net_mask='255.0.0.0')
                    setup.set_up_inet(g[0], if_name='eth0', ip_addr=site_info['gw_ctl_addr'], net_mask=site_info['gw_ctl_mask'])
                    setup.prepare_gateway_TP(g[0], site_info['tpAddr'])
        #print("*** Start Gateway TPs")
        #for g in self.gatewayTPs:
        #    setup.start_gateway_TP(g[0])
        print("*** Start Controllers")
        for c in self.controllers:
            c[0].start()
        print("*** Map Switches to Controllers")
        for s in self.switches:
            s[0].start([c[0] for c in self.controllers if c[1]==s[1]])
        self.net =  net


    def test_intra_zone_icmp(self):
        '''
        Test if all connections work for intra zone traffic
        both intra and inter domain for ICMP
        '''
        test.prefix = test.prefix+"[ICMP Intra Zone Test] "
        host_dict = self.get_host_dict()
        success = True
        zone1 = ['h11', 'h12', 'h13', 'h21', 'h31']
        zone2 = ['h22', 'h23', 'h32']
        zone3 = ['h33']
        print("*** ICMP Intra Zone Test started")
        # ICMP
        # Zone 1
        for h1 in zone1:
            for h2 in zone1:
                if h1!=h2:
                    success = success and test.test_icmp(host_dict[h1], host_dict[h2])
        # Zone 2
        for h1 in zone2:
            for h2 in zone2:
                if h1!=h2:
                    success = success and test.test_icmp(host_dict[h1], host_dict[h2])
        # Zone 3
        for h1 in zone3:
            for h2 in zone3:
                if h1!=h2:
                    success = success and test.test_icmp(host_dict[h1], host_dict[h2])
        if success:
            print("*** ICMP Intra Zone SUCCESS")
        else:
            print("*** ICMP Intra Zone FAIL")
        test.prefix = ""
        return success

    def test_intra_zone_tcp(self):
        '''
        Test if all connections work for intra zone traffic
        both intra and inter domain for TCP
        '''
        test.prefix = test.prefix+"[TCP Intra Zone Test] "
        host_dict = self.get_host_dict()
        success = True
        zone1 = ['h11', 'h12', 'h13', 'h21', 'h31']
        zone2 = ['h22', 'h23', 'h32']
        zone3 = ['h33']
        print("*** TCP Intra Zone Test started")
        # TCP
        # Zone 1
        for h1 in zone1:
            for h2 in zone1:
                if h1!=h2:
                    success = success and test.test_tcp(host_dict[h2], host_dict[h1], listen_timeout=60)
                    if not success:
                        # Try again on failure to check if the test is just shitty
                        success = test.test_tcp(host_dict[h2], host_dict[h1], listen_timeout=120)
                        if success:
                            print("TCP Test was just instable")
        # Zone 2
        for h1 in zone2:
            for h2 in zone2:
                if h1!=h2:
                    success = success and test.test_tcp(host_dict[h2], host_dict[h1], listen_timeout=60)
                    if not success:
                        # Try again on failure to check if the test is just shitty
                        success = test.test_tcp(host_dict[h2], host_dict[h1], listen_timeout=120)
                        if success:
                            print("TCP Test was just instable")
        # Zone 3
        for h1 in zone3:
            for h2 in zone3:
                if h1!=h2:
                    success = success and test.test_tcp(host_dict[h2], host_dict[h1], listen_timeout=60)
                    if not success:
                        # Try again on failure to check if the test is just shitty
                        success = test.test_tcp(host_dict[h2], host_dict[h1], listen_timeout=120)
                        if success:
                            print("TCP Test was just instable")
        if success:
            print("*** TCP Intra Zone SUCCESS")
        else:
            print("*** TCP Intra Zone FAIL")
        test.prefix = ""
        return success

    def test_intra_zone_udp(self):
        '''
        Test if all connections work for intra zone traffic
        both intra and inter domain for UDP
        '''
        test.prefix = test.prefix+"[UDP Intra Zone Test] "
        host_dict = self.get_host_dict()
        success = True
        zone1 = ['h11', 'h12', 'h13', 'h21', 'h31']
        zone2 = ['h22', 'h23', 'h32']
        zone3 = ['h33']
        print("*** UDP Intra Zone Test started")
        # UDP
        # Zone 1
        for h1 in zone1:
            for h2 in zone1:
                if h1!=h2:
                    success = success and test.test_udp(host_dict[h2], host_dict[h1])
        # Zone 2
        for h1 in zone2:
            for h2 in zone2:
                if h1!=h2:
                    success = success and test.test_udp(host_dict[h2], host_dict[h1])
        # Zone 3
        for h1 in zone3:
            for h2 in zone3:
                if h1!=h2:
                    success = success and test.test_udp(host_dict[h2], host_dict[h1])
        if success:
            print("*** UDP Intra Zone SUCCESS")
        else:
            print("*** UDP Intra Zone FAIL")
        test.prefix = ""
        return success


    def test_intra_zone(self):
        '''
        Test if all connections work for intra zone traffic (Zone 1)
        both intra and inter domain for the protocols TCP, UDP and ICMP
        '''

        icmp_result = self.test_intra_zone_icmp()
        udp_result = self.test_intra_zone_udp()
        tcp_result = self.test_intra_zone_tcp()
        test.prefix = "[Intra Zone Test] "
        print(test.prefix+"Intra Zone Test Results:")
        if icmp_result:
            print(test.prefix+"     "+"ICMP Intra Zone Test: SUCCESS")
        else:
            print(test.prefix+"     "+"ICMP Intra Zone Test: FAIL")
        if udp_result:
            print(test.prefix+"     "+"UDP Intra Zone Test: SUCCESS")
        else:
            print(test.prefix+"     "+"UDP Intra Zone Test: FAIL")
        if tcp_result:
            print(test.prefix+"     "+"TCP Intra Zone Test: SUCCESS")
        else:
            print(test.prefix+"     "+"TCP Intra Zone Test: FAIL")
        test.prefix = ""

    
    def test_inter_zone(self):
        #TODO: This is still the old code --> not valid now
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
        
    
    def capture_traffic(self):
        d = self.gatewayTPs[0][0]
        cmd = "tcpdump -w /vol1/egress.pcap -i "+str(d.name)+"-eth0 &"
        print(cmd)
        #print(d.cmd(cmd))
        cmd = "tcpdump -w /vol1/ingress.pcap -i "+str(d.name)+"-eth1 &"
        print(cmd)
        #print(d.cmd(cmd))
        self.startCLI()
        # Generate some traffic
        host_dict = self.get_host_dict()
        test.test_icmp(src=host_dict['h11'], dest=host_dict['h21'])
        test.test_icmp(src=host_dict['h21'], dest=host_dict['h11'])
        test.test_udp(src=host_dict['h11'], dest=host_dict['h21'])
        test.test_udp(src=host_dict['h21'], dest=host_dict['h11'])
        test.test_tcp(src=host_dict['h11'], dest=host_dict['h21'])
        test.test_tcp(src=host_dict['h21'], dest=host_dict['h11'])


        


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
    topo = MondrianTestbed()
    topo.topology()
    #Make sure that everything is ready
    time.sleep(3)
    #topo.test_intra_zone()
    #topo.test_inter_zone()
    #topo.test()
    #topo.capture_traffic()
    topo.startCLI()
    topo.stopNet()