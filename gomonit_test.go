package gomonit

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/monai/gomonit"
)

func TestParse(t *testing.T) {
	decoder := NewFakeDecoder(t)
	parser := NewParser(strings.NewReader(""))
	parser.Decoder = decoder

	go func() {
		decoder.AssertDecodeElement(nil)
		decoder.Close()
	}()

	parser.Parse()

	decoder.AssertDone(t)
}

type Call interface{}

type FakeDecoder struct {
	t     *testing.T
	Calls chan Call
}

func NewFakeDecoder(t *testing.T) *FakeDecoder {
	return &FakeDecoder{t, make(chan Call)}
}

type decodeElementCall struct {
	v     interface{}
	start *xml.StartElement
}
type decodeElementResp struct{ err error }

func (d *FakeDecoder) DecodeElement(v interface{}, start *xml.StartElement) error {
	d.Calls <- &decodeElementCall{v, start}
	return (<-d.Calls).(*decodeElementResp).err
}

func (d *FakeDecoder) AssertDecodeElement(err error) {
	call := (<-d.Calls).(*decodeElementCall)

	switch t := call.v.(type) {
	case *Monit:
	default:
		d.t.Errorf("XML should be unmarshaled to type *gomonit.Monit, got %T", t)
	}

	d.Calls <- &decodeElementResp{err}
}

func (d *FakeDecoder) Close() {
	close(d.Calls)
}

func (d *FakeDecoder) AssertDone(t *testing.T) {
	if _, more := <-d.Calls; more {
		t.Fatal("Did not expect more calls")
	}
}

func TestServiceTypes(t *testing.T) {
	cases := []struct {
		Name string
		Type uint
	}{
		{"Filesystem", 0},
		{"Directory", 1},
		{"File", 2},
		{"Process", 3},
		{"System", 5},
		{"Fifo", 6},
		{"Program", 7},
		{"Net", 8},
	}

	for i, c := range cases {
		service := new(Service)
		service.Name = c.Name
		service.Type = c.Type
		serviceValue := reflect.ValueOf(service)

		ni := i + 1
		if ni >= len(cases) {
			ni = 0
		}
		wrongName := cases[ni].Name
		methodName := "Get" + c.Name

		typeOf := reflect.TypeOf(service)
		method, found := typeOf.MethodByName(methodName)

		if !found {
			t.Errorf("Method %s doesn't exist", methodName)
		}

		resValue := method.Func.Call([]reflect.Value{serviceValue})

		if !resValue[1].IsNil() {
			t.Errorf("Method %s returned error with type %s", methodName, service.Name)
		}

		methodName = "Get" + wrongName
		method, _ = typeOf.MethodByName(methodName)
		resValue = method.Func.Call([]reflect.Value{serviceValue})

		if resValue[1].IsNil() {
			t.Errorf("Method %s didn't return error with type %s", methodName, service.Name)
		}
	}
}

func ExampleCollector() {
	// create channel and pass it to the collector
	channel := make(chan *gomonit.Monit)
	collector := gomonit.NewCollector(channel)
	http.Handle("/collector", collector)

	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal("http.ListenAndServe: ", err)
		}
	}()

	// consume notifications
	for monit := range channel {
		fmt.Println(monit.Server.Uptime)
	}
}
