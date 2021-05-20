from ryu.base import app_manager
from ryu.controller import ofp_event
from ryu.controller.handler import MAIN_DISPATCHER
from ryu.controller.handler import set_ev_cls
from ryu.ofproto import ofproto_v1_0, ofproto_v1_2, ofproto_v1_3, ofproto_v1_4, ofproto_v1_5

from ryu.lib.packet import packet
from ryu.lib.packet import ethernet
from ryu.lib.packet import ether_types
from ryu.lib.packet import ethernet
from ryu.lib.packet import ether_types
from ryu.lib.packet import in_proto
from ryu.lib.packet import ipv4
from ryu.lib.packet import icmp
from ryu.lib.packet import tcp
from ryu.lib.packet import udp

from code_base.types import Packet, proto_dict, Policy, Zone, Subnet

class EndpointTP(app_manager.RyuApp):
    OFP_VERSIONS = [ofproto_v1_3.OFP_VERSION]

    def debug(self, string):
        print("*** DEBUG INFO *** "+str(string))

    def __init__(self, *args, **kwargs):
        super(EndpointTP, self).__init__(*args, **kwargs)

    @set_ev_cls(ofp_event.EventOFPPacketIn, MAIN_DISPATCHER)
    def packet_in_handler(self, ev):
        
        msg = ev.msg
        datapath = msg.datapath
        ofproto = datapath.ofproto
        parser = datapath.ofproto_parser
        #in_port = msg.match['in_port']

        pkt = packet.Packet(msg.data)
        eth = pkt.get_protocols(ethernet.ethernet)[0]

        #if eth.ethertype == ether_types.ETH_TYPE_LLDP:
        #    # ignore lldp packet
        #    return
        #dst = eth.dst
        #src = eth.src

        #dpid = datapath.id
        #self.mac_to_port.setdefault(dpid, {})

        #self.logger.info("packet in %s %s %s %s", dpid, src, dst, in_port)

        ## learn a mac address to avoid FLOOD next time.
        #self.mac_to_port[dpid][src] = in_port

        #if dst in self.mac_to_port[dpid]:
        #    out_port = self.mac_to_port[dpid][dst]
        #else:
        #    out_port = ofproto.OFPP_FLOOD

        #actions = [parser.OFPActionOutput(out_port)]

        # install a flow to avoid packet_in next time
        #if out_port != ofproto.OFPP_FLOOD:

        # check IP Protocol and create a match for IP
        actions = []

        dstip = None
        srcip = None
        dstport = None
        srcport = None
        proto = None

        if eth.ethertype == ether_types.ETH_TYPE_IP:
            ip = pkt.get_protocol(ipv4.ipv4)
            srcip = ip.src
            dstip = ip.dst
            proto = ip.proto
            
            # if ICMP Protocol
            if proto == in_proto.IPPROTO_ICMP:
                pass
            #  if TCP Protocol
            elif proto == in_proto.IPPROTO_TCP:
                t = pkt.get_protocol(tcp.tcp)
                dstport = t.dst_port
                srcport = t.src_port        
            #  If UDP Protocol 
            elif proto == in_proto.IPPROTO_UDP:
                u = pkt.get_protocol(udp.udp)
                dstport = u.dst_port
                srcport = u.src_port
            
            myPacket = Packet(destIP=dstip, srcIP=srcip, destPort=dstport, srcPort=srcport, proto=proto)
            myPacket.print_packet()

            srcZone = Zone(zoneID=1, name="Zone 1")
            dstZone = Zone(zoneID=2, name="Zone 2")

            myPolicy = Policy(policyID=1, action="forwarding", destZoneID=dstZone.zoneID, srcZoneID=srcZone.zoneID, destPort=123, srcPort=456)
            myPolicy.print_policy()
            srcZone.print_zone()
            mySubnet = Subnet(netAddr="1.2.3.0/24", zoneID=srcZone.zoneID, tpAddr="127.0.0.1")
            mySubnet.print_subnet()
           
        

        