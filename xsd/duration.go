package xsd

import (
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"
)

type Duration time.Duration

var (
	pattern = regexp.MustCompile(`^(?P<sign>-)?P((?P<Y>\d+)Y)?((?P<M>\d+)M)?((?P<D>\d+)D)?(T((?P<h>\d+)H)?((?P<m>\d+)M)?((?P<s>\d+)(?P<ms>\.\d+)?S)?)?$`)

	invalidFormatError = errors.New("format of string no valid duration")
	errNoMonth         = errors.New("non-zero value for months is not allowed")
)

func (d *Duration) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	if d == nil {
		return xml.Attr{}, nil
	}
	return xml.Attr{Name: name, Value: d.String()}, nil
}

func (d Duration) String() string {
	var buf [32]byte
	w := len(buf)

	u := uint64(d)
	neg := d < 0
	if neg {
		u = -u
	}

	w--
	buf[w] = 'S'

	w, u = fmtFrac(buf[:w], u, 9)

	if u == 0 {
		w--
		buf[w] = '0'
	} else {
		// u is now integer seconds
		w = fmtInt(buf[:w], u%60)
		u /= 60
	}

	// u is now integer minutes
	if u > 0 {
		w--
		buf[w] = 'M'
		w = fmtInt(buf[:w], u%60)
		u /= 60

		// u is now integer hours
		// Stop at hours because days can be different lengths.
		if u > 0 {
			w--
			buf[w] = 'H'
			w = fmtInt(buf[:w], u%24)
			u /= 24

			// only add 'T' if we have added some time before
			if w != len(buf) {
				w--
				buf[w] = 'T'
			}

			if u > 0 {
				w--
				buf[w] = 'D'
				w = fmtInt(buf[:w], u%365)
				u /= 365

				if u > 0 {
					w--
					buf[w] = 'Y'
					w = fmtInt(buf[:w], u)
				}
			}
		} else {
			w--
			buf[w] = 'T'
		}
	} else {
		w--
		buf[w] = 'T'
	}

	w--
	buf[w] = 'P'

	if neg {
		w--
		buf[w] = '-'
	}

	return string(buf[w:])

}

// from time.go
func fmtFrac(buf []byte, v uint64, prec int) (nw int, nv uint64) {
	// Omit trailing zeros up to and including decimal point.
	w := len(buf)
	p := false
	for i := 0; i < prec; i++ {
		digit := v % 10
		p = p || digit != 0
		if p {
			w--
			buf[w] = byte(digit) + '0'
		}
		v /= 10
	}
	if p {
		w--
		buf[w] = '.'
	}
	return w, v
}

// from time.go
func fmtInt(buf []byte, v uint64) int {
	w := len(buf)
	if v == 0 {
		w++
	} else {
		for v > 0 {
			w--
			buf[w] = byte(v%10) + '0'
			v /= 10
		}
	}
	return w
}

func DurationFromString2(str string) (*Duration, error) {
	var (
		match        []string
		dur          = int64(0)
		Sign         = int64(1)
		MSec         = int64(1000000)
		Sec          = 1000 * MSec
		Hour         = 3600 * Sec
		Day          = 24 * Hour
		Year         = 365 * Day
		NoMatchFound = true
	)

	if pattern.MatchString(str) == false {
		return nil, invalidFormatError
	}

	match = pattern.FindStringSubmatch(str)

	for i, name := range pattern.SubexpNames() {
		strVal := match[i]
		if i == 0 || name == "" || strVal == "" {
			continue
		}

		NoMatchFound = false

		if name == "sign" {
			Sign = -1
			continue
		}

		if name == "ms" {
			val, err := strconv.ParseFloat(strVal, 64)
			if err != nil {
				return nil, err
			}

			dur += int64(val * 1000000000)
			continue
		}

		val, err := strconv.ParseInt(strVal, 10, 64)
		if err != nil {
			return nil, err
		}

		switch name {
		case "Y":
			dur += val * Year
		case "M":
			if val != 0 {
				return nil, errNoMonth
			}
		case "D":
			dur += val * Day
		case "h":
			dur += val * Hour
		case "m":
			dur += val * 60 * Sec
		case "s":
			dur += val * Sec
		default:
			return nil, errors.New(fmt.Sprintf("unknown field %s", name))
		}
	}

	if NoMatchFound {
		return nil, invalidFormatError
	}

	dur *= Sign

	return (*Duration)(&dur), nil
}

// The implementation of DurationFromString2 was done with a Regex so this of course has
// an impact on performance. Inspired by String() I tried to write the parser by hand and
// on my local machine the benchmark shows quite a significant difference.
// I also wanted to try the benchmarking out.
//
// BenchmarkDurationFromString-8    	 5000000	       241 ns/op
// BenchmarkDurationFromString2-8   	 1000000	      1924 ns/op
func DurationFromString(str string) (*Duration, error) {
	var (
		dur = int64(0)
		buf = []byte(str)
		c   = len(buf) - 1

		msec = int64(1000000)
		sec  = 1000 * msec
		hour = 3600 * sec
		day  = 24 * hour
		year = 365 * day

		timeParsed = false
	)

	parsePartial := func(indicator byte, multiplier int64) error {
		if buf[c] != indicator {
			return nil
		}

		c--

		i := lookupInt(&buf, c)
		if i == c+1 {
			return nil
		}

		val, err := atoi(buf[i : c+1])
		if err != nil {
			return err
		}

		dur += val * multiplier
		c = i - 1

		timeParsed = true

		return nil
	}

	if buf[c] == 'S' {
		c--
		s := lookupInt(&buf, c)
		if s == c+1 {
			return nil, invalidFormatError
		}

		if buf[s-1] == '.' {
			nsVal, err := atoi(buf[s : c+1])
			if err != nil {
				return nil, err
			}
			remainingZeroes := 8 - (c - s)
			dur += nsVal * int64(math.Pow10(remainingZeroes))
			c = s - 2

			s = lookupInt(&buf, c)

			if s == c+1 {
				return nil, invalidFormatError
			}
		}

		sVal, err := atoi(buf[s : c+1])
		if err != nil {
			return nil, err
		}
		dur += sec * sVal
		c = s - 1

		timeParsed = true
	}

	if err := parsePartial('M', 60*sec); err != nil {
		return nil, err
	}

	if err := parsePartial('H', hour); err != nil {
		return nil, err
	}

	if timeParsed {
		if buf[c] != 'T' {
			return nil, invalidFormatError
		}

		c--
	}

	if err := parsePartial('D', day); err != nil {
		return nil, err
	}

	if buf[c] == 'M' && buf[c+1] != '0' {
		return nil, errNoMonth
	}

	if err := parsePartial('Y', year); err != nil {
		return nil, err
	}

	if buf[c] == 'P' {
		if c == 1 && buf[c-1] == '-' {
			dur *= -1
		} else if c != 0 {
			return nil, invalidFormatError
		}
	} else {
		return nil, invalidFormatError
	}

	return (*Duration)(&dur), nil
}

func atoi(buf []byte) (x int64, err error) {
	for i := 0; i < len(buf); i++ {
		c := buf[i]
		if x > (1<<63-1)/10 {
			// overflow
			return 0, invalidFormatError
		}
		x = x*10 + int64(c) - '0'
		if x < 0 {
			// overflow
			return 0, invalidFormatError
		}
	}

	return x, err
}

func lookupInt(buf *[]byte, i int) int {
	for ; i >= 0; i-- {
		c := (*buf)[i]
		if c < '0' || c > '9' {
			return i + 1
		}
	}

	return i
}

func (d *Duration) UnmarshalXMLAttr(attr xml.Attr) error {
	dur, err := DurationFromString(attr.Value)
	if err != nil {
		return err
	}

	*d = *dur
	return nil
}

// check interfaces
var (
	dur                     = Duration(0)
	_   xml.MarshalerAttr   = &dur
	_   xml.UnmarshalerAttr = &dur
)
