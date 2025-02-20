package ebpf

import (
	"time"
)

// Config stores all flags used by the eBPF tracer
type Config struct {
	// CollectTCPConns specifies whether the tracer should collect traffic statistics for TCP connections
	CollectTCPConns bool

	// CollectUDPConns specifies whether the tracer should collect traffic statistics for UDP connections
	CollectUDPConns bool

	// CollectIPv6Conns specifics whether the tracer should capture traffic for IPv6 TCP/UDP connections
	CollectIPv6Conns bool

	// CollectLocalDNS specifies whether the tracer should capture traffic for local DNS calls
	CollectLocalDNS bool

	// DNSInspection specifies whether the tracer should enhance connection data with domain names by inspecting DNS traffic
	// Notice this does *not* depend on CollectLocalDNS
	DNSInspection bool

	// UDPConnTimeout determines the length of traffic inactivity between two (IP, port)-pairs before declaring a UDP
	// connection as inactive.
	// Note: As UDP traffic is technically "connection-less", for tracking, we consider a UDP connection to be traffic
	//       between a source and destination IP and port.
	UDPConnTimeout time.Duration

	// TCPConnTimeout is like UDPConnTimeout, but for TCP connections. TCP connections are cleared when
	// the BPF module receives a tcp_close call, but TCP connections also age out to catch cases where
	// tcp_close is not intercepted for some reason.
	TCPConnTimeout time.Duration

	// MaxTrackedConnections specifies the maximum number of connections we can track, this will be the size of the eBPF + Conntrack.
	MaxTrackedConnections uint

	// MaxClosedConnectionsBuffered represents the maximum number of closed connections we'll buffer in memory. These closed connections
	// get flushed on every client request (default 30s check interval)
	MaxClosedConnectionsBuffered int

	// MaxConnectionsStateBuffered represents the maximum number of state objects that we'll store in memory. These state objects store
	// the stats for a connection so we can accurately determine traffic change between client requests.
	MaxConnectionsStateBuffered int

	// ClientStateExpiry specifies the max time a client (e.g. process-agent)'s state will be stored in memory before being evicted.
	ClientStateExpiry time.Duration

	// ProcRoot is the root path to the proc filesystem
	ProcRoot string

	// BPFDebug enables bpf debug logs
	BPFDebug bool

	// EnableConntrack enables probing conntrack for network address translation via netlink
	EnableConntrack bool

	// ConntrackShortTermBufferSize is the maximum number of short term conntracked connections that will
	// held in memory at once
	ConntrackShortTermBufferSize int

	// DebugPort specifies a port to run golang's expvar and pprof debug endpoint
	DebugPort int

	// ClosedChannelSize specifies the size for closed channel for the tracer
	ClosedChannelSize int

	// ExcludedSourceConnections is a map of source connections to blacklist
	ExcludedSourceConnections map[string][]string

	// ExcludedDestinationConnections is a map of destination connections to blacklist
	ExcludedDestinationConnections map[string][]string
}

// NewDefaultConfig enables traffic collection for all connection types
func NewDefaultConfig() *Config {
	return &Config{
		CollectTCPConns:       true,
		CollectUDPConns:       true,
		CollectIPv6Conns:      true,
		CollectLocalDNS:       false,
		DNSInspection:         true,
		UDPConnTimeout:        30 * time.Second,
		TCPConnTimeout:        2 * time.Minute,
		MaxTrackedConnections: 65536,
		ProcRoot:              "/proc",
		BPFDebug:              false,
		EnableConntrack:       true,
		// With clients checking connection stats roughly every 30s, this gives us roughly ~1.6k + ~2.5k objects a second respectively.
		MaxClosedConnectionsBuffered: 50000,
		MaxConnectionsStateBuffered:  75000,
		ClientStateExpiry:            2 * time.Minute,
		ClosedChannelSize:            500,
	}
}

// EnabledKProbes returns a map of kprobes that are enabled per config settings.
// This map does not include the probes used exclusively in the offset guessing process.
func (c *Config) EnabledKProbes(isRHELOrCentos bool) map[KProbeName]struct{} {
	enabled := make(map[KProbeName]struct{}, 0)

	if c.CollectTCPConns {
		if isRHELOrCentos {
			enabled[TCPSendMsgRHEL] = struct{}{}
		} else {
			enabled[TCPSendMsg] = struct{}{}
		}
		enabled[TCPCleanupRBuf] = struct{}{}
		enabled[TCPClose] = struct{}{}
		enabled[TCPRetransmit] = struct{}{}
		enabled[InetCskAcceptReturn] = struct{}{}
		enabled[TCPv4DestroySock] = struct{}{}

		if c.BPFDebug {
			enabled[TCPSendMsgReturn] = struct{}{}
		}
	}

	if c.CollectUDPConns {
		enabled[UDPRecvMsgReturn] = struct{}{}
		enabled[UDPRecvMsg] = struct{}{}
		if isRHELOrCentos {
			enabled[UDPSendMsgRHEL] = struct{}{}
		} else {
			enabled[UDPSendMsg] = struct{}{}
		}

	}

	return enabled
}
