package main

import (
	"bytes"
	"code.google.com/p/go-charset/charset"
	_ "code.google.com/p/go-charset/data"
	"encoding/xml"
	"github.com/davecgh/go-spew/spew"
	"io"
	"log"
	"net/http"
    // "os"
)

type Server struct {
	Uptime        int         `xml:"uptime"`
	Poll          int         `xml:"poll"`
	StartDelay    int         `xml:"startdelay"`
	LocalHostname string      `xml:"localhostname"`
	ControlFile   string      `xml:"controlfile"`
	Httpd         Httpd       `xml:"httpd"`
	Credentials   Credentials `xml:"credentials"`
}

type Httpd struct {
	Address string `xml:"address"`
	Port    int    `xml:"port"`
	Ssl     int    `xml:"ssl"`
}

type Credentials struct {
	Username string `xml:"username"`
	Password string `xml:"password"`
}

type Platform struct {
	Name    string `xml:"name"`
	Release string `xml:"release"`
	Version string `xml:"version"`
	Machine string `xml:"machine"`
	Cpu     string `xml:"cpu"`
	Memory  string `xml:"memory"`
	Swap    string `xml:"swap"`
}

const (
	ServiceTypeFilesystem int = 0
	ServiceTypeDirectory      = 1
	ServiceTypeFile           = 2
	ServiceTypeProcess        = 3
	ServiceTypeSystem         = 5
	ServiceTypeFifo           = 6
	ServiceTypeProgram        = 7
	ServiceTypeNet            = 8
)

type Service struct {
	Name              string `xml:"name,attr"`
	Type              int    `xml:"type"`
	CollectedSec      int    `xml:"collected_sec"`
	CollectedUsec     int    `xml:"collected_usec"`
	Status            int    `xml:"status"`
	StatusHint        int    `xml:"status_hint"`
	Monitor           int    `xml:"monitor"`
	MonitorMode       int    `xml:"monitormode"`
	PendingAction     int    `xml:"pendingaction"`
	FilesystemDetails Filesystem
	DirectoryDetails  Directory
	FileDetails       File
	ProcessDetails    Process
	SystemDetails     System `xml:"system"`
	FifoDetails       Fifo
	ProgramDetails    Program
	NetDetails        Net
}

type Filesystem struct {
	Mode  string         `xml:"mode"`
	Uid   uint           `xml:"uid"`
	Gid   uint           `xml:"gid"`
	Flags uint           `xml:"flags"`
	Block FilesystemSize `xml:"block"`
	Inode FilesystemSize `xml:"inode"`
}

type FilesystemSize struct {
	Percent float32 `xml:"percent"`
	Usage   float64 `xml:"usage"`
	Total   float64 `xml:"total"`
}

type Directory struct {
	Mode      string `xml:"mode"`
	Uid       uint   `xml:"uid"`
	Gid       uint   `xml:"gid"`
	Timestamp uint64 `xml:"timestamp"`
}

type File struct {
	Mode      string `xml:"mode"`
	Uid       uint   `xml:"uid"`
	Gid       uint   `xml:"gid"`
	Timestamp uint64 `xml:"timestamp"`
	Size      uint64 `xml:"size"`
}

type Process struct {
	Pid      uint       `xml:"pid"`
	PPid     uint       `xml:"ppid"`
	Euid     uint       `xml:"euid"`
	Gid      uint       `xml:"gid"`
	Uptime   uint64     `xml:"uptime"`
	Children uint       `xml:"children"`
	Memory   Memory     `xml:"memory"`
	Cpu      ProcessCpu `xml:"cpu"`
}

type Fifo struct {
	Mode      string `xml:"mode"`
	Uid       uint   `xml:"uid"`
	Gid       uint   `xml:"gid"`
	Timestamp uint64 `xml:"timestamp"`
}

type Program struct {
	Started uint64 `xml:"started"`
	Status  uint   `xml:"status"`
	Output  string `xml:"output,chardata"`
}

type Net struct {
	Link NetLink `xml:"link"`
}

type NetLink struct {
	State     uint         `xml:"state"`
	Speed     uint64       `xml:"speed"`
	Duplex    uint         `xml:"duplex"`
	DlPackets NetLinkCount `xml:"download>packets"`
	DlBytes   NetLinkCount `xml:"download>bytes"`
	DlErrors  NetLinkCount `xml:"download>errors"`
	UlPackets NetLinkCount `xml:"upload>packets"`
	UlBytes   NetLinkCount `xml:"upload>bytes"`
	UlErrors  NetLinkCount `xml:"upload>errors"`
}

type NetLinkCount struct {
	Now   uint64 `xml:"now"`
	Total uint64 `xml:"total"`
}

type ServiceGroup struct {
	Name    string `xml:"name,attr"`
	Service string `xml:"service"`
}

type System struct {
	Cpu    SystemCpu `xml:"cpu"`
	Memory Memory    `xml:"memory"`
	Load   Load      `xml:"load"`
	Swap   Swap      `xml:"swap"`
}

type Load struct {
	Avg01 float64 `xml:"avg01"`
	Avg05 float64 `xml:"avg05"`
	Avg15 float64 `xml:"avg15"`
}

type SystemCpu struct {
	User   float64 `xml:"user"`
	System float64 `xml:"system"`
	Wait   float64 `xml:"wait"`
}

type ProcessCpu struct {
	Percent      float64 `xml:"percent"`
	PercentTotal float64 `xml:"percenttotal"`
}

type Memory struct {
	Percent       float64 `xml:"percent"`
	PercentTotal  float64 `xml:"percenttotal"`
	Kilobyte      uint    `xml:"kilobyte"`
	KilobyteTotal uint    `xml:"kilobytetotal"`
}

type Swap struct {
	Percent  float64 `xml:"percent"`
	Kilobyte int     `xml:"kilobyte"`
}

type Event struct {
	CollectedSec  int    `xml:"collected_sec"`
	CollectedUsec int    `xml:"collected_usec"`
	Service       string `xml:"service"`
	Type          int    `xml:"type"`
	Id            int    `xml:"id"`
	State         int    `xml:"state"`
	Action        int    `xml:"action"`
	Message       string `xml:"message,chardata"`
	Token         string `xml:"token"`
}

type Monit struct {
	Id            string         `xml:"id,attr"`
	Incarnation   string         `xml:"incarnation,attr"`
	Version       string         `xml:"version,attr"`
	Server        Server         `xml:"server"`
	Platform      Platform       `xml:"platform"`
	Services      []Service      `xml:"services>service"`
	ServiceGroups []ServiceGroup `xml:"servicegroups>servicegroup"`
	Event         Event          `xml:"event"`
}

type ServiceType struct {
	Type int `xml:"type"`
}

// func (service *Service) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
//     var err error
//     var st ServiceType
//     // var p Process
//
//     start2 := start.Copy()
//
//     if err = d.DecodeElement(&st, &start); err != nil {
//         spew.Dump(err)
//         return err
//     }
//
//     if err = d.DecodeElement(&st, &start2); err != nil {
//         spew.Dump(err)
//         return err
//     }
//
//     if st.Type == 3 {
//         // if err = d.DecodeElement(&p, &start); err != nil {
//         //     spew.Dump(err)
//         //     return err
//         // }
//     }
//
//     // spew.Dump(p)
//
//     return nil
// }

func Parse(reader io.Reader) Monit {
	var monit Monit

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReader
	decoder.DecodeElement(&monit, nil)

	return monit
}

type Collector struct {
	channel chan *Monit
}

func NewCollector(channel chan *Monit) *Collector {
	return &Collector{channel}
}

func (collector *Collector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	MakeHTTPHandler(collector.channel)(w, r)
}

func (collector *Collector) Serve() {
	http.HandleFunc("/collector", collector.ServeHTTP)
	err := http.ListenAndServe(":5001", nil)
	if err != nil {
		log.Fatal("http.ListenAndServe: ", err)
	}
}

func MakeHTTPHandler(out chan *Monit) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		monit := Parse(r.Body)
		out <- &monit
	}
}

func decode(reader io.Reader) string {
	b := new(bytes.Buffer)
	b.ReadFrom(reader)

	return b.String()
}

func main() {
	// var monit Monit

	// file, _ := os.Open("stub.xml")
	// _ = Parse(file)

	// spew.Dump(monit)

	channel := make(chan *Monit, 1)

	collector := NewCollector(channel)
	go collector.Serve()

	for monit := range channel {
		spew.Dump(monit.Services)
	}
}
