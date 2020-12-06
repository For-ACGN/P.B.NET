package nmap

import (
	"strconv"
	"strings"

	"project/internal/system"
)

// Options contains job options and scanner default job options.
type Options struct {
	// ---------------------------------basic scan---------------------------------
	// ScanTech is use specified scan technique.
	// TCP Scan default technique is -sS, TCP SYN.
	// UDP Scan will be set automatically if Job.Protocol is "udp",
	// don't set it again.
	// Must not add "-" before it, we added it.
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
	//
	// In Scanner default option, it will random select one to the next
	// scan job if it Options is nil.
	LocalIP []string `toml:"local_ip" json:"local_ip"`

	// --------------------------------performance---------------------------------

	// HostTimeout is the nmap argument "--host-timeout", it is used to
	// Give up on target after this long.
	HostTimeout string `toml:"host_timeout" json:"host_timeout"`

	// MaxRTTTimeout is the nmap argument "--max-rtt-timeout", it is
	// used to specifies probe round trip time.
	MaxRTTTimeout string `toml:"max_rtt_timeout" json:"max_rtt_timeout"`

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
		args = append(args, "-"+opts.ScanTech)
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
		args = append(args, "-S", opts.LocalIP[0])
	}
	// --------------------------------performance---------------------------------
	if opts.HostTimeout != "" {
		args = append(args, "--host-timeout", opts.HostTimeout)
	}
	if opts.MaxRTTTimeout != "" {
		args = append(args, "--max-rtt-timeout", opts.MaxRTTTimeout)
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

// String is used to print command line.
func (opts *Options) String() string {
	return strings.Join(opts.ToArgs(), " ")
}
