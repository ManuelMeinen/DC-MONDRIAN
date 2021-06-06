from mininet.net import Mininet
import time

class TestUtil:

    def __init__(self, prefix=""):
        self.prefix = prefix

    def test_tcp(self, src, dest, srcPort=None, destPort=None):
        '''
        Send a file using nc from src to dest via TCP
        return success
        '''
        max_no_runs = 2
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
            print(test_prefix+"src: "+str(src.IP())+" dest: "+str(dest.IP())+" srcPort: "+str_srcPort+" destPort: "+str_destPort)
            with open("test_data/_test.out", "w") as f:
                f.write("FAIL")
            listen_cmd = "timeout 5 nc -t -l "+str_destPort+" > test_data/_test.out"
            dest.sendCmd(listen_cmd)
            print(test_prefix+dest.name + ' ' + listen_cmd)
            # Wait such that nc is listening before something is sent
            time.sleep(3)
            send_cmd = "timeout 2 nc -t -p "+str_srcPort+" "+dest.IP()+" "+str_destPort+" < test_data/_test.in"
            src.sendCmd(send_cmd)
            print(test_prefix+src.name + ' ' + send_cmd)
            src_out = src.waitOutput()
            dest_out = dest.waitOutput()
            if src_out != "" or dest_out != "":
                if src_out != "":
                    print(test_prefix+src.name+": "+src_out[0:-2])
                else:
                    print(test_prefix+dest.name+": "+dest_out[0:-2])
                #NOTE: if the reverse connection was tested then we would need to wait for tcp_fin_timeout=60sec
                print(test_prefix+"Waiting for 60sec for the port to be released")
                time.sleep(60)
                continue
            else:
                with open("test_data/_test.out", "r") as f:
                    line = f.readline()
                    if line=="SUCCESS":
                        print(test_prefix+"Test successfull!")
                        return True
                    else:
                        print(test_prefix+"Test failed!")
                        return False

    def test_udp(self, src, dest, srcPort=None, destPort=None):
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
            listen_cmd = "timeout 5 nc -u -l "+str_destPort+" > test_data/_test.out"
            dest.sendCmd(listen_cmd)
            print(test_prefix+dest.name + ' ' + listen_cmd)
            # Wait such that nc is listening before something is sent
            time.sleep(1)
            send_cmd = "timeout 2 nc -u -p "+str_srcPort+" "+dest.IP()+" "+str_destPort+" < test_data/_test.in"
            src.sendCmd(send_cmd)
            print(test_prefix+src.name + ' ' + send_cmd)
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
