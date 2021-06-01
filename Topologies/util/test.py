from mininet.net import Mininet
import time

class TestUtil:

    def __init__(self):
        pass

    def test_tcp(self, src, dest, srcPort=123, destPort=9999):
        '''
        Send a file using nc from src to dest via TCP
        return success
        '''
        test_prefix = "[TCP Test] "
        print("*** Running TCP Test")
        print(test_prefix+"src: "+str(src.IP())+" dest: "+str(dest.IP())+" srcPort: "+str(srcPort)+" destPort: "+str(destPort))
        with open("test_data/_test.out", "w") as f:
            f.write("FAIL")
        listen_cmd = "timeout 5 nc -t -l "+str(destPort)+" > test_data/_test.out"
        dest.sendCmd(listen_cmd)
        print(test_prefix+dest.name + ' ' + listen_cmd)
        # Wait such that nc is listening before something is sent
        time.sleep(1)
        send_cmd = "timeout 2 nc -t -p "+str(srcPort)+" "+dest.IP()+" "+str(destPort)+" < test_data/_test.in"
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
                return False

    def test_udp(self, src, dest, srcPort=123, destPort=9999):
        '''
        Send a file using nc from src to dest via UDP
        return success
        '''
        test_prefix = "[UDP Test] "
        print("*** Running UDP Test")
        print(test_prefix+"src: "+str(src.IP())+" dest: "+str(dest.IP())+" srcPort: "+str(srcPort)+" destPort: "+str(destPort))
        with open("test_data/_test.out", "w") as f:
            f.write("FAIL")
        listen_cmd = "timeout 5 nc -u -l "+str(destPort)+" > test_data/_test.out"
        dest.sendCmd(listen_cmd)
        print(test_prefix+dest.name + ' ' + listen_cmd)
        # Wait such that nc is listening before something is sent
        time.sleep(1)
        send_cmd = "timeout 2 nc -u -p "+str(srcPort)+" "+dest.IP()+" "+str(destPort)+" < test_data/_test.in"
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
                return False

    def test_icmp(self, src, dest):
        '''
        Let src ping dest via ICMP
        return success
        '''
        print("*** Running ICMP Test")
        test_prefix = "[ICMP Test] "
        print(test_prefix+"src: "+str(src.IP())+" dest: "+str(dest.IP()))
        cmd = "ping -c1 "+str(dest.IP())
        res = src.cmd(cmd)
        print(test_prefix+str(src.name)+" "+cmd)
        sent, received = Mininet._parsePing(res)
        if sent == received:
            print(test_prefix+"Test successfull!")
            return True
        else:
            print(test_prefix+"Test failed!")
            return False
