
import time


class SetupUtil:
    
    def __init__(self):
        pass

    def set_up_interface(self, host, if_name, ip_addr, net_mask):
        '''
        Set up an interface on the host
        '''
        cmd = 'ifconfig '+str(host.name)+'-'+str(if_name)+' '+str(ip_addr)+' netmask '+str(net_mask)+' up'
        print("[Set up interface] "+str(host.name)+' '+cmd)
        host.cmd(cmd)

    def set_up_forwarding(self, host):
        '''
        Set up IP-forwarding
        '''
        cmd = 'sysctl -w net.ipv4.ip_forward=1'
        print("[Set up forwarding] "+str(host.name)+' '+cmd)
        host.cmd(cmd)

    def set_up_default_gw(self, host, gw):
        '''
        Set up the default gateway
        '''
        cmd = 'ip route add default via '+str(gw)
        print("[Set up default gw] "+str(host.name)+' '+cmd)
        host.cmd(cmd) 
        cmd = 'ip route change default via '+str(gw)  
        print("[Set up default gw] "+str(host.name)+' '+cmd) 
        host.cmd(cmd) 

    def set_up_route(self, host, dest, via):
        '''
        Add static routing information
        '''
        cmd = 'ip route add '+str(dest)+' via '+str(via)
        print("[Set up route] "+str(host.name)+' '+cmd)
        host.cmd(cmd)
    
    def set_up_inet(self, host, if_name, ip_addr, net_mask):
        '''
        Set up an interface on the host
        '''
        cmd = 'ifconfig '+str(if_name)+' '+str(ip_addr)+' netmask '+str(net_mask)+' up'
        print("[Set up interface] "+str(host.name)+' '+cmd)
        host.cmd(cmd)
        gw = ip_addr[:-1]+"1"
        cmd = 'ip route add default via '+str(gw)
        print("[Set up default gw] "+str(host.name)+' '+cmd)
        host.cmd(cmd) 
        cmd = 'ip route change default via '+str(gw)  
        print("[Set up default gw] "+str(host.name)+' '+cmd) 
        host.cmd(cmd) 

    def prepare_gateway_TP(self, host, tp_addr):
        '''
        Compile and prepare /Gateway_TP/main
        '''
        cmd = "cd /Gateway_TP"
        print("[Start Gateway TP] "+str(host.name)+' '+cmd)
        host.cmd(cmd)
        cmd = 'python3 create_config.py '+str(host.name)+" "+str(tp_addr)+" > /vol1/log/"+str(host.name)+".txt"
        print("[Start Gateway TP] "+str(host.name)+' '+cmd)
        host.cmd(cmd)
        cmd = "go build main.go"+" >> /vol1/log/"+str(host.name)+".txt"
        print("[Start Gateway TP] "+str(host.name)+' '+cmd)
        host.cmd(cmd)
    
    def start_gateway_TP(self, host):
        '''
        Start /Gateway_TP/main
        '''
        cmd = "/Gateway_TP/main >> /vol1/log/"+str(host.name)+".txt &"
        print("[Start Gateway TP] "+str(host.name)+' '+cmd)
        time.sleep(2) # Probably more reliable... sometimes it doesn't get started
        host.sendCmd(cmd)
