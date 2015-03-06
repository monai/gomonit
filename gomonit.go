package main

import (
	// "bytes"
	"code.google.com/p/go-charset/charset"
	_ "code.google.com/p/go-charset/data"
	"encoding/xml"
	// "fmt"
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
}

type Monit struct {
	Id          string    `xml:"id,attr"`
	Incarnation string    `xml:"incarnation,attr"`
	Version     string    `xml:"version,attr"`
	Server      Server    `xml:"server"`
	Platform    Platform  `xml:"platform"`
	Services    []Service `xml:"services>service"`
}

func MonitServer(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	var monit Monit

	decoder := xml.NewDecoder(req.Body)
	decoder.CharsetReader = charset.NewReader
	decoder.DecodeElement(&monit, nil)

	log.Println("Got message from", monit)

	// b := new(bytes.Buffer)
	// b.ReadFrom(req.Body)
	// log.Fatal(b.String())

	io.WriteString(w, "hello, world!\n")
}

func main() {
	http.HandleFunc("/collector", MonitServer)
	err := http.ListenAndServe(":5001", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
