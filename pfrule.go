package virtualbox

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// PFRule represents a port forwarding rule.
type PFRule struct {
	Proto     PFProto
	HostIP    net.IP // can be nil to match any host interface
	HostPort  uint16
	GuestIP   net.IP // can be nil if guest IP is leased from built-in DHCP
	GuestPort uint16
}

// PFProto represents the protocol of a port forwarding rule.
type PFProto string

const (
	PFTCP = PFProto("tcp")
	PFUDP = PFProto("udp")
)

// String returns a human-friendly representation of the port forwarding rule.
func (r PFRule) String() string {
	hostip := ""
	if r.HostIP != nil {
		hostip = r.HostIP.String()
	}
	guestip := ""
	if r.GuestIP != nil {
		guestip = r.GuestIP.String()
	}
	return fmt.Sprintf("%s://%s:%d --> %s:%d",
		r.Proto, hostip, r.HostPort,
		guestip, r.GuestPort)
}

// Format returns the string needed as a command-line argument to VBoxManage.
func (r PFRule) Format() string {
	hostip := ""
	if r.HostIP != nil {
		hostip = r.HostIP.String()
	}
	guestip := ""
	if r.GuestIP != nil {
		guestip = r.GuestIP.String()
	}
	return fmt.Sprintf("%s,%s,%d,%s,%d", r.Proto, hostip, r.HostPort, guestip, r.GuestPort)
}

func (m *Machine) parseForwarading(key, value string) error {
	if !strings.HasPrefix(key, "Forwarding(") {
		return nil
	}

	vals := strings.Split(value, ",")
	if len(vals) != 6 {
		return badForwarding(key, value, "wrong number of parameters")
	}

	hostip, err := getIP(vals[2])
	if err != nil {
		return badForwarding(err.Error(), key, value)
	}

	hostport, err := strconv.Atoi(vals[3])
	if err != nil {
		return badForwarding(err.Error(), key, value)
	}

	guestip, err := getIP(vals[4])
	if err != nil {
		return badForwarding(err.Error(), key, value)
	}

	guestport, err := strconv.Atoi(vals[5])
	if err != nil {
		return badForwarding(err.Error(), key, value)
	}

	if m.PFRules == nil {
		m.PFRules = make(map[string]PFRule)
	}

	m.PFRules[vals[0]] = PFRule{
		Proto:     PFProto(vals[1]),
		HostIP:    hostip,
		HostPort:  uint16(hostport),
		GuestIP:   guestip,
		GuestPort: uint16(guestport),
	}
	return nil
}

// GetPFRUle get a rule either by name or guest port returning (rule,true) on success
//
// If both name and port is given, it prefer name lookup, either can be empty ("" or 0)
func (m *Machine) GetPFRUle(name string, guestPort uint16) (PFRule, bool) {
	if m.PFRules == nil {
		return PFRule{}, false
	}
	if name != "" {
		if r, ok := m.PFRules[name]; ok {
			return r, true
		}
	}
	if guestPort > 0 {
		for _, r := range m.PFRules {
			if r.GuestPort == guestPort {
				return r, true
			}
		}
	}
	return PFRule{}, false
}

func badForwarding(msg, key, value string) error {
	return fmt.Errorf("Bad Forwarding entry: %s (%s=%s)", msg, key, value)

}

func getIP(s string) (net.IP, error) {
	if s == "" {
		return nil, nil
	}

	n := net.ParseIP(s)
	if n == nil {
		return nil, errors.New("Bad IP")
	}
	return n, nil
}
