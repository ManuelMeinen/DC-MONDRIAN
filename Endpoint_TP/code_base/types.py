from ryu.lib.packet import in_proto
from code_base.const import Const

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
            print("Proto: "+str(self.proto))
        print("------------------------------")
    
    def to_string(self):
        pkt = "Packet <"
        pkt+="Dest IP: "+str(self.destIP)
        pkt+=", Src IP: "+str(self.srcIP)
        if self.destPort == None:
            pkt+=", Dest Port: NONE"
        else:
            pkt+=", Dest Port: "+str(self.destPort)
        if self.srcPort == None:
            pkt+=(", Src Port: NONE")
        else:
            pkt+=", Src Port: "+str(self.srcPort)
        if self.proto == None:
            pkt+=", Proto: NONE"
        else:
            pkt+=", Proto: "+str(self.proto)
        pkt+=">"
        return pkt


class Policy:
    def __init__(self, policyID, action, destZoneID=None, srcZoneID=None, destPort=None, srcPort=None, proto=None):
        self.policyID = policyID
        self.destZoneID = destZoneID
        self.srcZoneID = srcZoneID
        self.destPort = destPort
        self.srcPort = srcPort
        self.proto = proto
        self.action = action

    def print_policy(self):
        print(self.to_string())

    def to_string(self):
        msg = "Policy ID: "+str(self.policyID)+ " <"
        msg += "destZone: "
        if self.destZoneID == None:
            msg += "*"+", "
        else:
            msg += str(self.destZoneID)+", "
        msg += "destPort: "
        if self.destPort == None:
            msg += "*"+", "
        else:
            msg += str(self.destPort)+", "
        msg += "srcZone: "
        if self.srcZoneID == None:
            msg += "*"+", "
        else:
            msg += str(self.srcZoneID)+", "
        msg += "srcPort: "
        if self.srcPort == None:
            msg += "*"+", "
        else:
            msg += str(self.srcPort)+", "
        msg += "proto: "
        if self.proto == None:
            msg += "*"
        else:
            msg += str(self.proto)
        msg += "> --> <"+str(self.action)+">"
        return msg

        


class Subnet:
    def __init__(self, netAddr, zoneID, tpAddr):
        self.netAddr = netAddr
        self.zoneID = zoneID
        self.tpAddr = tpAddr
    
    def print_subnet(self):
        print(self.to_string())
    
    def to_string(self):
        msg = "Subnet: <"+str(self.netAddr)+", "+str(self.zoneID)+", "+str(self.tpAddr)+">"
        return msg

class Zone:
    def __init__(self, zoneID, name):
        self.zoneID = zoneID
        self.name = name
    
    def print_zone(self):
        print(self.to_string())
    
    def to_string(self):
        msg = "Zone: <"+str(self.zoneID)+", "+str(self.name)+">"
        return msg
