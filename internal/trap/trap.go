package trap

import (
	"fmt"
	g "github.com/gosnmp/gosnmp"
	"github.com/sirupsen/logrus"
	"log"
	"net"
	"strings"
)

type Listener struct {
	trapListener      *g.TrapListener
	listenAddr        string
	community         string
	strictByCommunity bool
	publish           func(data interface{}) error
}

func NewTrapListener(listenAddress, community string, strictByCommunity bool) *Listener {
	listener := &Listener{}
	tl := g.NewTrapListener()

	g.Default.Community = community
	tl.Params = g.Default
	tl.OnNewTrap = listener.myTrapHandler

	listener.listenAddr = listenAddress
	listener.community = community
	listener.strictByCommunity = strictByCommunity
	listener.trapListener = tl

	return listener
}

func (l *Listener) SetPublisher(publisher func(data interface{}) error) {
	l.publish = publisher

}

func (l *Listener) Listen() error {
	err := l.trapListener.Listen(l.listenAddr)
	if err != nil {
		log.Panicf("error in listen: %s", err)
	}
	return err
}

type Trap struct {
	Host        string                            `json:"host"`
	Version     string                            `json:"version"`
	Community   string                            `json:"community"`
	ObjectIdent string                            `json:"object"`
	TimeTicks   uint                              `json:"timeticks"`
	Data        map[string]map[string]interface{} `json:"data"`
}

func NewTrap() *Trap {
	return &Trap{
		Host:      "",
		Version:   "",
		Community: "",
		Data:      make(map[string]map[string]interface{}),
	}
}

func (l *Listener) myTrapHandler(packet *g.SnmpPacket, addr *net.UDPAddr) {
	if l.community != "" && packet.Community != l.community {
		logrus.Errorf("got trapdata from %s with community %s, waiting for community %s", addr.IP, packet.Community, g.Default.Community)
		return
	}
	trap := NewTrap()
	trap.Host = addr.IP.String()
	trap.Version = packet.Version.String()
	dataRaw := make(map[string]map[string]interface{})
	trap.Community = packet.Community
	trap.Data = dataRaw
	logrus.Debugf("new packet received from %s with version %s and community %v", addr.IP, packet.Version, packet.Community)
	for _, v := range packet.Variables {
		if _, ok := dataRaw[v.Name]; !ok {
			dataRaw[v.Name] = make(map[string]interface{})
		}
		dataRaw[v.Name]["type"] = "UnknownType"
		mustDelete := false
		switch v.Type {
		case g.EndOfContents:
			dataRaw[v.Name]["type"] = "EndOfContents"
		case g.Boolean:
			b := v.Value.(bool)
			dataRaw[v.Name]["value"] = b
			dataRaw[v.Name]["type"] = "Boolean"
		case g.Integer:
			i := v.Value.(int)
			dataRaw[v.Name]["value"] = i
			dataRaw[v.Name]["type"] = "Integer"
		case g.BitString:
			b := v.Value.([]byte)
			dataRaw[v.Name]["value"] = string(b)
			dataRaw[v.Name]["type"] = "BitString"
		case g.OctetString:
			b := v.Value.([]byte)
			dataRaw[v.Name]["value"] = string(b)
			dataRaw[v.Name]["type"] = "OctetString"
			dataRaw[v.Name]["hex"] = BytesToHexSeparated(b, ":")
		case g.Null:
			dataRaw[v.Name]["value"] = nil
			dataRaw[v.Name]["type"] = "Null"
		case g.ObjectIdentifier:
			oid := v.Value.(string)
			dataRaw[v.Name]["value"] = oid
			dataRaw[v.Name]["type"] = "ObjectIdentifier"
			mustDelete = true
			trap.ObjectIdent = oid
		case g.ObjectDescription:
			desc := v.Value.(string)
			dataRaw[v.Name]["value"] = desc
			dataRaw[v.Name]["type"] = "ObjectDescription"
			dataRaw[v.Name]["hex"] = BytesToHexSeparated([]byte(desc), ":")
		case g.IPAddress:
			ip := v.Value.(string)
			dataRaw[v.Name]["value"] = ip
			dataRaw[v.Name]["type"] = "IPAddress"
			dataRaw[v.Name]["hex"] = BytesToHexSeparated([]byte(ip), ":")
		case g.Counter32:
			c := v.Value.(uint32)
			dataRaw[v.Name]["value"] = c
			dataRaw[v.Name]["type"] = "Counter32"
		case g.Gauge32:
			g32 := v.Value.(uint)
			dataRaw[v.Name]["value"] = g32
			dataRaw[v.Name]["type"] = "Gauge32"
		case g.TimeTicks:
			t := v.Value.(uint32)
			dataRaw[v.Name]["value"] = t
			dataRaw[v.Name]["type"] = "TimeTicks"
			mustDelete = true
			trap.TimeTicks = uint(t)
		case g.Opaque:
			o := v.Value.([]byte)
			dataRaw[v.Name]["value"] = string(o)
			dataRaw[v.Name]["type"] = "Opaque"
		case g.NsapAddress:
			addr := v.Value.(string)
			dataRaw[v.Name]["value"] = addr
			dataRaw[v.Name]["type"] = "NsapAddress"
		case g.Counter64:
			c := v.Value.(uint64)
			dataRaw[v.Name]["value"] = c
			dataRaw[v.Name]["type"] = "Counter64"
		case g.Uinteger32:
			u := v.Value.(uint32)
			dataRaw[v.Name]["value"] = u
			dataRaw[v.Name]["type"] = "UnsignedInteger32"
		case g.OpaqueFloat:
			f := v.Value.(float32)
			dataRaw[v.Name]["value"] = f
			dataRaw[v.Name]["type"] = "OpaqueFloat"
		case g.OpaqueDouble:
			d := v.Value.(float64)
			dataRaw[v.Name]["value"] = d
			dataRaw[v.Name]["type"] = "OpaqueDouble"
		case g.NoSuchObject:
			dataRaw[v.Name]["value"] = nil
			dataRaw[v.Name]["type"] = "NoSuchObject"
		case g.NoSuchInstance:
			dataRaw[v.Name]["value"] = nil
			dataRaw[v.Name]["type"] = "NoSuchInstance"
		case g.EndOfMibView:
			dataRaw[v.Name]["value"] = nil
			dataRaw[v.Name]["type"] = "EndOfMibView"
		}
		logrus.Debugf("%v -> %v = %v: %v", addr.IP, v.Name, dataRaw[v.Name]["type"], dataRaw[v.Name]["value"])
		if mustDelete {
			delete(dataRaw, v.Name)
		}
	}
	if l.publish == nil {
		return
	}
	err := l.publish(trap)
	if err != nil {
		logrus.Errorf("error sending trap to handler: %v", err)
	}
}

func BytesToHexSeparated(data []byte, sep string) string {
	hexString := fmt.Sprintf("%x", data)
	var result strings.Builder
	for i := 0; i < len(hexString); i += 2 {
		if i > 0 {
			result.WriteString(sep)
		}
		result.WriteString(hexString[i : i+2])
	}
	return result.String()
}
