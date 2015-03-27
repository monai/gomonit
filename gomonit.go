package main

import (
	"bytes"
	"code.google.com/p/go-charset/charset"
	_ "code.google.com/p/go-charset/data"
	"encoding/xml"
	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/structs"
	"io"
	"log"
	"net/http"
)

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
	ServiceTypeFilesystem uint = 0
	ServiceTypeDirectory       = 1
	ServiceTypeFile            = 2
	ServiceTypeProcess         = 3
	ServiceTypeSystem          = 5
	ServiceTypeFifo            = 6
	ServiceTypeProgram         = 7
	ServiceTypeNet             = 8
)

type Service struct {
	Name          string         `xml:"name,attr"`
	Type          uint           `xml:"type"`
	CollectedSec  uint           `xml:"collected_sec"`
	CollectedUsec uint           `xml:"collected_usec"`
	Status        uint           `xml:"status"`
	StatusHint    uint           `xml:"status_hint"`
	Monitor       uint           `xml:"monitor"`
	MonitorMode   uint           `xml:"monitormode"`
	PendingAction uint           `xml:"pendingaction"`
	Mode          string         `xml:"mode"`
	Uid           uint           `xml:"uid"`
	Gid           uint           `xml:"gid"`
	Flags         uint           `xml:"flags"`
	Block         FilesystemSize `xml:"block"`
	Inode         FilesystemSize `xml:"inode"`
	Timestamp     uint64         `xml:"timestamp"`
	Size          uint64         `xml:"size"`
	Pid           uint           `xml:"pid"`
	PPid          uint           `xml:"ppid"`
	Euid          uint           `xml:"euid"`
	Uptime        uint64         `xml:"uptime"`
	Children      uint           `xml:"children"`
	Memory        Memory         `xml:"memory"`
	Cpu           ProcessCpu     `xml:"cpu"`
	System        ServiceSystem  `xml:"system"`
	Program       ServiceProgram `xml:"program"`
	Link          Link           `xml:"link"`
}

func (service *Service) GetFilesystem() (Filesystem, error) {
	var filesystem Filesystem
	copy(service, &filesystem)
	return filesystem, nil
}

func (service *Service) GetDirectory() (Directory, error) {
	var directory Directory
	copy(service, &directory)
	return directory, nil
}

func (service *Service) GetFile() (File, error) {
	var file File
	copy(service, &file)
	return file, nil
}

func (service *Service) GetProcess() (Process, error) {
	var process Process
	copy(service, &process)
	return process, nil
}

func (service *Service) GetSystem() (System, error) {
	var system System

	copyCommon(service, &system)
	system.Cpu = service.System.Cpu
	system.Memory = service.System.Memory
	system.Load = service.System.Load
	system.Swap = service.System.Swap

	return system, nil
}

func (service *Service) GetFifo() (Fifo, error) {
	var fifo Fifo
	copy(service, &fifo)
	return fifo, nil
}

func (service *Service) GetProgram() (Program, error) {
	var program Program

	copyCommon(service, &program)
	program.Status = service.Program.Status
	program.Started = service.Program.Started
	program.Output = service.Program.Output

	return program, nil
}

func (service *Service) GetNet() (Net, error) {
	var net Net

	copyCommon(service, &net)
	net.State = service.Link.State
	net.Speed = service.Link.Speed
	net.Duplex = service.Link.Duplex
	net.DlPackets = service.Link.DlPackets
	net.DlBytes = service.Link.DlBytes
	net.DlErrors = service.Link.DlErrors
	net.UlPackets = service.Link.UlPackets
	net.UlBytes = service.Link.UlBytes
	net.UlErrors = service.Link.UlErrors

	return net, nil
}

func copy(src interface{}, dest interface{}) {
	srcStruct := structs.New(src)
	destStruct := structs.New(dest)

	for _, destField := range destStruct.Fields() {
		srcField, ok := srcStruct.FieldOk(destField.Name())
		srcValue := srcField.Value()

		if ok {
			destField.Set(srcValue)
		}

	}
}

func copyCommon(src interface{}, dest interface{}) {
	keys := [9]string{
		"Name",
		"Type",
		"CollectedSec",
		"CollectedUsec",
		"Status",
		"StatusHint",
		"Monitor",
		"MonitorMode",
		"PendingAction"}

	srcMap := structs.Map(src)
	destStruct := structs.New(dest)

	for _, key := range keys {
		field, ok := destStruct.FieldOk(key)
		if ok {
			field.Set(srcMap[key])
		}
	}
}

type Filesystem struct {
	Name          string         `xml:"name,attr"`
	Type          uint           `xml:"type"`
	CollectedSec  uint           `xml:"collected_sec"`
	CollectedUsec uint           `xml:"collected_usec"`
	Status        uint           `xml:"status"`
	StatusHint    uint           `xml:"status_hint"`
	Monitor       uint           `xml:"monitor"`
	MonitorMode   uint           `xml:"monitormode"`
	PendingAction uint           `xml:"pendingaction"`
	Mode          string         `xml:"mode"`
	Uid           uint           `xml:"uid"`
	Gid           uint           `xml:"gid"`
	Flags         uint           `xml:"flags"`
	Block         FilesystemSize `xml:"block"`
	Inode         FilesystemSize `xml:"inode"`
}

type FilesystemSize struct {
	Percent float32 `xml:"percent"`
	Usage   float64 `xml:"usage"`
	Total   float64 `xml:"total"`
}

type Directory struct {
	Name          string `xml:"name,attr"`
	Type          uint   `xml:"type"`
	CollectedSec  uint   `xml:"collected_sec"`
	CollectedUsec uint   `xml:"collected_usec"`
	Status        uint   `xml:"status"`
	StatusHint    uint   `xml:"status_hint"`
	Monitor       uint   `xml:"monitor"`
	MonitorMode   uint   `xml:"monitormode"`
	PendingAction uint   `xml:"pendingaction"`
	Mode          string `xml:"mode"`
	Uid           uint   `xml:"uid"`
	Gid           uint   `xml:"gid"`
	Timestamp     uint64 `xml:"timestamp"`
}

type File struct {
	Name          string `xml:"name,attr"`
	Type          uint   `xml:"type"`
	CollectedSec  uint   `xml:"collected_sec"`
	CollectedUsec uint   `xml:"collected_usec"`
	Status        uint   `xml:"status"`
	StatusHint    uint   `xml:"status_hint"`
	Monitor       uint   `xml:"monitor"`
	MonitorMode   uint   `xml:"monitormode"`
	PendingAction uint   `xml:"pendingaction"`
	Mode          string `xml:"mode"`
	Uid           uint   `xml:"uid"`
	Gid           uint   `xml:"gid"`
	Timestamp     uint64 `xml:"timestamp"`
	Size          uint64 `xml:"size"`
}

type Process struct {
	Name          string     `xml:"name,attr"`
	Type          uint       `xml:"type"`
	CollectedSec  uint       `xml:"collected_sec"`
	CollectedUsec uint       `xml:"collected_usec"`
	Status        uint       `xml:"status"`
	StatusHint    uint       `xml:"status_hint"`
	Monitor       uint       `xml:"monitor"`
	MonitorMode   uint       `xml:"monitormode"`
	PendingAction uint       `xml:"pendingaction"`
	Pid           uint       `xml:"pid"`
	PPid          uint       `xml:"ppid"`
	Euid          uint       `xml:"euid"`
	Gid           uint       `xml:"gid"`
	Uptime        uint64     `xml:"uptime"`
	Children      uint       `xml:"children"`
	Memory        Memory     `xml:"memory"`
	Cpu           ProcessCpu `xml:"cpu"`
}

type System struct {
	Name          string    `xml:"name,attr"`
	Type          uint      `xml:"type"`
	CollectedSec  uint      `xml:"collected_sec"`
	CollectedUsec uint      `xml:"collected_usec"`
	Status        uint      `xml:"status"`
	StatusHint    uint      `xml:"status_hint"`
	Monitor       uint      `xml:"monitor"`
	MonitorMode   uint      `xml:"monitormode"`
	PendingAction uint      `xml:"pendingaction"`
	Cpu           SystemCpu `xml:"system>cpu"`
	Memory        Memory    `xml:"system>memory"`
	Load          Load      `xml:"system>load"`
	Swap          Swap      `xml:"system>swap"`
}

type Fifo struct {
	Name          string `xml:"name,attr"`
	Type          uint   `xml:"type"`
	CollectedSec  uint   `xml:"collected_sec"`
	CollectedUsec uint   `xml:"collected_usec"`
	Status        uint   `xml:"status"`
	StatusHint    uint   `xml:"status_hint"`
	Monitor       uint   `xml:"monitor"`
	MonitorMode   uint   `xml:"monitormode"`
	PendingAction uint   `xml:"pendingaction"`
	Mode          string `xml:"mode"`
	Uid           uint   `xml:"uid"`
	Gid           uint   `xml:"gid"`
	Timestamp     uint64 `xml:"timestamp"`
}

type Program struct {
	Name          string `xml:"name,attr"`
	Type          uint   `xml:"type"`
	CollectedSec  uint   `xml:"collected_sec"`
	CollectedUsec uint   `xml:"collected_usec"`
	Status        uint   `xml:"status"`
	StatusHint    uint   `xml:"status_hint"`
	Monitor       uint   `xml:"monitor"`
	MonitorMode   uint   `xml:"monitormode"`
	PendingAction uint   `xml:"pendingaction"`
	Started       uint64 `xml:"started"`
	Output        string `xml:"output"`
}

type Net struct {
	Name          string       `xml:"name,attr"`
	Type          uint         `xml:"type"`
	CollectedSec  uint         `xml:"collected_sec"`
	CollectedUsec uint         `xml:"collected_usec"`
	Status        uint         `xml:"status"`
	StatusHint    uint         `xml:"status_hint"`
	Monitor       uint         `xml:"monitor"`
	MonitorMode   uint         `xml:"monitormode"`
	PendingAction uint         `xml:"pendingaction"`
	State         uint         `xml:"state"`
	Speed         uint64       `xml:"speed"`
	Duplex        uint         `xml:"duplex"`
	DlPackets     NetLinkCount `xml:"download>packets"`
	DlBytes       NetLinkCount `xml:"download>bytes"`
	DlErrors      NetLinkCount `xml:"download>errors"`
	UlPackets     NetLinkCount `xml:"upload>packets"`
	UlBytes       NetLinkCount `xml:"upload>bytes"`
	UlErrors      NetLinkCount `xml:"upload>errors"`
}

type ServiceSystem struct {
	Cpu    SystemCpu `xml:"cpu"`
	Memory Memory    `xml:"memory"`
	Load   Load      `xml:"load"`
	Swap   Swap      `xml:"swap"`
}

type ServiceProgram struct {
	Status  uint   `xml:"status"`
	Started uint64 `xml:"started"`
	Output  string `xml:"output"`
}

type Memory struct {
	Percent       float64 `xml:"percent"`
	PercentTotal  float64 `xml:"percenttotal"`
	Kilobyte      uint    `xml:"kilobyte"`
	KilobyteTotal uint    `xml:"kilobytetotal"`
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

type Load struct {
	Avg01 float64 `xml:"avg01"`
	Avg05 float64 `xml:"avg05"`
	Avg15 float64 `xml:"avg15"`
}

type Swap struct {
	Percent  float64 `xml:"percent"`
	Kilobyte int     `xml:"kilobyte"`
}

type Link struct {
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
		for _, service := range monit.Services {
			if service.Type == ServiceTypeSystem {
				system, _ := service.GetSystem()
				spew.Dump(system)
			}
		}
	}
}
