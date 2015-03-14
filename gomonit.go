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

type Service struct {
	Name          string `xml:"name,attr"`
	Type          int    `xml:"type"`
	CollectedSec  int    `xml:"collected_sec"`
	CollectedUsec int    `xml:"collected_usec"`
	Status        int    `xml:"status"`
	StatusHint    int    `xml:"status_hint"`
	Monitor       int    `xml:"monitor"`
	MonitorMode   int    `xml:"monitormode"`
	PendingAction int    `xml:"pendingaction"`
	Pid           int    `xml:"pid"`
	PPid          int    `xml:"ppid"`
	Uptime        int    `xml:"uptime"`
	Children      int    `xml:"children"`
	Cpu           Cpu    `xml:"cpu"`
	Memory        Memory `xml:"memory"`
	System        System `xml:"system"`
}

type ServiceGroup struct {
	Name    string `xml:"name,attr"`
	Service string `xml:"service"`
}

type System struct {
	Cpu    Cpu    `xml:"cpu"`
	Memory Memory `xml:"memory"`
	Load   Load   `xml:"load"`
	Swap   Swap   `xml:"swap"`
}

type Load struct {
	Avg01 float64 `xml:"avg01"`
	Avg05 float64 `xml:"avg05"`
	Avg15 float64 `xml:"avg15"`
}

type Cpu struct {
	User   float64 `xml:"user"`
	System float64 `xml:"system"`
	Wait   float64 `xml:"wait"`
}

type Memory struct {
	Percent  float64 `xml:"percent"`
	Kilobyte int     `xml:"kilobyte"`
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
	channel := make(chan *Monit, 1)

	collector := NewCollector(channel)
	go collector.Serve()

	for monit := range channel {
		spew.Dump(monit)
	}
}
