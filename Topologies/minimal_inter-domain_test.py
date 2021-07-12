#!/usr/bin/python
from mininet.net import Mininet
from mininet.node import Controller, Node, RemoteController, OVSKernelSwitch, IVSSwitch, UserSwitch
from mininet.link import Link, TCLink
from mininet.cli import CLI
from mininet.log import setLogLevel

def topology():

    "Create a network."
    net = Mininet( controller=RemoteController, link=TCLink, switch=OVSKernelSwitch )

    print ("*** Creating nodes")
    host_ip = {}
    host_ip['h1']='10.0.1.1/16' # Site 1 Zone 1
    host_ip['h2']='10.0.1.2/16' # Site 1 Zone 1
    host_ip['h3']='10.0.2.1/16' # Site 1 Zone 2
    host_ip['h4']='20.0.1.1/16' # Site 2 Zone 1
    host_ip['h5']='20.0.2.1/16' # Site 2 Zone 2
    gateway_ip = {}
    gateway_ip['g1'] = '10.0.0.1/8'
    gateway_ip['g2'] = '20.0.0.1/8'
    hosts = []
    # Add hosts
    for key, value in host_ip.items():
        if value.split(".")[0]=='10':
            hosts.append(net.addHost(key, ip=value, defaultRoute = "via "+str(gateway_ip['g1'])))
        if value.split(".")[0]=='20':
            hosts.append(net.addHost(key, ip=value, defaultRoute = "via "+str(gateway_ip['g2'])))
    # Add Gateway TPs
    for key, value in gateway_ip.items():
        if value.split(".")[0]=='10':
            # Default gateway of g1 is g2
            g1 = net.addHost(key, ip=value, defaultRoute = "via 30.0.0.2/8")
            #g1 = net.addHost(key, ip=value, defaultRoute = "via 0.0.0.0")
            g1.cmd('sysctl -w net.ipv4.ip_forward=1')
            g1.cmd('ifconfig g1-eth0 10.0.0.1 netmask 255.0.0.0 up')
            g1.cmd('ifconfig g1-eth1 30.0.0.1 netmask 255.0.0.0 up')
            g1.cmd('ip route 20.0.0.0/8 via 30.0.0.2/8')
            #g1.cmd('iptables -A FORWARD -i g1-eth0 -s 10.0.0.0/8 -o g1-eth1 -d 20.0.0.0/8 -j ACCEPT')
            #g1.cmd('iptables -A FORWARD -i g1-eth1 -d 20.0.0.0/8 -o g1-eth0 -s 10.0.0.0/8 -j ACCEPT')
        if value.split(".")[0]=='20':
            # Default gateway of g2 is g1
            g2 = net.addHost(key, ip=value, defaultRoute = "via 30.0.0.1/8")
            #g2 = net.addHost(key, ip=value, defaultRoute = "via 0.0.0.0")
            g2.cmd('sysctl -w net.ipv4.ip_forward=1')
            g2.cmd('ifconfig g2-eth0 20.0.0.1 netmask 255.0.0.0 up')
            g2.cmd('ifconfig g2-eth1 30.0.0.2 netmask 255.0.0.0 up')
            g2.cmd('ip route 10.0.0.0/8 via 30.0.0.1/8')
            #g2.cmd('iptables -A FORWARD -i g2-eth0 -s 20.0.0.0/8 -o g2-eth1 -d 10.0.0.0/8 -j ACCEPT')
            #g2.cmd('iptables -A FORWARD -i g2-eth1 -d 10.0.0.0/8 -o g2-eth0 -s 20.0.0.0/8 -j ACCEPT')

    # Controller Site 1
    c1 = net.addController('c1',ip='127.0.0.1',port=6633)
    # Controller Site 2
    c2 = net.addController('c2',ip='127.0.0.1',port=6633)
    # Internet controller (normal router)
    #c3 = net.addController('c3',ip='127.0.0.1',port=6635)
    
    # Switch Site 1
    s1 = net.addSwitch('s1')
    # Switch Site 2
    s2 = net.addSwitch('s2')
    # Internet
    #r1 = net.addHost('r1', cls=Node, ip='0.0.0.0')
    #r1.cmd('sysctl -w net.ipv4.ip_forward=1')
    #r1.cmd('ifconfig r1-eth0 10.1.0.2 netmask 255.255.0.0 up')
    #r1.cmd('ifconfig r1-eth1 20.1.0.2 netmask 255.255.0.0 up')
    #r1.cmd('iptables -A FORWARD -i r1-eth0 -s 10.0.0.0/8 -o r1-eth1 -d 20.0.0.0/8 -j ACCEPT')
    #r1.cmd('iptables -A FORWARD -i r1-eth1 -d 20.0.0.0/8 -o r1-eth0 -s 10.0.0.0/8 -j ACCEPT')
    

    print ("*** Creating links")
    # Connect hosts to the switch of this site
    for host in hosts:
        if host_ip[host.name].split(".")[0]=='10':
            net.addLink(host, s1)
        if host_ip[host.name].split(".")[0]=='20':
            net.addLink(host, s2)
    
    # Connect the sites
    net.addLink(s1, g1)
    net.addLink(s2, g2)
    net.addLink(g1, g2)
    #net.addLink(g2, r1)
    #inf_g1_to_g2 = g1.connectionsTo(g2)
    #g1.cmdPrint("route add -net 20.0.0.0/8 gw 30.0.0.2")
    #g2.cmdPrint("route add -net 10.0.0.0/8 gw 30.0.0.1")

    #test = net.addHost('h100', ip='10.0.0.3/24', defaultRoute="via 30.0.0.2")
    #net.addLink(test, s2)

    print ("*** Starting network")
    net.build()
    # Start Controller and connect the right Switch
    # Site 1
    c1.start()
    s1.start( [c1] )
    # Site 2
    c2.start()
    s2.start( [c2] )
    # Internet
    #c3.start()
    #internet.start([c3])

    #g1.cmdPrint("route add -net 20.0.0.0/8 gw 30.0.0.2")
    #g2.cmdPrint("route add -net 10.0.0.0/8 gw 30.0.0.1")
    
    print ("*** Running CLI")
    CLI( net )
    
    print ("*** Stopping network")
    net.stop()

if __name__ == '__main__':
    setLogLevel( 'info' )
    topology()