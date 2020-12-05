package nmap

import (
	"strconv"
	"time"

	"project/internal/system"
)

// Job is contains job information and scanner options.
type Job struct {
	// Protocol is port type, it can be "tcp", "udp" and "all".
	Protocol string `toml:"protocol" json:"protocol"`

	// Target can be IP address or domain name, support IPv4 & IPv6.
	Target string `toml:"target" json:"target"`

	// Ports is the nmap argument, it can be single or a range like "80, 81-1000",
	// but you can't not use argument like "U:53,111,137,T:21-25,80,139,8080,S:9",
	// if you want to scan TCP and UDP, set protocol field "all", because performance.
	// if it is empty, nmap will use the top common ports.
	Ports string `toml:"ports" json:"ports"`

	// Extra is used to store extra information like unit.
	Extra string `toml:"extra" json:"extra"`

	// Options is used to set the special options for this job
	Options *Options `toml:"options" json:"options"`
}

// Options contains job options and scanner default job options.
type Options struct {
	// ---------------------------------basic scan---------------------------------
	// ScanTech is use specified scan technique.
	// TCP Scan default technique is -sS, TCP SYN.
	// UDP Scan will be set automatically if Job.Protocol is "udp" or "all".
	//
	// -sS/sT/sA/sW/sM: TCP SYN/Connect()/ACK/Window/Maimon scans
	// -sU: UDP Scan
	// -sN/sF/sX: TCP Null, FIN, and Xmas scans
	// --scanflags <flags>: Customize TCP scan flags
	// -sI <zombie host[:probe port]>: Idle scan
	// -sY/sZ: SCTP INIT/COOKIE-ECHO scans
	// -sO: IP protocol scan
	// -b <FTP relay host>: FTP bounce scan
	ScanTech string `toml:"scan_tech" json:"scan_tech"`

	// NoPing is the nmap argument "-Pn", treat all hosts as online.
	NoPing bool `toml:"no_ping" json:"no_ping"`

	// Service is the nmap argument "-sV", probe open ports to determine
	// service/version information.
	Service bool `toml:"service" json:"service"`

	// OS is the nmap argument "-O", it is used to enable OS detection.
	OS bool `toml:"os" json:"os"`

	// -------------------------------advanced scan--------------------------------

	// Device is the nmap argument "-e", it is use specified network
	// interface device, you can execute command "nmap -iflist" to
	// get device list.
	Device string `toml:"device" json:"device"`

	// LocalIP is the nmap argument "-S", it is used to specify local
	// IP address, it can spoof source address.
	LocalIP []string `toml:"local_ip" json:"local_ip"`

	// --------------------------------performance---------------------------------

	// HostTimeout is the nmap argument "--host-timeout", it is used to
	// Give up on target after this long.
	HostTimeout time.Duration `toml:"host_timeout" json:"host_timeout"`

	// MaxRTTTimeout is the nmap argument "--max-rtt-timeout", it is
	// used to specifies probe round trip time.
	MaxRTTTimeout time.Duration `toml:"max_rtt_timeout" json:"max_rtt_timeout"`

	// MinRate is the nmap argument "--min-rate", send packets no slower
	// than <number> per second.
	MinRate int `toml:"min_rate" json:"min_rate"`

	// MaxRate is the nmap argument "--max-rate", send packets no faster
	// than <number> per second.
	MaxRate int `toml:"max_rate" json:"max_rate"`

	// ----------------------------------custom------------------------------------

	// Arguments contains extra arguments of nmap, please not conflict
	// with above already exists options.
	Arguments string `toml:"arguments" json:"arguments"`
}

// ToArgs is used to convert options to exec arguments.
func (opts *Options) ToArgs() []string {
	args := make([]string, 0, 8)
	// ---------------------------------basic scan---------------------------------
	if opts.ScanTech != "" {
		args = append(args, opts.ScanTech)
	}
	if opts.NoPing {
		args = append(args, "-Pn")
	}
	if opts.Service {
		args = append(args, "-sV")
	}
	if opts.OS {
		args = append(args, "-O")
	}
	// -------------------------------advanced scan--------------------------------
	if opts.Device != "" {
		args = append(args, "-e", opts.Device)
	}
	if len(opts.LocalIP) != 0 {
		ipList := opts.LocalIP[0]
		for i := 1; i < len(opts.LocalIP); i++ {
			ipList += "," + opts.LocalIP[i]
		}
		args = append(args, "-S", ipList)
	}
	// --------------------------------performance---------------------------------
	if opts.HostTimeout > 0 {
		args = append(args, "--host-timeout", opts.HostTimeout.String())
	}
	if opts.MaxRTTTimeout > 0 {
		args = append(args, "--max-rtt-timeout", opts.MaxRTTTimeout.String())
	}
	if opts.MinRate > 0 {
		args = append(args, "--min-rate", strconv.Itoa(opts.MinRate))
	}
	if opts.MaxRate > 0 {
		args = append(args, "--max-rate", strconv.Itoa(opts.MaxRate))
	}
	// ----------------------------------custom------------------------------------
	if opts.Arguments != "" {
		args = append(args, system.CommandLineToArgv(opts.Arguments)...)
	}
	return args
}
