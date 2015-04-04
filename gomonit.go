// Package gomonit consumes and parses Monit status and event notifications. It disguises as M/Monit collector server.
package gomonit

import (
	"code.google.com/p/go-charset/charset"
	_ "code.google.com/p/go-charset/data"
	"encoding/xml"
	"errors"
	"github.com/fatih/structs"
	"io"
	"net/http"
	"time"
)

// Monit struct represents root XML node.
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

// Server represents monit>server node.
type Server struct {
	Uptime        int         `xml:"uptime"`
	Poll          int         `xml:"poll"`
	StartDelay    int         `xml:"startdelay"`
	LocalHostname string      `xml:"localhostname"`
	ControlFile   string      `xml:"controlfile"`
	Httpd         Httpd       `xml:"httpd"`
	Credentials   Credentials `xml:"credentials"`
}

// Httpd represents monit>server>httpd node.
type Httpd struct {
	Address string `xml:"address"`
	Port    int    `xml:"port"`
	Ssl     int    `xml:"ssl"`
}

// Credentials represents monit>server>credentials node.
type Credentials struct {
	Username string `xml:"username"`
	Password string `xml:"password"`
}

// Platform represents monit>platform node.
type Platform struct {
	Name    string `xml:"name"`
	Release string `xml:"release"`
	Version string `xml:"version"`
	Machine string `xml:"machine"`
	Cpu     string `xml:"cpu"`
	Memory  string `xml:"memory"`
	Swap    string `xml:"swap"`
}

// Monit service type identifiers.
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

// Service struct represents generalized monit>services>service XML node. It's used internally for XML parsing purposes and concrete service types should be used.
type Service struct {
	Name          string         `xml:"name,attr"`
	Type          uint           `xml:"type"`
	CollectedSec  int64          `xml:"collected_sec"`
	CollectedUsec int64          `xml:"collected_usec"`
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
	Timestamp     int64          `xml:"timestamp"`
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

// Returns concrete Filesystem service struct.
func (service *Service) GetFilesystem() (Filesystem, error) {
	var filesystem Filesystem
	var err error = nil

	if service.Type == ServiceTypeFilesystem {
		copy(service, &filesystem)
	} else {
		err = errors.New("Service type is not Filesystem")
	}

	return filesystem, err
}

// Returns concrete Directory service struct.
func (service *Service) GetDirectory() (Directory, error) {
	var directory Directory
	var err error = nil

	if service.Type == ServiceTypeDirectory {
		copy(service, &directory)
	} else {
		err = errors.New("Service type is not Directory")
	}

	return directory, err
}

// Returns concrete File service struct.
func (service *Service) GetFile() (File, error) {
	var file File
	var err error = nil

	if service.Type == ServiceTypeFile {
		copy(service, &file)
	} else {
		err = errors.New("Service type is not File")
	}

	return file, err
}

// Returns concrete Process service struct.
func (service *Service) GetProcess() (Process, error) {
	var process Process
	var err error = nil

	if service.Type == ServiceTypeProcess {
		copy(service, &process)
	} else {
		err = errors.New("Service type is not Process")
	}

	return process, err
}

// Returns concrete System service struct.
func (service *Service) GetSystem() (System, error) {
	var system System
	var err error = nil

	if service.Type == ServiceTypeSystem {
		copyCommon(service, &system)
		system.Cpu = service.System.Cpu
		system.Memory = service.System.Memory
		system.Load = service.System.Load
		system.Swap = service.System.Swap
	} else {
		err = errors.New("Service type is not System")
	}

	return system, err
}

// Returns concrete Fifo service struct.
func (service *Service) GetFifo() (Fifo, error) {
	var fifo Fifo
	var err error = nil

	if service.Type == ServiceTypeFifo {
		copy(service, &fifo)
	} else {
		err = errors.New("Service type is not Fifo")
	}

	return fifo, err
}

// Returns concrete Program service struct.
func (service *Service) GetProgram() (Program, error) {
	var program Program
	var err error = nil

	if service.Type == ServiceTypeProgram {
		copyCommon(service, &program)
		program.Status = service.Program.Status
		program.Started = service.Program.Started
		program.Output = service.Program.Output
	} else {
		err = errors.New("Service type is not Program")
	}

	return program, err
}

// Returns concrete Net service struct.
func (service *Service) GetNet() (Net, error) {
	var net Net
	var err error = nil

	if service.Type == ServiceTypeNet {
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
	} else {
		err = errors.New("Service type is not Net")
	}

	return net, err
}

func copy(src interface{}, dest interface{}) {
	srcStruct := structs.New(src)
	destStruct := structs.New(dest)

	for _, destField := range destStruct.Fields() {
		destFieldName := destField.Name()
		srcField, ok := srcStruct.FieldOk(destFieldName)

		if ok {
			srcValue := srcField.Value()

			if destFieldName == "Timestamp" {
				destField.Set(time.Unix(srcValue.(int64), 0))
			} else {
				destField.Set(srcValue)
			}
		}
	}

	copyTime(src, dest)
}

func copyCommon(src interface{}, dest interface{}) {
	keys := [9]string{
		"Name",
		"Type",
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

	copyTime(src, dest)
}

func copyTime(src interface{}, dest interface{}) {
	srcMap := structs.Map(src)
	destStruct := structs.New(dest)

	field := destStruct.Field("Time")
	field.Set(time.Unix(srcMap["CollectedSec"].(int64), srcMap["CollectedUsec"].(int64)))
}

// Filesystem represents concrete service XML node.
type Filesystem struct {
	Name          string
	Type          uint
	Time          time.Time
	Status        uint
	StatusHint    uint
	Monitor       uint
	MonitorMode   uint
	PendingAction uint
	Mode          string
	Uid           uint
	Gid           uint
	Flags         uint
	Block         FilesystemSize
	Inode         FilesystemSize
}

// FilesystemSize represents filesystem size XML node.
type FilesystemSize struct {
	Percent float32 `xml:"percent"`
	Usage   float64 `xml:"usage"`
	Total   float64 `xml:"total"`
}

// Directory represents concrete service XML node.
type Directory struct {
	Name          string
	Type          uint
	Time          time.Time
	Status        uint
	StatusHint    uint
	Monitor       uint
	MonitorMode   uint
	PendingAction uint
	Mode          string
	Uid           uint
	Gid           uint
	Timestamp     time.Time
}

// File represents concrete service XML node.
type File struct {
	Name          string
	Type          uint
	Time          time.Time
	Status        uint
	StatusHint    uint
	Monitor       uint
	MonitorMode   uint
	PendingAction uint
	Mode          string
	Uid           uint
	Gid           uint
	Timestamp     time.Time
	Size          uint64
}

// Process represents concrete service XML node.
type Process struct {
	Name          string
	Type          uint
	Time          time.Time
	Status        uint
	StatusHint    uint
	Monitor       uint
	MonitorMode   uint
	PendingAction uint
	Pid           uint
	PPid          uint
	Euid          uint
	Gid           uint
	Uptime        uint64
	Children      uint
	Memory        Memory
	Cpu           ProcessCpu
}

// System represents concrete service XML node.
type System struct {
	Name          string
	Type          uint
	Time          time.Time
	Status        uint
	StatusHint    uint
	Monitor       uint
	MonitorMode   uint
	PendingAction uint
	Cpu           SystemCpu
	Memory        Memory
	Load          Load
	Swap          Swap
}

// Fifo represents concrete service XML node.
type Fifo struct {
	Name          string
	Type          uint
	Time          time.Time
	Status        uint
	StatusHint    uint
	Monitor       uint
	MonitorMode   uint
	PendingAction uint
	Mode          string
	Uid           uint
	Gid           uint
	Timestamp     time.Time
}

// Program represents concrete service XML node.
type Program struct {
	Name          string
	Type          uint
	Time          time.Time
	Status        uint
	StatusHint    uint
	Monitor       uint
	MonitorMode   uint
	PendingAction uint
	Started       uint64
	Output        string
}

// net represents concrete service XML node.
type Net struct {
	Name          string
	Type          uint
	Time          time.Time
	Status        uint
	StatusHint    uint
	Monitor       uint
	MonitorMode   uint
	PendingAction uint
	State         uint
	Speed         uint64
	Duplex        uint
	DlPackets     NetLinkCount
	DlBytes       NetLinkCount
	DlErrors      NetLinkCount
	UlPackets     NetLinkCount
	UlBytes       NetLinkCount
	UlErrors      NetLinkCount
}

// ServiceSystem represents monit>service>system XML node.
type ServiceSystem struct {
	Cpu    SystemCpu `xml:"cpu"`
	Memory Memory    `xml:"memory"`
	Load   Load      `xml:"load"`
	Swap   Swap      `xml:"swap"`
}

// ServiceProgram represents monit>service>program XML node.
type ServiceProgram struct {
	Status  uint   `xml:"status"`
	Started uint64 `xml:"started"`
	Output  string `xml:"output"`
}

// Memory represents memory XML node.
type Memory struct {
	Percent       float64 `xml:"percent"`
	PercentTotal  float64 `xml:"percenttotal"`
	Kilobyte      uint    `xml:"kilobyte"`
	KilobyteTotal uint    `xml:"kilobytetotal"`
}

// SystemCpu represents monit>service>system>cpu XML node.
type SystemCpu struct {
	User   float64 `xml:"user"`
	System float64 `xml:"system"`
	Wait   float64 `xml:"wait"`
}

// ProcessSystem represents monit>service>cpu XML node.
type ProcessCpu struct {
	Percent      float64 `xml:"percent"`
	PercentTotal float64 `xml:"percenttotal"`
}

// Load represents monit>service>system>load XML node.
type Load struct {
	Avg01 float64 `xml:"avg01"`
	Avg05 float64 `xml:"avg05"`
	Avg15 float64 `xml:"avg15"`
}

// Swap represents monit>service>system>swap XML node.
type Swap struct {
	Percent  float64 `xml:"percent"`
	Kilobyte int     `xml:"kilobyte"`
}

// Link represents monit>service>link XML node.
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

// NetLinkCount represents monit>service>link upload/download counter XML node.
type NetLinkCount struct {
	Now   uint64 `xml:"now"`
	Total uint64 `xml:"total"`
}

// ServiceGroup represents monit>servicegroup XML node.
type ServiceGroup struct {
	Name    string `xml:"name,attr"`
	Service string `xml:"service"`
}

// Event represents monit>event XML node.
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

// Decoder implements XML decoder.
type Decoder interface {
	DecodeElement(interface{}, *xml.StartElement) error
}

// Parser is Monit notification decoder.
type Parser struct {
	Decoder Decoder
}

// NewParser returns a new Parser.
func NewParser(reader io.Reader) *Parser {
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReader

	return &Parser{decoder}
}

// Parse returns decoded Monit notification.
func (parser *Parser) Parse() Monit {
	var monit Monit
	parser.Decoder.DecodeElement(&monit, nil)
	return monit
}

// Collector is implementation of M/Monit servers /collector endpoint.
type Collector struct {
	Channel chan *Monit
	Handler http.HandlerFunc
}

// NewCollector returns a new Collector.
func NewCollector(channel chan *Monit) *Collector {
	handler := MakeHTTPHandler(channel)
	return &Collector{channel, handler}
}

// ServeHTTP implements an http.Handler that collects Monit monitifications.
func (collector *Collector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	collector.Handler(w, r)
}

// MakeHTTPHandler returns http.HandlerFunc function that parses HTTP request body and
// pipes result to provided channel.
func MakeHTTPHandler(out chan *Monit) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		parser := NewParser(r.Body)
		monit := parser.Parse()
		out <- &monit
	}
}
