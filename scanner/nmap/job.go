package nmap

import (
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/net/idna"

	"project/internal/nettool"
)

// Job is contains job information and scanner options.
type Job struct {
	// Protocol is port type, it can be "tcp" and "udp".
	Protocol string `toml:"protocol" json:"protocol"`

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

	// Target can be IP address or domain name, support IPv4 & IPv6. It is NOT
	// support IP range or CIDR, only support single IP or domain name, if you
	// want to scan a host list[not recommend], not use this field and add host
	// to the end of Options.Arguments or use "-iL" argument.
	Target string `toml:"target" json:"target"`

	// Port is the nmap argument, it can be single or a range like "80, 81-1000",
	// but you can't not use argument like "U:53,111,137,T:21-25,80,139,8080,S:9",
	// if you want to scan TCP and UDP, set protocol field "all", because improve
	// performance. if it is empty, nmap will use the top common ports.
	Port string `toml:"port" json:"port"`

	// Extra is used to store extra information like unit.
	// It is not the nmap argument.
	Extra string `toml:"extra" json:"extra"`

	// Options is used to set the special options for this job.
	Options *Options `toml:"options" json:"options" testsuite:"-"`

	// --------------------------------inner used----------------------------------

	// isScanner is used to specify it is scanner default job option.
	isScanner bool

	// outputPath is the output path, scanner will set it in scanner.process()
	outputPath string
}

// ToArgs is used to convert Job config to exec arguments.
func (job *Job) ToArgs() ([]string, error) {
	scanTech, err := job.selectScanTech()
	if err != nil {
		return nil, err
	}
	target, needResolve, err := job.checkTarget()
	if err != nil {
		return nil, err
	}
	port, err := job.checkPort()
	if err != nil {
		return nil, err
	}
	args := make([]string, 0, 3+8)
	// set scan technique
	if scanTech != "" {
		args = append(args, "-"+scanTech)
	}
	// set scan port range
	if port != "" {
		args = append(args, "-p", port)
	}
	// set options
	if job.Options != nil {
		args = append(args, job.Options.ToArgs()...)
	}
	// set output path
	args = append(args, "-oX", job.outputPath)
	// set scan target
	if target != "" {
		if !needResolve {
			args = append(args, "-n")
		}
		args = append(args, target)
	}
	return args, nil
}

func (job *Job) selectScanTech() (string, error) {
	var scanTech string
	switch strings.ToLower(job.Protocol) {
	case "tcp":
		if job.ScanTech != "" {
			if job.ScanTech == "sU" {
				return "", errors.New("invalid TCP scan technique: sU")
			}
			scanTech = job.ScanTech
		} else {
			scanTech = "sS"
		}
	case "udp":
		if job.ScanTech != "" && job.ScanTech != "sU" {
			return "", errors.New("UDP scan not support technique field except sU")
		}
		scanTech = "sU"
	case "":
		return "", errors.New("protocol is empty")
	default:
		return "", errors.Errorf("invalid protocol: \"%s\"", job.Protocol)
	}
	return scanTech, nil
}

// return target, need resolve domain and error.
func (job *Job) checkTarget() (string, bool, error) {
	if job.Target == "" {
		return "", false, nil
	}
	// check target is a IP address
	ip := net.ParseIP(job.Target)
	if ip != nil {
		return ip.String(), false, nil
	}
	// check target is a domain name
	domain, _ := idna.ToASCII(job.Target)
	if !isDomainName(domain) {
		return "", false, errors.Errorf(job.Target + " is not a valid domain name")
	}
	return domain, true, nil
}

// copy from internal/dns/protocol.go
func isDomainName(s string) bool {
	l := len(s)
	if l == 0 || l > 254 || l == 254 && s[l-1] != '.' {
		return false
	}
	last := byte('.')
	nonNumeric := false // true once we've seen a letter or hyphen
	partLen := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		ok := false
		checkChar(c, last, &nonNumeric, &partLen, &ok)
		if !ok {
			return false
		}
		last = c
	}
	if last == '-' || partLen > 63 {
		return false
	}
	return nonNumeric
}

func checkChar(c byte, last byte, nonNumeric *bool, partLen *int, ok *bool) {
	switch {
	case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_':
		*nonNumeric = true
		*partLen++
		*ok = true
	case '0' <= c && c <= '9':
		// fine
		*partLen++
		*ok = true
	case c == '-':
		// Byte before dash cannot be dot.
		if last == '.' {
			return
		}
		*partLen++
		*nonNumeric = true
		*ok = true
	case c == '.':
		// Byte before dot cannot be dot, dash.
		if last == '.' || last == '-' {
			return
		}
		if *partLen > 63 || *partLen == 0 {
			return
		}
		*partLen = 0
		*ok = true
	}
}

func (job *Job) checkPort() (string, error) {
	if job.Port == "" {
		return "", nil
	}
	for _, ports := range strings.Split(strings.ReplaceAll(job.Port, " ", ""), ",") {
		p := strings.SplitN(ports, "-", 2)
		switch len(p) {
		case 0:
			return "", errors.Errorf("invalid port range: \"%s\"", ports)
		case 1: // single port like "80"
			err := nettool.CheckPortString(p[0])
			if err != nil {
				return "", err
			}
		case 2: // port range like "81-82"
			err := nettool.CheckPortString(p[0])
			if err != nil {
				return "", err
			}
			err = nettool.CheckPortString(p[1])
			if err != nil {
				return "", err
			}
			begin, _ := strconv.Atoi(p[0])
			end, _ := strconv.Atoi(p[1])
			if begin > end {
				const format = "invalid port begin or end range: %d-%d"
				return "", errors.Errorf(format, begin, end)
			}
		}
	}
	return job.Port, nil
}

// String is used to print command line.
func (job *Job) String() string {
	args, err := job.ToArgs()
	if err != nil {
		return err.Error()
	}
	return strings.Join(args, " ")
}
