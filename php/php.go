package php

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Time time.Time

// DateTimeInterface constants - PHP date format constants
const (
	DateTimeATOM            = `Y-m-d\TH:i:sP`
	DateTimeCOOKIE          = `l, d-M-Y H:i:s T`
	DateTimeISO8601         = `Y-m-d\TH:i:sO`
	DateTimeISO8601Expanded = `X-m-d\TH:i:sP`
	DateTimeRFC822          = `D, d M y H:i:s O`
	DateTimeRFC850          = `l, d-M-y H:i:s T`
	DateTimeRFC1036         = `D, d M y H:i:s O`
	DateTimeRFC1123         = `D, d M Y H:i:s O`
	DateTimeRFC7231         = `D, d M Y H:i:s \G\M\T`
	DateTimeRFC2822         = `D, d M Y H:i:s O`
	DateTimeRFC3339         = `Y-m-d\TH:i:sP`
	DateTimeRFC3339Extended = `Y-m-d\TH:i:s.vP`
	DateTimeRSS             = `D, d M Y H:i:s O`
	DateTimeW3C             = `Y-m-d\TH:i:sP`
)

// Format formats using PHP's date() format string.
// See https://www.php.net/manual/en/datetime.format.php
// Literal characters can be escaped with a backslash.
func (t *Time) Format(format string) string {
	goTime := time.Time(*t)
	var result strings.Builder

	runes := []rune(format)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		if ch == '\\' && i+1 < len(runes) {
			i++
			result.WriteRune(runes[i])
			continue
		}
		result.WriteString(phpChar(goTime, ch))
	}

	return result.String()
}

// Parse parses value using PHP's date() format string, the inverse of Format.
// Tokens that produce non-reversible output (S, N, w, z, W, t, L, o, I, B, X, Z, U, e)
// return an error.
func Parse(format, value string) (Time, error) {
	layout, err := phpFormatToGoLayout(format)
	if err != nil {
		return Time{}, err
	}
	t, err := time.Parse(layout, value)
	if err != nil {
		return Time{}, err
	}
	return Time(t), nil
}

func phpFormatToGoLayout(format string) (string, error) {
	var sb strings.Builder
	runes := []rune(format)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		if ch == '\\' && i+1 < len(runes) {
			i++
			sb.WriteRune(runes[i])
			continue
		}
		switch ch {
		case 'd':
			sb.WriteString("02")
		case 'D':
			sb.WriteString("Mon")
		case 'j':
			sb.WriteString("2")
		case 'l':
			sb.WriteString("Monday")
		case 'F':
			sb.WriteString("January")
		case 'm':
			sb.WriteString("01")
		case 'M':
			sb.WriteString("Jan")
		case 'n':
			sb.WriteString("1")
		case 'Y':
			sb.WriteString("2006")
		case 'y':
			sb.WriteString("06")
		case 'a':
			sb.WriteString("pm")
		case 'A':
			sb.WriteString("PM")
		case 'g':
			sb.WriteString("3")
		case 'G', 'H':
			sb.WriteString("15")
		case 'h':
			sb.WriteString("03")
		case 'i':
			sb.WriteString("04")
		case 's':
			sb.WriteString("05")
		case 'v':
			sb.WriteString("000")
		case 'u':
			sb.WriteString("000000")
		case 'O':
			sb.WriteString("-0700")
		case 'P':
			sb.WriteString("-07:00")
		case 'p':
			sb.WriteString("Z07:00")
		case 'T':
			sb.WriteString("MST")
		case 'c':
			sb.WriteString("2006-01-02T15:04:05-07:00")
		case 'r':
			sb.WriteString("Mon, 02 Jan 2006 15:04:05 -0700")
		case 'N', 'S', 'w', 'z', 'W', 't', 'L', 'o', 'I', 'B', 'X', 'Z', 'U', 'e':
			return "", fmt.Errorf("parse: unsupported format specifier %q", string(ch))
		default:
			sb.WriteRune(ch)
		}
	}
	return sb.String(), nil
}

func phpChar(t time.Time, ch rune) string {
	switch ch {
	// Day
	case 'd':
		return t.Format("02")
	case 'D':
		return t.Format("Mon")
	case 'j':
		return t.Format("2")
	case 'l':
		return t.Format("Monday")
	case 'N':
		d := int(t.Weekday())
		if d == 0 {
			d = 7
		}
		return strconv.Itoa(d)
	case 'S':
		switch t.Day() {
		case 1, 21, 31:
			return "st"
		case 2, 22:
			return "nd"
		case 3, 23:
			return "rd"
		default:
			return "th"
		}
	case 'w':
		return strconv.Itoa(int(t.Weekday()))
	case 'z':
		return strconv.Itoa(t.YearDay() - 1)

	// Week
	case 'W':
		_, week := t.ISOWeek()
		return fmt.Sprintf("%02d", week)

	// Month
	case 'F':
		return t.Format("January")
	case 'm':
		return t.Format("01")
	case 'M':
		return t.Format("Jan")
	case 'n':
		return t.Format("1")
	case 't':
		return strconv.Itoa(time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day())

	// Year
	case 'L':
		year := t.Year()
		if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
			return "1"
		}
		return "0"
	case 'o':
		isoYear, _ := t.ISOWeek()
		return strconv.Itoa(isoYear)
	case 'Y':
		return t.Format("2006")
	case 'y':
		return t.Format("06")

	// Time
	case 'a':
		return t.Format("pm")
	case 'A':
		return t.Format("PM")
	case 'B':
		// Swatch internet time: seconds since midnight at UTC+1, divided by 86.4
		utcPlusOne := t.UTC().Add(time.Hour)
		secs := float64(utcPlusOne.Hour()*3600+utcPlusOne.Minute()*60+utcPlusOne.Second()) +
			float64(utcPlusOne.Nanosecond())/1e9
		return fmt.Sprintf("%03d", int(secs/86.4))
	case 'g':
		return t.Format("3")
	case 'G':
		return strconv.Itoa(t.Hour())
	case 'h':
		return t.Format("03")
	case 'H':
		return t.Format("15")
	case 'i':
		return t.Format("04")
	case 's':
		return t.Format("05")
	case 'u':
		return fmt.Sprintf("%06d", t.Nanosecond()/1000)
	case 'v':
		return fmt.Sprintf("%03d", t.Nanosecond()/1_000_000)

	// Timezone
	case 'e':
		return t.Location().String()
	case 'I':
		jan := time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, t.Location())
		jul := time.Date(t.Year(), time.July, 1, 0, 0, 0, 0, t.Location())
		_, janOff := jan.Zone()
		_, julOff := jul.Zone()
		_, curOff := t.Zone()
		if janOff == julOff || curOff != max(janOff, julOff) {
			return "0"
		}
		return "1"
	case 'O':
		return t.Format("-0700")
	case 'P':
		return t.Format("-07:00")
	case 'p':
		_, offset := t.Zone()
		if offset == 0 {
			return "Z"
		}
		return t.Format("-07:00")
	case 'T':
		return t.Format("MST")
	case 'Z':
		_, offset := t.Zone()
		return strconv.Itoa(offset)

	// Full date/time
	case 'c':
		return t.Format("2006-01-02T15:04:05-07:00")
	case 'r':
		return t.Format("Mon, 02 Jan 2006 15:04:05 -0700")
	case 'U':
		return strconv.FormatInt(t.Unix(), 10)

	default:
		return string(ch)
	}
}
