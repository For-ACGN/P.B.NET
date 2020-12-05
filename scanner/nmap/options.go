package nmap

// Job is contains job information.
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
	// ScanTech is use specified scan technique.
	// TCP Scan default technique is -sS, TCP SYN.
	// UDP Scan will be set automatically if Job.Protocol is "udp" or "all".
	//
	// -sS/sT/sA/sW/sM: TCP SYN/Connect()/ACK/Window/Maimon scans
	// -sU: UDP Scan
	// -sN/sF/sX: TCP Null, FIN, and Xmas scans
	// --scanflags <flags>: Customize TCP scan flags
	// -sI <zombie host[:probeport]>: Idle scan
	// -sY/sZ: SCTP INIT/COOKIE-ECHO scans
	// -sO: IP protocol scan
	// -b <FTP relay host>: FTP bounce scan
	ScanTech string `toml:"scan_tech" json:"scan_tech"`

	// NoPing is the nmap argument "-Pn", treat all hosts as online.
	NoPing bool `toml:"no_ping" json:"no_ping"`

	// Service is the nmap argument "-sV", probe open ports to determine
	// service/version information.
	Service bool

	// OS is the nmap argument "-O", it is used to enable OS detection.
	OS bool

	// Device is the nmap argument "-e", it is use specified network
	// interface device, you can execute command "nmap -iflist" to
	// get device list.
	Device string `toml:"device" json:"device"`

	// LocalIP is the nmap argument "-S", it is used to specify local
	// IP address, it can spoof source address.
	LocalIP []string

	// Arguments contains extra arguments of nmap, please not conflict
	// with above already exists options.
	Arguments string `toml:"arguments" json:"arguments"`
}
