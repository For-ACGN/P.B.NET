package nmap

import (
	"encoding/xml"
	"strconv"
	"time"
)

// Timestamp represents time as a UNIX timestamp in seconds.
type Timestamp time.Time

// str2time converts a string containing a UNIX timestamp to to a time.Time.
func (t *Timestamp) str2time(s string) error {
	ts, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*t = Timestamp(time.Unix(ts, 0))
	return nil
}

// time2str formats the time.Time value as a UNIX timestamp string.
// XXX these might also need to be changed to pointers. See str2time
// and UnmarshalXMLAttr.
func (t Timestamp) time2str() string {
	return strconv.FormatInt(time.Time(t).Unix(), 10)
}

// MarshalJSON is used to implement json.Marshaler interface.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(t.time2str()), nil
}

// UnmarshalJSON is used to implement json.Unmarshaler interface.
func (t *Timestamp) UnmarshalJSON(b []byte) error {
	return t.str2time(string(b))
}

// MarshalXMLAttr is used to implement xml.MarshalerAttr interface.
func (t Timestamp) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: t.time2str()}, nil
}

// UnmarshalXMLAttr is used to implement xml.UnmarshalerAttr interface.
func (t *Timestamp) UnmarshalXMLAttr(attr xml.Attr) (err error) {
	return t.str2time(attr.Value)
}

// Output contains all the data for a single nmap scan.
type Output struct {
	Scanner          string         `xml:"scanner,attr" json:"scanner"`
	Args             string         `xml:"args,attr" json:"args"`
	Start            Timestamp      `xml:"start,attr" json:"start"`
	StartStr         string         `xml:"startstr,attr" json:"startstr"`
	Version          string         `xml:"version,attr" json:"version"`
	ProfileName      string         `xml:"profile_name,attr" json:"profile_name"`
	XMLOutputVersion string         `xml:"xmloutputversion,attr" json:"xmloutputversion"`
	ScanInfo         ScanInfo       `xml:"scaninfo" json:"scaninfo"`
	Verbose          Verbose        `xml:"verbose" json:"verbose"`
	Debugging        Debugging      `xml:"debugging" json:"debugging"`
	TaskBegin        []Task         `xml:"taskbegin" json:"taskbegin"`
	TaskProgress     []TaskProgress `xml:"taskprogress" json:"taskprogress"`
	TaskEnd          []Task         `xml:"taskend" json:"taskend"`
	PreScripts       []Script       `xml:"prescript>script" json:"prescripts"`
	PostScripts      []Script       `xml:"postscript>script" json:"postscripts"`
	Hosts            []Host         `xml:"host" json:"hosts"`
	Targets          []Target       `xml:"target" json:"targets"`
	RunStats         RunStats       `xml:"runstats" json:"runstats"`
}

// ScanInfo contains informational regarding how the scan was run.
type ScanInfo struct {
	Type        string `xml:"type,attr" json:"type"`
	Protocol    string `xml:"protocol,attr" json:"protocol"`
	NumServices int    `xml:"numservices,attr" json:"numservices"`
	Services    string `xml:"services,attr" json:"services"`
	ScanFlags   string `xml:"scanflags,attr" json:"scanflags"`
}

// Verbose contains the verbosity level for the Nmap scan.
type Verbose struct {
	Level int `xml:"level,attr" json:"level"`
}

// Debugging contains the debugging level for the Nmap scan.
type Debugging struct {
	Level int `xml:"level,attr" json:"level"`
}

// Task contains information about started and stopped Nmap tasks.
type Task struct {
	Task      string    `xml:"task,attr" json:"task"`
	Time      Timestamp `xml:"time,attr" json:"time"`
	ExtraInfo string    `xml:"extrainfo,attr" json:"extrainfo"`
}

// TaskProgress contains information about the progression of a Task.
type TaskProgress struct {
	Task      string    `xml:"task,attr" json:"task"`
	Time      Timestamp `xml:"time,attr" json:"time"`
	Percent   float32   `xml:"percent,attr" json:"percent"`
	Remaining int       `xml:"remaining,attr" json:"remaining"`
	Etc       Timestamp `xml:"etc,attr" json:"etc"`
}

// Target is found in the Nmap xml spec. I have no idea what it actually is.
type Target struct {
	Specification string `xml:"specification,attr" json:"specification"`
	Status        string `xml:"status,attr" json:"status"`
	Reason        string `xml:"reason,attr" json:"reason"`
}

// Host contains all information about a single host.
type Host struct {
	StartTime     Timestamp     `xml:"starttime,attr" json:"starttime"`
	EndTime       Timestamp     `xml:"endtime,attr" json:"endtime"`
	Comment       string        `xml:"comment,attr" json:"comment"`
	Status        Status        `xml:"status" json:"status"`
	Addresses     []Address     `xml:"address" json:"addresses"`
	Hostnames     []Hostname    `xml:"hostnames>hostname" json:"hostnames"`
	Smurfs        []Smurf       `xml:"smurf" json:"smurfs"`
	Ports         []Port        `xml:"ports>port" json:"ports"`
	ExtraPorts    []ExtraPorts  `xml:"ports>extraports" json:"extraports"`
	Os            Os            `xml:"os" json:"os"`
	Distance      Distance      `xml:"distance" json:"distance"`
	Uptime        Uptime        `xml:"uptime" json:"uptime"`
	TcpSequence   TcpSequence   `xml:"tcpsequence" json:"tcpsequence"`
	IpIdSequence  IpIdSequence  `xml:"ipidsequence" json:"ipidsequence"`
	TcpTsSequence TcpTsSequence `xml:"tcptssequence" json:"tcptssequence"`
	HostScripts   []Script      `xml:"hostscript>script" json:"hostscripts"`
	Trace         Trace         `xml:"trace" json:"trace"`
	Times         Times         `xml:"times" json:"times"`
}

// Status is the host's status. Up, down, etc.
type Status struct {
	State     string  `xml:"state,attr" json:"state"`
	Reason    string  `xml:"reason,attr" json:"reason"`
	ReasonTTL float32 `xml:"reason_ttl,attr" json:"reason_ttl"`
}

// Address contains a IPv4 or IPv6 address for a Host.
type Address struct {
	Addr     string `xml:"addr,attr" json:"addr"`
	AddrType string `xml:"addrtype,attr" json:"addrtype"`
	Vendor   string `xml:"vendor,attr" json:"vendor"`
}

// Hostname is a single name for a Host.
type Hostname struct {
	Name string `xml:"name,attr" json:"name"`
	Type string `xml:"type,attr" json:"type"`
}

// Smurf contains repsonses from a smurf attack. I think. Smurf attacks, really?
type Smurf struct {
	Responses string `xml:"responses,attr" json:"responses"`
}

// ExtraPorts contains the information about the closed|filtered ports.
type ExtraPorts struct {
	State   string   `xml:"state,attr" json:"state"`
	Count   int      `xml:"count,attr" json:"count"`
	Reasons []Reason `xml:"extrareasons" json:"reasons"`
}

// Reason is the ExtraPorts reason.
type Reason struct {
	Reason string `xml:"reason,attr" json:"reason"`
	Count  int    `xml:"count,attr" json:"count"`
}

// Port contains all the information about a scanned port.
type Port struct {
	Protocol string   `xml:"protocol,attr" json:"protocol"`
	PortId   int      `xml:"portid,attr" json:"id"`
	State    State    `xml:"state" json:"state"`
	Owner    Owner    `xml:"owner" json:"owner"`
	Service  Service  `xml:"service" json:"service"`
	Scripts  []Script `xml:"script" json:"scripts"`
}

// State contains information about a given ports status.
// State will be open, closed, etc.
type State struct {
	State     string  `xml:"state,attr" json:"state"`
	Reason    string  `xml:"reason,attr" json:"reason"`
	ReasonTTL float32 `xml:"reason_ttl,attr" json:"reason_ttl"`
	ReasonIP  string  `xml:"reason_ip,attr" json:"reason_ip"`
}

// Owner contains the name of Port.Owner.
type Owner struct {
	Name string `xml:"name,attr" json:"name"`
}

// Service contains detailed information about a Port's service details.
type Service struct {
	Name       string `xml:"name,attr" json:"name"`
	Conf       int    `xml:"conf,attr" json:"conf"`
	Method     string `xml:"method,attr" json:"method"`
	Version    string `xml:"version,attr" json:"version"`
	Product    string `xml:"product,attr" json:"product"`
	ExtraInfo  string `xml:"extrainfo,attr" json:"extrainfo"`
	Tunnel     string `xml:"tunnel,attr" json:"tunnel"`
	Proto      string `xml:"proto,attr" json:"proto"`
	Rpcnum     string `xml:"rpcnum,attr" json:"rpcnum"`
	Lowver     string `xml:"lowver,attr" json:"lowver"`
	Highver    string `xml:"hiver,attr" json:"hiver"`
	Hostname   string `xml:"hostname,attr" json:"hostname"`
	OsType     string `xml:"ostype,attr" json:"ostype"`
	DeviceType string `xml:"devicetype,attr" json:"devicetype"`
	ServiceFp  string `xml:"servicefp,attr" json:"servicefp"`
	CPEs       []CPE  `xml:"cpe" json:"cpes"`
}

// CPE (Common Platform Enumeration) is a standardized way to name
// software applications, operating systems, and hardware platforms.
type CPE string

// Script contains information from Nmap Scripting Engine.
type Script struct {
	Id       string    `xml:"id,attr" json:"id"`
	Output   string    `xml:"output,attr" json:"output"`
	Tables   []Table   `xml:"table" json:"tables"`
	Elements []Element `xml:"elem" json:"elements"`
}

// Table contains the output of the script in a more parse-able form.
type Table struct {
	Key      string    `xml:"key,attr" json:"key"`
	Elements []Element `xml:"elem" json:"elements"`
	Table    []Table   `xml:"table" json:"tables"`
}

// Element contains the output of the script, with detailed information
type Element struct {
	Key   string `xml:"key,attr" json:"key"`
	Value string `xml:",chardata" json:"value"`
}
