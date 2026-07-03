// Package mdns advertises the daemon on the local network so it is reachable
// at a friendly <host>.local name (e.g. pomo.local).
//
// On a machine already running avahi-daemon (common on Linux desktops), we must
// NOT start a second responder on UDP 5353. Instead we register an address
// record through avahi over D-Bus. Only when avahi is unreachable do we fall
// back to an embedded pure-Go responder.
package mdns

import (
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/godbus/dbus/v5"
	hmdns "github.com/hashicorp/mdns"
)

// Advertise publishes an A record mapping <host>.local to this machine's primary
// IPv4 address and returns a Closer that withdraws the advertisement.
func Advertise(host string, port int) (io.Closer, error) {
	ip, err := primaryIPv4()
	if err != nil {
		return nil, fmt.Errorf("could not determine local IP: %w", err)
	}
	fqdn := host + ".local"

	ad, err := advertiseViaAvahi(fqdn, ip)
	if err == nil {
		slog.Info("mdns: registered via avahi (D-Bus)", "name", fqdn, "ip", ip)
		return ad, nil
	}
	slog.Debug("mdns: avahi unavailable, using embedded responder", "err", err)
	return advertiseEmbedded(host, port, ip)
}

// avahi D-Bus constants (see avahi-common/defs.h).
const (
	avahiIfaceUnspec = int32(-1)
	avahiProtoUnspec = int32(-1)
	// Skip the reverse PTR: this machine's real hostname already owns the PTR
	// for our IP, so claiming it again yields a "Local name collision". We only
	// need the forward <host>.local → IP mapping.
	avahiPublishNoReverse = uint32(16)
)

// advertiseViaAvahi registers an address record with the system avahi-daemon
// over D-Bus, the same mechanism `avahi-publish -a` uses. This lets avahi keep
// ownership of port 5353 so there is no responder conflict.
func advertiseViaAvahi(fqdn string, ip net.IP) (io.Closer, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("system bus: %w", err)
	}

	server := conn.Object("org.freedesktop.Avahi", dbus.ObjectPath("/"))
	var version string
	if err := server.Call("org.freedesktop.Avahi.Server.GetVersionString", 0).Store(&version); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("avahi not reachable: %w", err)
	}

	var groupPath dbus.ObjectPath
	if err := server.Call("org.freedesktop.Avahi.Server.EntryGroupNew", 0).Store(&groupPath); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("EntryGroupNew: %w", err)
	}
	group := conn.Object("org.freedesktop.Avahi", groupPath)

	if call := group.Call("org.freedesktop.Avahi.EntryGroup.AddAddress", 0,
		avahiIfaceUnspec, avahiProtoUnspec, avahiPublishNoReverse, fqdn, ip.String()); call.Err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("AddAddress: %w", call.Err)
	}
	if call := group.Call("org.freedesktop.Avahi.EntryGroup.Commit", 0); call.Err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("commit: %w", call.Err)
	}

	return &avahiAd{conn: conn, group: group}, nil
}

type avahiAd struct {
	conn  *dbus.Conn
	group dbus.BusObject
}

func (a *avahiAd) Close() error {
	a.group.Call("org.freedesktop.Avahi.EntryGroup.Free", 0)
	return a.conn.Close()
}

// advertiseEmbedded runs a pure-Go mDNS responder. Only safe when no other
// responder (avahi) holds port 5353.
func advertiseEmbedded(host string, port int, ip net.IP) (io.Closer, error) {
	svc, err := hmdns.NewMDNSService(
		host, "_http._tcp", "", host+".local.", port, []net.IP{ip}, []string{"pomo daemon"},
	)
	if err != nil {
		return nil, fmt.Errorf("mdns service: %w", err)
	}
	srv, err := hmdns.NewServer(&hmdns.Config{Zone: svc})
	if err != nil {
		return nil, fmt.Errorf("mdns server: %w", err)
	}
	slog.Info("mdns: registered via embedded responder", "name", host+".local", "ip", ip)
	return closerFunc(srv.Shutdown), nil
}

type closerFunc func() error

func (f closerFunc) Close() error { return f() }

// primaryIPv4 returns the machine's primary outbound IPv4 address, skipping
// loopback and common container bridge ranges.
func primaryIPv4() (net.IP, error) {
	// The UDP "connection" sends no packets; it just resolves the route the
	// kernel would use, giving the primary interface address.
	if conn, err := net.Dial("udp", "8.8.8.8:80"); err == nil {
		defer func() { _ = conn.Close() }()
		if ip := conn.LocalAddr().(*net.UDPAddr).IP.To4(); ip != nil && !isContainerIP(ip) {
			return ip, nil
		}
	}

	// Fallback: first non-loopback, non-container IPv4 on any interface.
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}
		if ip := ipnet.IP.To4(); ip != nil && !isContainerIP(ip) {
			return ip, nil
		}
	}
	return nil, fmt.Errorf("no suitable IPv4 address found")
}

// isContainerIP reports whether ip is in the docker default bridge range
// (172.17.0.0/16–172.31.0.0/16) that we don't want to advertise.
func isContainerIP(ip net.IP) bool {
	ip = ip.To4()
	return ip != nil && ip[0] == 172 && ip[1] >= 17 && ip[1] <= 31
}
