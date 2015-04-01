package gomonit

import (
	"encoding/xml"
    // "github.com/davecgh/go-spew/spew"
    // "reflect"
	"strings"
	"testing"
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
