#!/usr/bin/python
from mininet.net import Mininet
from mininet.node import Controller, RemoteController, OVSKernelSwitch, IVSSwitch, UserSwitch
from mininet.link import Link, TCLink
from mininet.cli import CLI
from mininet.log import setLogLevel
import time

def topology():

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
    net.addLink(h1, h2)
    net.addLink(h2, h3)
    net.addLink(h3, h4)

    print ("*** Starting network")
    net.build()

    h1.cmd('ifconfig h1-eth0 10.0.0.2 netmask 255.0.0.0 up')
    #set_up_interface(host=h1, if_name='eth0', ip_addr='10.0.0.2', net_mask='255.0.0.0') #TODO: why doesn't that work?...
    h2.cmd('ifconfig h2-eth1 30.0.0.1 netmask 255.0.0.0 up') 
    #set_up_interface(host=h2, if_name='eth1', ip_addr='30.0.0.1', net_mask='255.0.0.0')
    h2.cmd('ifconfig h2-eth0 10.0.0.1 netmask 255.0.0.0 up') 
    #set_up_interface(host=h2, if_name='eth0', ip_addr='10.0.0.1', net_mask='255.0.0.0')
    h3.cmd('ifconfig h3-eth1 20.0.0.1 netmask 255.0.0.0 up')
    h3.cmd('ifconfig h3-eth0 30.0.0.2 netmask 255.0.0.0 up')
    h4.cmd('ifconfig h4-eth0 20.0.0.2 netmask 255.0.0.0 up')
    
    h2.cmd('sysctl -w net.ipv4.ip_forward=1')  
    h3.cmd('sysctl -w net.ipv4.ip_forward=1')
    
    h1.cmd('ip route add default via 10.0.0.1')    
    h1.cmd('ip route change default via 10.0.0.1') 
    h2.cmd('ip route add default via 30.0.0.2')    
    h2.cmd('ip route change default via 30.0.0.2') 
    h3.cmd('ip route add default via 30.0.0.1')    
    h3.cmd('ip route change default via 30.0.0.1') 
    h4.cmd('ip route add default via 20.0.0.1')    
    h4.cmd('ip route change default via 12.0.0.1') 
    
    h2.cmd('ip route add 20.0.0.0/8 via 30.0.0.2')
    h3.cmd('ip route add 10.0.0.0/8 via 30.0.0.1')
    


    
    

    
    

    
    
    
    print ("*** Running CLI")
    CLI( net )
    
    print ("*** Stopping network")
    net.stop()

def set_up_interface(host, if_name, ip_addr, net_mask):
    cmd = str(host.name)+' ifconfig '+str(host.name)+'-'+str(if_name)+' '+str(ip_addr)+' netmask '+str(net_mask)+' up'
    print("[Set up interface] "+cmd)
    print(type(host))
    host.cmd(cmd)

if __name__ == '__main__':
    setLogLevel( 'info' )
    topology()