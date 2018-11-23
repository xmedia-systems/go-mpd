package mpd

import (
	"bytes"
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	msec = 1000000
	sec  = 1000 * msec
	hour = 3600 * sec
)

func TestXSDDuration_String(t *testing.T) {
	assert.Equal(t, "PT1S", XSDDuration(sec).String())
	assert.Equal(t, "PT0.11S", XSDDuration(110*msec).String())
	assert.Equal(t, "PT1M", XSDDuration(60*sec).String())
	assert.Equal(t, "PT1M1S", XSDDuration(61*sec).String())
	assert.Equal(t, "PT1M1.1S", XSDDuration(61*sec+100*msec).String())
	assert.Equal(t, "PT1H", XSDDuration(hour).String())
	assert.Equal(t, "PT1H1M", XSDDuration(hour+60*sec).String())
	assert.Equal(t, "PT1H1M1S", XSDDuration(hour+61*sec).String())
	assert.Equal(t, "P1D", XSDDuration(24*hour).String())
	assert.Equal(t, "P1DT1H1M1S", XSDDuration(25*hour+61*sec).String())
	assert.Equal(t, "P1Y", XSDDuration(365*24*hour).String())
	assert.Equal(t, "P1Y1DT1H1M1S", XSDDuration(366*24*hour+hour+61*sec).String())

	assert.Equal(t, "-PT1S", XSDDuration(-sec).String())
	assert.Equal(t, "-P1Y1DT1H1M1S", XSDDuration(-(366*24*hour + hour + 61*sec)).String())
}

func checkDurationFromString(t *testing.T, str string, val int) {
	dur, err := XSDDurationFromString(str)
	assert.Nil(t, err)
	assert.Equal(t, XSDDuration(val), *dur)
}

func TestXSDDurationFromString(t *testing.T) {
	checkDurationFromString(t, "PT0S", 0)
	checkDurationFromString(t, "PT1S", sec)
	checkDurationFromString(t, "PT0.11S", 110*msec)

	checkDurationFromString(t, "PT1M", 60*sec)
	checkDurationFromString(t, "PT1M1S", 61*sec)
	checkDurationFromString(t, "PT1M1.1S", 61*sec+100*msec)
	checkDurationFromString(t, "PT1H", hour)
	checkDurationFromString(t, "PT1H1M", hour+60*sec)
	checkDurationFromString(t, "PT1H1M1S", hour+61*sec)
	checkDurationFromString(t, "P1D", 24*hour)
	checkDurationFromString(t, "P1DT1H1M1S", 25*hour+61*sec)
	checkDurationFromString(t, "P1Y", 365*24*hour)
	checkDurationFromString(t, "P1Y1DT1H1M1S", 366*24*hour+hour+61*sec)

	checkDurationFromString(t, "-PT1S", -sec)
	checkDurationFromString(t, "-P1Y1DT1H1M1S", -(366*24*hour + hour + 61*sec))

	_, err := XSDDurationFromString("PT")
	assert.Equal(t, invalidFormatError, err)

	_, err = XSDDurationFromString("P1.")
	assert.Equal(t, invalidFormatError, err)

	_, err = XSDDurationFromString("PT1.")
	assert.Equal(t, invalidFormatError, err)

	_, err = XSDDurationFromString("PT1.S")
	assert.Equal(t, invalidFormatError, err)
}

type DurationAttr struct {
	Duration *XSDDuration `xml:"duration,attr"`
}

func TestXSDDuration_UnmarshalXMLAttr(t *testing.T) {
	dur := DurationAttr{}
	err := xml.Unmarshal([]byte(`<foo duration="PT1S"></foo>`), &dur)
	assert.Nil(t, err)
	assert.NotNil(t, dur.Duration)
	assert.Equal(t, XSDDuration(sec), *dur.Duration)
}

func TestXSDDuration_MarshalXMLAttr(t *testing.T) {
	val := XSDDuration(2 * sec)
	dur := DurationAttr{Duration: &val}

	b := new(bytes.Buffer)
	e := xml.NewEncoder(b)
	err := e.Encode(dur)

	assert.Nil(t, err)
	assert.Equal(t, `<DurationAttr duration="PT2S"></DurationAttr>`, b.String())
}
