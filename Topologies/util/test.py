from mininet.net import Mininet
import time

class TestUtil:

    def __init__(self, prefix=""):
        self.prefix = prefix

    def test_tcp(self, src, dest, srcPort=None, destPort=None, listen_timeout=20):
        '''
        Send a file using nc from src to dest via TCP
        return success
        '''
        max_no_runs = 3
        no_runs = 0
        if srcPort == None:
            str_srcPort = "120"
        else:
            str_srcPort = str(srcPort)
        
        if destPort == None:
            str_destPort = "160"
        else:
            str_destPort = str(destPort)
        test_prefix = self.prefix+"[TCP Test] "
        print(self.prefix+"*** Running TCP Test")
        while max_no_runs>no_runs:
            no_runs += 1
            print(test_prefix+"src: "+str(src.IP())+" dest: "+str(dest.IP())+" srcPort: "+str_srcPort+" destPort: "+str_destPort+" listen_timeout = "+str(listen_timeout))
            with open("test_data/_test.out", "w") as f:
                f.write("FAIL")
            listen_cmd = "timeout "+str(listen_timeout)+" nc -t -l "+str_destPort+" > test_data/_test.out &"
            dest.cmd(listen_cmd)
            print(test_prefix+dest.name + ' ' + listen_cmd)
            # Wait such that nc is listening before something is sent
            time.sleep(3)
            send_cmd = "timeout 10 nc -t -p "+str_srcPort+" "+dest.IP()+" "+str_destPort+" < test_data/_test.in &"
            src.cmd(send_cmd)
            print(test_prefix+src.name + ' ' + send_cmd)
            time.sleep(listen_timeout-3)
            with open("test_data/_test.out", "r") as f:
                line = f.readline()
                if line=="SUCCESS":
                    print(test_prefix+"Test successfull!")
                    return True
                else:
                    print(test_prefix+"Test failed!")
            if no_runs>=1:
                if srcPort==None and destPort==None:
                    print(test_prefix+"Try with other ports")
                    str_srcPort = str(int(str_srcPort)+1)
                    str_destPort = str(int(str_destPort)+1)
                else:
                #NOTE: if the reverse connection was tested then we would need to wait for tcp_fin_timeout=60sec
                    print(test_prefix+"Waiting for 60sec for the port to be released")
                    time.sleep(60)        
        return False

    def test_udp(self, src, dest, srcPort=None, destPort=None, listen_timeout=20):
        '''
        Send a file using nc from src to dest via UDP
        return success
        '''
        max_no_runs = 2
        no_runs = 0
        if srcPort == None:
            str_srcPort = "120"
        else:
            str_srcPort = str(srcPort)
        
        if destPort == None:
            str_destPort = "140"
        else:
            str_destPort = str(destPort)

        test_prefix = self.prefix+"[UDP Test] "
        print(self.prefix+"*** Running UDP Test")
        print(test_prefix+"src: "+str(src.IP())+" dest: "+str(dest.IP())+" srcPort: "+str_srcPort+" destPort: "+str_destPort)
        with open("test_data/_test.out", "w") as f:
            f.write("FAIL")
        while max_no_runs>no_runs: 
            listen_cmd = "timeout "+str(listen_timeout)+" nc -u -l "+str_destPort+" > test_data/_test.out"
            dest.sendCmd(listen_cmd)
            print(test_prefix+dest.name + ' ' + listen_cmd)
            # Wait such that nc is listening before something is sent
            time.sleep(3)
            send_cmd = "timeout 5 nc -u -p "+str_srcPort+" "+dest.IP()+" "+str_destPort+" < test_data/_test.in"
            src.sendCmd(send_cmd)
            print(test_prefix+src.name + ' ' + send_cmd)
            time.sleep(listen_timeout-3)
            src.waitOutput()
            dest.waitOutput()
            with open("test_data/_test.out", "r") as f:
                line = f.readline()
                if line=="SUCCESS":
                    print(test_prefix+"Test successfull!")
                    return True
                else:
                    print(test_prefix+"Test failed!")
                    if no_runs == 0:
                        print(test_prefix+"Run test again because it might be that the data just got corrupted due to unreliable data transfer of UDP.")
                    no_runs += 1
        return False

    def test_icmp(self, src, dest):
        '''
        Let src ping dest via ICMP
        return success
        '''
        print(self.prefix+"*** Running ICMP Test")
        test_prefix = self.prefix+"[ICMP Test] "
        print(test_prefix+"src: "+str(src.IP())+" dest: "+str(dest.IP()))
        cmd = "ping -c5 "+str(dest.IP()) #Send more than one since some could get lost due to unreliable data trasfer
        res = src.cmd(cmd)
        print(test_prefix+str(src.name)+" "+cmd)
        sent, received = Mininet._parsePing(res)
        if received>0:
            print(test_prefix+"Test successfull!")
            return True
        else:
            print(test_prefix+"Test failed!")
            return False

    def test_bandwidth(self, server, client, proto=None, timeout=None):
        print(self.prefix+"*** Running ICMP Test")
        test_prefix = self.prefix+"[ICMP Test] "
        cmd = "iperf -s"
        if proto=="TCP":
            cmd = cmd+" -t"
        if proto=="UDP":
            cmd = cmd+" -u"
        if timeout != None:
            cmd = "timeout "+str(timeout)+" "+cmd
        cmd = cmd+" &"
        print(test_prefix+str(server.name)+" "+cmd)
        server.cmd(cmd)
        cmd = "iperf -c "+str(server.IP())
        if proto=="TCP":
            cmd = cmd+" -t"
        if proto=="UDP":
            cmd = cmd+" -u"
        print(test_prefix+str(client.name)+" "+cmd)
        res = client.cmd(cmd)
        if timeout != None:
            time.sleep(timeout)
        