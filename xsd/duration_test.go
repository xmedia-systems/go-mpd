package xsd

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

func TestDuration_String(t *testing.T) {
	assert.Equal(t, "PT1S", Duration(sec).String())
	assert.Equal(t, "PT0.11S", Duration(110*msec).String())
	assert.Equal(t, "PT1M", Duration(60*sec).String())
	assert.Equal(t, "PT1M1S", Duration(61*sec).String())
	assert.Equal(t, "PT1M1.1S", Duration(61*sec+100*msec).String())
	assert.Equal(t, "PT1H", Duration(hour).String())
	assert.Equal(t, "PT1H1M", Duration(hour+60*sec).String())
	assert.Equal(t, "PT1H1M1S", Duration(hour+61*sec).String())
	assert.Equal(t, "P1D", Duration(24*hour).String())
	assert.Equal(t, "P1DT1H1M1S", Duration(25*hour+61*sec).String())
	assert.Equal(t, "P1Y", Duration(365*24*hour).String())
	assert.Equal(t, "P1Y1DT1H1M1S", Duration(366*24*hour+hour+61*sec).String())

	assert.Equal(t, "-PT1S", Duration(-sec).String())
	assert.Equal(t, "-P1Y1DT1H1M1S", Duration(-(366*24*hour + hour + 61*sec)).String())
}

type DurationFromStringType func(string) (*Duration, error)

func checkDurationFromString(t *testing.T, fromString DurationFromStringType, str string, val int) {
	dur, err := fromString(str)
	assert.Nil(t, err)
	assert.Equal(t, Duration(val), *dur)
}

func checkDurationWithMethod(t *testing.T, fromString DurationFromStringType) {
	checkDurationFromString(t, fromString, "PT0S", 0)
	checkDurationFromString(t, fromString, "PT1S", sec)
	checkDurationFromString(t, fromString, "PT0.11S", 110*msec)

	checkDurationFromString(t, fromString, "PT1M", 60*sec)
	checkDurationFromString(t, fromString, "PT1M1S", 61*sec)
	checkDurationFromString(t, fromString, "PT1M1.1S", 61*sec+100*msec)
	checkDurationFromString(t, fromString, "PT1H", hour)
	checkDurationFromString(t, fromString, "PT1H1M", hour+60*sec)
	checkDurationFromString(t, fromString, "PT1H1M1S", hour+61*sec)
	checkDurationFromString(t, fromString, "P1D", 24*hour)
	checkDurationFromString(t, fromString, "P1DT1H1M1S", 25*hour+61*sec)
	checkDurationFromString(t, fromString, "P1Y", 365*24*hour)
	checkDurationFromString(t, fromString, "P1Y1DT1H1M1S", 366*24*hour+hour+61*sec)

	checkDurationFromString(t, fromString, "-PT1S", -sec)
	checkDurationFromString(t, fromString, "-P1Y1DT1H1M1S", -(366*24*hour + hour + 61*sec))

	_, err := fromString("PT")
	assert.Equal(t, invalidFormatError, err)

	_, err = fromString("P1.")
	assert.Equal(t, invalidFormatError, err)

	_, err = fromString("PT1.")
	assert.Equal(t, invalidFormatError, err)

	_, err = fromString("PT1.S")
	assert.Equal(t, invalidFormatError, err)

	_, err = fromString("xPT1S")
	assert.Equal(t, invalidFormatError, err)

	_, err = fromString("PT1Sx")
	assert.Equal(t, invalidFormatError, err)
}

func TestDurationFromString(t *testing.T) {
	checkDurationWithMethod(t, DurationFromString)
}

func TestDurationFromString2(t *testing.T) {
	checkDurationWithMethod(t, DurationFromString2)
}

type DurationAttr struct {
	Duration *Duration `xml:"duration,attr"`
}

func TestDuration_UnmarshalXMLAttr(t *testing.T) {
	dur := DurationAttr{}
	err := xml.Unmarshal([]byte(`<foo duration="PT1S"></foo>`), &dur)
	assert.Nil(t, err)
	assert.NotNil(t, dur.Duration)
	assert.Equal(t, Duration(sec), *dur.Duration)
}

func TestDuration_MarshalXMLAttr(t *testing.T) {
	val := Duration(2 * sec)
	dur := DurationAttr{Duration: &val}

	b := new(bytes.Buffer)
	e := xml.NewEncoder(b)
	err := e.Encode(dur)

	assert.Nil(t, err)
	assert.Equal(t, `<DurationAttr duration="PT2S"></DurationAttr>`, b.String())
}

func BenchmarkDuration_String(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Duration(366*24*hour + hour + 61*sec).String()
	}
}

func BenchmarkDurationFromString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = DurationFromString("P1Y2DT3H4M5.67S")
	}
}

func BenchmarkDurationFromString2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = DurationFromString2("P1Y2DT3H4M5.67S")
	}
}
