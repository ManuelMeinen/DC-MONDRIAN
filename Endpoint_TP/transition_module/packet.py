from ryu.lib.packet import in_proto

proto_dict = {
    in_proto.IPPROTO_IP:'IPPROTO_IP',
    in_proto.IPPROTO_HOPOPTS:'IPPROTO_HOPOPTS',
    in_proto.IPPROTO_ICMP:'IPPROTO_ICMP',
    in_proto.IPPROTO_IGMP:'IPPROTO_IGMP',
    in_proto.IPPROTO_TCP:'IPPROTO_TCP',
    in_proto.IPPROTO_UDP:'IPPROTO_UDP',
    in_proto.IPPROTO_ROUTING:'IPPROTO_ROUTING',
    in_proto.IPPROTO_FRAGMENT:'IPPROTO_FRAGMENT',
    in_proto.IPPROTO_GRE:'IPPROTO_GRE',
    in_proto.IPPROTO_AH:'IPPROTO_AH',
    in_proto.IPPROTO_ICMPV6:'IPPROTO_ICMPV6',
    in_proto.IPPROTO_NONE:'IPPROTO_NONE',
    in_proto.IPPROTO_DSTOPTS:'IPPROTO_DSTOPTS',
    in_proto.IPPROTO_OSPF:'IPPROTO_OSPF',
    in_proto.IPPROTO_VRRP:'IPPROTO_VRRP',
    in_proto.IPPROTO_SCTP:'IPPROTO_SCTP'
}

class Packet:
    
    def __init__(self, destIP, srcIP, destPort=None, srcPort=None, proto=None):
        self.destIP = destIP
        self.srcIP = srcIP
        self.destPort = destPort
        self.srcPort = srcPort
        self.proto = proto

    def print_packet(self):
        print("----------- PACKET -----------")
        print("Dest IP: "+str(self.destIP))
        print("Src IP: "+str(self.srcIP))
        if self.destPort == None:
            print("Dest Port: NONE")
        else:
            print("Dest Port: "+str(self.destPort))
        if self.srcPort == None:
            print("Src Port: NONE")
        else:
            print("Src Port: "+str(self.srcPort))
        if self.proto == None:
            print("Proto: NONE")
        else:
            print("Proto: "+str(proto_dict[self.proto]))
        print("------------------------------")
