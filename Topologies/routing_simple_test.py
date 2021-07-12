#!/usr/bin/python
from mininet.net import Mininet
from mininet.node import Controller, RemoteController, OVSKernelSwitch, IVSSwitch, UserSwitch
from mininet.link import Link, TCLink
from mininet.cli import CLI
from mininet.log import setLogLevel
import time
from util.setup import SetupUtil
from util.test import TestUtil

def topology():
    setup = SetupUtil()
    test = TestUtil()
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

    test.test_udp(h1, h4)
    test.test_udp(h4, h1)
    test.test_tcp(h1, h4)
    test.test_tcp(h4, h1)
    test.test_icmp(h1, h4)
    test.test_icmp(h4, h1)
    
    print ("*** Running CLI")
    CLI( net )
    
    print ("*** Stopping network")
    net.stop()
#
#def set_up_interface(host, if_name, ip_addr, net_mask):
#    cmd = 'ifconfig '+str(host.name)+'-'+str(if_name)+' '+str(ip_addr)+' netmask '+str(net_mask)+' up'
#    print("[Set up interface] "+str(host.name)+' '+cmd)
#    host.cmd(cmd)
#
#def set_up_forwarding(host):
#    cmd = 'sysctl -w net.ipv4.ip_forward=1'
#    print("[Set up forwarding] "+str(host.name)+' '+cmd)
#    host.cmd(cmd)
#
#def set_up_default_gw(host, gw):
#    cmd = 'ip route add default via '+str(gw)
#    print("[Set up default gw] "+str(host.name)+' '+cmd)
#    host.cmd(cmd) 
#    cmd = 'ip route change default via '+str(gw)  
#    print("[Set up default gw] "+str(host.name)+' '+cmd) 
#    host.cmd(cmd) 
#
#def set_up_route(host, dest, via):
#    cmd = 'ip route add '+str(dest)+' via '+str(via)
#    print("[Set up route] "+str(host.name)+' '+cmd)
#    host.cmd(cmd)

#def test_tcp(src, dest, srcPort=123, destPort=9999):
#    '''
#    Send a file using nc from src to dest via TCP
#    return success
#    '''
#    test_prefix = "[TCP Test] "
#    print("*** Running TCP Test")
#    print(test_prefix+"src: "+str(src.IP())+" dest: "+str(dest.IP())+" srcPort: "+str(srcPort)+" destPort: "+str(destPort))
#    with open("_test.out", "w") as f:
#        f.write("FAIL")
#    listen_cmd = "timeout 5 nc -t -l "+str(destPort)+" > _test.out"
#    dest.sendCmd(listen_cmd)
#    print(test_prefix+dest.name + ' ' + listen_cmd)
#    # Wait such that nc is listening before something is sent
#    time.sleep(1)
#    send_cmd = "timeout 2 nc -t -p "+str(srcPort)+" "+dest.IP()+" "+str(destPort)+" < _test.in"
#    src.sendCmd(send_cmd)
#    print(test_prefix+src.name + ' ' + send_cmd)
#    src.waitOutput()
#    dest.waitOutput()
#    with open("_test.out", "r") as f:
#        line = f.readline()
#        if line=="SUCCESS":
#            print(test_prefix+"Test successfull!")
#            return True
#        else:
#            print(test_prefix+"Test failed!")
#            return False
#
#def test_udp(src, dest, srcPort=123, destPort=9999):
#    '''
#    Send a file using nc from src to dest via UDP
#    return success
#    '''
#    test_prefix = "[UDP Test] "
#    print("*** Running UDP Test")
#    print(test_prefix+"src: "+str(src.IP())+" dest: "+str(dest.IP())+" srcPort: "+str(srcPort)+" destPort: "+str(destPort))
#    with open("_test.out", "w") as f:
#        f.write("FAIL")
#    listen_cmd = "timeout 5 nc -u -l "+str(destPort)+" > _test.out"
#    dest.sendCmd(listen_cmd)
#    print(test_prefix+dest.name + ' ' + listen_cmd)
#    # Wait such that nc is listening before something is sent
#    time.sleep(1)
#    send_cmd = "timeout 2 nc -u -p "+str(srcPort)+" "+dest.IP()+" "+str(destPort)+" < _test.in"
#    src.sendCmd(send_cmd)
#    print(test_prefix+src.name + ' ' + send_cmd)
#    src.waitOutput()
#    dest.waitOutput()
#    with open("_test.out", "r") as f:
#        line = f.readline()
#        if line=="SUCCESS":
#            print(test_prefix+"Test successfull!")
#            return True
#        else:
#            print(test_prefix+"Test failed!")
#            return False
#    
#def test_icmp(src, dest):
#    '''
#    Let src ping dest via ICMP
#    return success
#    '''
#    print("*** Running ICMP Test")
#    test_prefix = "[ICMP Test] "
#    print(test_prefix+"src: "+str(src.IP())+" dest: "+str(dest.IP()))
#    cmd = "ping -c1 "+str(dest.IP())
#    res = src.cmd(cmd)
#    print(test_prefix+str(src.name)+" "+cmd)
#    sent, received = Mininet._parsePing(res)
#    if sent == received:
#        print(test_prefix+"Test successfull!")
#        return True
#    else:
#        print(test_prefix+"Test failed!")
#        return False

            

if __name__ == '__main__':
    setLogLevel( 'info' )
    topology()