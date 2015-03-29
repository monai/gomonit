package gomonit

import (
	"fmt"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	xml := `
        <?xml version="1.0" encoding="ISO-8859-1"?>
        <monit id="046455007b101404405f6741927c0072" incarnation="1427018447" version="5.6">
        </monit>
    `

	reader := strings.NewReader(xml)
	monit := Parse(reader)

	fmt.Println(monit)
}
