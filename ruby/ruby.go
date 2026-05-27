package ruby

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type RubyTime time.Time

// Format formats using Ruby's Time#strftime format string.
// See https://docs.ruby-lang.org/en/master/strftime_formatting_rdoc.html
func (t *RubyTime) Format(format string) string {
	goTime := time.Time(*t)
	var out strings.Builder

	runes := []rune(format)
	i := 0
	for i < len(runes) {
		ch := runes[i]
		if ch != '%' {
			out.WriteRune(ch)
			i++
			continue
		}
		i++ // consume '%'
		if i >= len(runes) {
			out.WriteByte('%')
			break
		}

		// Parse flags, tracking raw characters for unknown-directive passthrough.
		var raw strings.Builder
		var nopad, spacePad, zeroPad, upper, swap bool
		colonCount := 0
		flagsDone := false
		for i < len(runes) && !flagsDone {
			switch runes[i] {
			case '-':
				nopad = true
				raw.WriteByte('-')
				i++
			case '_':
				spacePad = true
				raw.WriteByte('_')
				i++
			case '0':
				zeroPad = true
				raw.WriteByte('0')
				i++
			case '^':
				upper = true
				raw.WriteByte('^')
				i++
			case '#':
				swap = true
				raw.WriteByte('#')
				i++
			case ':':
				colonCount++
				raw.WriteByte(':')
				i++
			default:
				flagsDone = true
			}
		}

		// Parse optional width.
		width := 0
		hasWidth := false
		for i < len(runes) && runes[i] >= '0' && runes[i] <= '9' {
			width = width*10 + int(runes[i]-'0')
			raw.WriteRune(runes[i])
			hasWidth = true
			i++
		}

		if i >= len(runes) {
			out.WriteByte('%')
			out.WriteString(raw.String())
			break
		}

		directive := runes[i]
		i++

		if directive == '%' {
			out.WriteByte('%')
			continue
		}

		value, defPad, defWidth := rubyValue(goTime, directive, colonCount, hasWidth, width)

		// Unknown directive: emit %flags+width+directive verbatim.
		if value == "" {
			out.WriteByte('%')
			out.WriteString(raw.String())
			out.WriteRune(directive)
			continue
		}

		// Apply case modifiers.
		if upper {
			value = strings.ToUpper(value)
		} else if swap {
			value = swapCaseStr(value)
		}

		// Determine effective padding.
		effectivePad := defPad
		effectiveWidth := defWidth
		if hasWidth && directive != 'N' {
			effectiveWidth = width
			if effectivePad == 0 {
				effectivePad = ' '
			}
		}
		if nopad {
			effectivePad = 0
		} else if zeroPad && effectivePad != 0 {
			effectivePad = '0'
		} else if spacePad && effectivePad != 0 {
			effectivePad = ' '
		}

		// Apply padding.
		rval := []rune(value)
		if effectivePad != 0 && effectiveWidth > 0 && len(rval) < effectiveWidth {
			padStr := strings.Repeat(string(effectivePad), effectiveWidth-len(rval))
			if effectivePad == '0' && len(rval) > 0 && (rval[0] == '+' || rval[0] == '-') {
				out.WriteRune(rval[0])
				out.WriteString(padStr)
				out.WriteString(string(rval[1:]))
			} else {
				out.WriteString(padStr)
				out.WriteString(value)
			}
			continue
		}

		out.WriteString(value)
	}

	return out.String()
}

// swapCaseStr uppercases s if it contains any lowercase letter; otherwise lowercases it.
func swapCaseStr(s string) string {
	for _, r := range s {
		if unicode.IsLower(r) {
			return strings.ToUpper(s)
		}
	}
	return strings.ToLower(s)
}

// rubyValue returns the raw (unpadded) value for a directive, along with the
// default pad character (0 = no padding) and default minimum width.
// Returns ("", 0, 0) for unknown directives.
func rubyValue(t time.Time, directive rune, colonCount int, hasWidthSpec bool, widthSpec int) (string, rune, int) {
	switch directive {
	// Year
	case 'Y':
		return strconv.Itoa(t.Year()), '0', 4
	case 'y':
		return t.Format("06"), '0', 2
	case 'C':
		return strconv.Itoa(t.Year() / 100), '0', 2

	// Month
	case 'm':
		return strconv.Itoa(int(t.Month())), '0', 2
	case 'B':
		return t.Format("January"), 0, 0
	case 'b', 'h':
		return t.Format("Jan"), 0, 0

	// Day
	case 'd':
		return strconv.Itoa(t.Day()), '0', 2
	case 'e':
		return strconv.Itoa(t.Day()), ' ', 2
	case 'j':
		return strconv.Itoa(t.YearDay()), '0', 3

	// Hour
	case 'H':
		return strconv.Itoa(t.Hour()), '0', 2
	case 'k':
		return strconv.Itoa(t.Hour()), ' ', 2
	case 'I':
		h := t.Hour() % 12
		if h == 0 {
			h = 12
		}
		return strconv.Itoa(h), '0', 2
	case 'l':
		h := t.Hour() % 12
		if h == 0 {
			h = 12
		}
		return strconv.Itoa(h), ' ', 2
	case 'P':
		return t.Format("pm"), 0, 0
	case 'p':
		return t.Format("PM"), 0, 0

	// Minute, second, subsecond
	case 'M':
		return strconv.Itoa(t.Minute()), '0', 2
	case 'S':
		return strconv.Itoa(t.Second()), '0', 2
	case 'L':
		return strconv.Itoa(t.Nanosecond() / 1_000_000), '0', 3
	case 'N':
		// Width controls fractional-second precision; default is 9 (nanoseconds).
		prec := 9
		if hasWidthSpec && widthSpec > 0 {
			prec = widthSpec
		}
		nano := fmt.Sprintf("%09d", t.Nanosecond())
		if prec <= 9 {
			return nano[:prec], '0', prec
		}
		return nano + strings.Repeat("0", prec-9), '0', prec
	case 's':
		return strconv.FormatInt(t.Unix(), 10), 0, 0

	// Timezone
	case 'z':
		_, offset := t.Zone()
		sign := "+"
		if offset < 0 {
			sign = "-"
			offset = -offset
		}
		h, m, s := offset/3600, (offset%3600)/60, offset%60
		switch colonCount {
		case 1:
			return fmt.Sprintf("%s%02d:%02d", sign, h, m), '0', 6
		case 2:
			return fmt.Sprintf("%s%02d:%02d:%02d", sign, h, m, s), '0', 9
		default:
			return fmt.Sprintf("%s%02d%02d", sign, h, m), '0', 5
		}
	case 'Z':
		name, _ := t.Zone()
		return name, 0, 0

	// Weekday
	case 'A':
		return t.Format("Monday"), 0, 0
	case 'a':
		return t.Format("Mon"), 0, 0
	case 'u':
		d := int(t.Weekday())
		if d == 0 {
			d = 7
		}
		return strconv.Itoa(d), '0', 1
	case 'w':
		return strconv.Itoa(int(t.Weekday())), '0', 1

	// Week numbers (non-ISO)
	case 'U':
		return strconv.Itoa(weekNumberSunday(t)), '0', 2
	case 'W':
		return strconv.Itoa(weekNumberMonday(t)), '0', 2

	// ISO 8601 week-date
	case 'G':
		isoYear, _ := t.ISOWeek()
		return strconv.Itoa(isoYear), '0', 4
	case 'g':
		isoYear, _ := t.ISOWeek()
		return strconv.Itoa(isoYear % 100), '0', 2
	case 'V':
		_, isoWeek := t.ISOWeek()
		return strconv.Itoa(isoWeek), '0', 2

	// Literals
	case 'n':
		return "\n", 0, 0
	case 't':
		return "\t", 0, 0

	// Composite directives
	case 'c':
		rt := RubyTime(t)
		return rt.Format("%a %b %e %H:%M:%S %Y"), 0, 0
	case 'D', 'x':
		rt := RubyTime(t)
		return rt.Format("%m/%d/%y"), 0, 0
	case 'F':
		rt := RubyTime(t)
		return rt.Format("%Y-%m-%d"), 0, 0
	case 'v':
		rt := RubyTime(t)
		return rt.Format("%e-%^b-%Y"), 0, 0
	case 'X', 'T':
		rt := RubyTime(t)
		return rt.Format("%H:%M:%S"), 0, 0
	case 'r':
		rt := RubyTime(t)
		return rt.Format("%I:%M:%S %p"), 0, 0
	case 'R':
		rt := RubyTime(t)
		return rt.Format("%H:%M"), 0, 0

	default:
		return "", 0, 0
	}
}

// weekNumberSunday returns the %U week number (Sunday-based, 0-53).
// Days before the first Sunday of the year are in week 0.
func weekNumberSunday(t time.Time) int {
	jan1 := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
	jan1Wday := int(jan1.Weekday()) // 0=Sun
	return (t.YearDay() + 6 - (7-jan1Wday)%7) / 7
}

// weekNumberMonday returns the %W week number (Monday-based, 0-53).
// Days before the first Monday of the year are in week 0.
func weekNumberMonday(t time.Time) int {
	jan1 := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
	jan1Wday := int(jan1.Weekday()) // 0=Sun
	return (t.YearDay() + 6 - (8-jan1Wday)%7) / 7
}
