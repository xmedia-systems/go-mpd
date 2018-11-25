package xsd

import (
	"encoding/xml"
	"errors"
	"fmt"
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

func DurationFromString(str string) (*Duration, error) {
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
