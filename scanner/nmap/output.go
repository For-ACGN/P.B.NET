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
