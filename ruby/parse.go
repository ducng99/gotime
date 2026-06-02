package ruby

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Parse parses value using Ruby's Time#strftime format string, the inverse of Format.
func Parse(format, value string) (Time, error) {
	t, err := rubyParse(expandComposites(format), value)
	if err != nil {
		return Time{}, err
	}
	return Time(t), nil
}

func expandComposites(format string) string {
	runes := []rune(format)
	var out []rune
	for i := 0; i < len(runes); i++ {
		if runes[i] == '%' && i+1 < len(runes) {
			switch runes[i+1] {
			case 'c':
				out = append(out, []rune(`%a %b %e %H:%M:%S %Y`)...)
				i++
				continue
			case 'D', 'x':
				out = append(out, []rune(`%m/%d/%y`)...)
				i++
				continue
			case 'F':
				out = append(out, []rune(`%Y-%m-%d`)...)
				i++
				continue
			case 'v':
				out = append(out, []rune(`%e-%^b-%Y`)...)
				i++
				continue
			case 'X', 'T':
				out = append(out, []rune(`%H:%M:%S`)...)
				i++
				continue
			case 'r':
				out = append(out, []rune(`%I:%M:%S %p`)...)
				i++
				continue
			case 'R':
				out = append(out, []rune(`%H:%M`)...)
				i++
				continue
			}
		}
		out = append(out, runes[i])
	}
	return string(out)
}

type rubyParseState struct {
	year  int
	month int
	day   int
	hour  int
	min   int
	sec   int
	nano  int
	loc   *time.Location

	use12h bool
	ampm   int // 0=unset 1=am 2=pm

	century int // -1 if not set

	hasUnix  bool
	unixSecs int64

	hasDayOfYear bool
	dayOfYear    int

	hasIsoWeek  bool
	isoWeekYear int
	isoWeek     int

	hasWeekday bool
	weekday    int // 0=Sun … 6=Sat

	hasWeekNumber bool
	weekNumber    int
	weekBasis     int // 0=Sunday-based (U), 1=Monday-based (W)
}

func newRubyParseState() *rubyParseState {
	return &rubyParseState{year: -1, month: -1, day: -1, hour: -1, min: -1, sec: -1, century: -1}
}

func rubyParse(format, value string) (time.Time, error) {
	fRunes := []rune(format)
	vRunes := []rune(value)
	vi := 0

	ps := newRubyParseState()

	for fi := 0; fi < len(fRunes); fi++ {
		ch := fRunes[fi]

		if ch == '%' {
			fi++
			if fi >= len(fRunes) {
				return time.Time{}, fmt.Errorf("parse: unexpected end of format after '%%'")
			}

			var nopad bool
			colonCount := 0
			for fi < len(fRunes) {
				switch fRunes[fi] {
				case '-':
					nopad = true
					fi++
				case '_', '0':
					fi++
				case '^', '#':
					fi++
				case ':':
					colonCount++
					fi++
				default:
					goto doneFlags
				}
			}
		doneFlags:

			width := 0
			hasWidth := false
			for fi < len(fRunes) && fRunes[fi] >= '0' && fRunes[fi] <= '9' {
				width = width*10 + int(fRunes[fi]-'0')
				hasWidth = true
				fi++
			}

			if fi >= len(fRunes) {
				return time.Time{}, fmt.Errorf("parse: unexpected end of format")
			}

			directive := fRunes[fi]
			rest := vRunes[vi:]

			var (
				consumed int
				err      error
			)

			switch directive {
			case 'Y':
				if nopad {
					consumed, err = rubyParseVarInt(rest, &ps.year, 1, 4)
				} else {
					consumed, err = rubyParseFixedInt(rest, &ps.year, 4)
				}
			case 'y':
				var n int
				if nopad {
					consumed, err = rubyParseVarInt(rest, &n, 1, 2)
				} else {
					consumed, err = rubyParseFixedInt(rest, &n, 2)
				}
				if err == nil {
					ps.year = rubyTwoDigitYear(n)
				}
			case 'C':
				var c int
				consumed, err = rubyParseFixedInt(rest, &c, 2)
				if err == nil {
					ps.century = c
				}

			case 'm':
				if nopad {
					consumed, err = rubyParseVarInt(rest, &ps.month, 1, 2)
				} else {
					consumed, err = rubyParsePaddedInt(rest, &ps.month, 1, 12)
				}
			case 'B':
				consumed, err = rubyParseMonthFull(rest, &ps.month)
			case 'b', 'h':
				consumed, err = rubyParseMonthAbbr(rest, &ps.month)

			case 'd':
				if nopad {
					consumed, err = rubyParseVarInt(rest, &ps.day, 1, 2)
				} else {
					consumed, err = rubyParsePaddedInt(rest, &ps.day, 1, 31)
				}
			case 'e':
				if nopad {
					consumed, err = rubyParseVarInt(rest, &ps.day, 1, 2)
				} else {
					consumed, err = rubyParseSpacePaddedInt(rest, &ps.day, 1, 31)
				}
			case 'j':
				var n int
				consumed, err = rubyParseFixedInt(rest, &n, 3)
				if err == nil {
					ps.dayOfYear = n
					ps.hasDayOfYear = true
				}

			case 'H':
				if nopad {
					consumed, err = rubyParseVarInt(rest, &ps.hour, 0, 23)
				} else {
					consumed, err = rubyParsePaddedInt(rest, &ps.hour, 0, 23)
				}
			case 'k':
				if nopad {
					consumed, err = rubyParseVarInt(rest, &ps.hour, 0, 23)
				} else {
					consumed, err = rubyParseSpacePaddedInt(rest, &ps.hour, 0, 23)
				}
			case 'I':
				if nopad {
					consumed, err = rubyParseVarInt(rest, &ps.hour, 1, 12)
				} else {
					consumed, err = rubyParsePaddedInt(rest, &ps.hour, 1, 12)
				}
			case 'l':
				if nopad {
					consumed, err = rubyParseVarInt(rest, &ps.hour, 1, 12)
				} else {
					consumed, err = rubyParseSpacePaddedInt(rest, &ps.hour, 1, 12)
				}

			case 'p':
				consumed, err = rubyParseAmPm(rest, true, &ps.ampm)
				if err == nil {
					ps.use12h = true
				}
			case 'P':
				consumed, err = rubyParseAmPm(rest, false, &ps.ampm)
				if err == nil {
					ps.use12h = true
				}

			case 'M':
				if nopad {
					consumed, err = rubyParseVarInt(rest, &ps.min, 0, 59)
				} else {
					consumed, err = rubyParsePaddedInt(rest, &ps.min, 0, 59)
				}
			case 'S':
				if nopad {
					consumed, err = rubyParseVarInt(rest, &ps.sec, 0, 59)
				} else {
					consumed, err = rubyParsePaddedInt(rest, &ps.sec, 0, 59)
				}

			case 'L':
				var ms int
				consumed, err = rubyParseFixedInt(rest, &ms, 3)
				if err == nil {
					ps.nano = ms * 1_000_000
				}
			case 'N':
				prec := 9
				if hasWidth && width > 0 {
					prec = width
				}
				if prec > 9 {
					prec = 9
				}
				var ns int
				consumed, err = rubyParseFixedInt(rest, &ns, prec)
				if err == nil {
					for i := prec; i < 9; i++ {
						ns *= 10
					}
					ps.nano = ns
				}

			case 's':
				consumed, err = rubyParseUnix(rest, &ps.unixSecs)
				if err == nil {
					ps.hasUnix = true
				}

			case 'z':
				consumed, err = rubyParseTzOffset(rest, colonCount, &ps.loc)
			case 'Z':
				consumed, err = rubyParseTzAbbr(rest, &ps.loc)

			case 'A':
				consumed, err = rubyParseWeekdayFull(rest, &ps.weekday)
				if err == nil {
					ps.hasWeekday = true
				}
			case 'a':
				consumed, err = rubyParseWeekdayAbbr(rest, &ps.weekday)
				if err == nil {
					ps.hasWeekday = true
				}
			case 'u':
				var n int
				consumed, err = rubyParseVarInt(rest, &n, 1, 1)
				if err == nil {
					if n < 1 || n > 7 {
						err = fmt.Errorf("parse: %%u weekday out of range: %d", n)
					} else {
						ps.weekday = n % 7
						ps.hasWeekday = true
					}
				}
			case 'w':
				var n int
				consumed, err = rubyParseVarInt(rest, &n, 1, 1)
				if err == nil {
					if n < 0 || n > 6 {
						err = fmt.Errorf("parse: %%w weekday out of range: %d", n)
					} else {
						ps.weekday = n
						ps.hasWeekday = true
					}
				}

			case 'U':
				var n int
				consumed, err = rubyParsePaddedInt(rest, &n, 0, 53)
				if err == nil {
					ps.weekNumber = n
					ps.weekBasis = 0
					ps.hasWeekNumber = true
				}
			case 'W':
				var n int
				consumed, err = rubyParsePaddedInt(rest, &n, 0, 53)
				if err == nil {
					ps.weekNumber = n
					ps.weekBasis = 1
					ps.hasWeekNumber = true
				}
			case 'V':
				var n int
				consumed, err = rubyParsePaddedInt(rest, &n, 1, 53)
				if err == nil {
					ps.isoWeek = n
					ps.hasIsoWeek = true
				}
			case 'G':
				consumed, err = rubyParseFixedInt(rest, &ps.isoWeekYear, 4)
				if err == nil {
					ps.hasIsoWeek = true
				}
			case 'g':
				var n int
				consumed, err = rubyParseFixedInt(rest, &n, 2)
				if err == nil {
					ps.isoWeekYear = rubyTwoDigitYear(n)
					ps.hasIsoWeek = true
				}

			case 'n':
				if len(rest) == 0 || rest[0] != '\n' {
					err = fmt.Errorf("parse: expected newline")
				} else {
					consumed = 1
				}
			case 't':
				if len(rest) == 0 || rest[0] != '\t' {
					err = fmt.Errorf("parse: expected tab")
				} else {
					consumed = 1
				}

			default:
				if len(rest) == 0 || rest[0] != directive {
					err = fmt.Errorf("parse: expected %q", string(directive))
				} else {
					consumed = 1
				}
			}

			if err != nil {
				return time.Time{}, err
			}
			vi += consumed
			continue
		}

		if vi >= len(vRunes) || vRunes[vi] != ch {
			return time.Time{}, fmt.Errorf("parse: expected %q at position %d", string(ch), vi)
		}
		vi++
	}

	if vi < len(vRunes) {
		return time.Time{}, fmt.Errorf("parse: unexpected trailing text: %q", string(vRunes[vi:]))
	}

	return ps.build()
}

func (ps *rubyParseState) build() (time.Time, error) {
	loc := ps.loc
	if loc == nil {
		loc = time.UTC
	}

	if ps.hasUnix {
		return time.Unix(ps.unixSecs, 0).In(loc), nil
	}

	year := max(ps.year, 0)

	if ps.century >= 0 {
		y := 0
		if ps.year >= 0 {
			y = ps.year % 100
		}
		year = ps.century*100 + y
	} else if ps.year < 0 && ps.isoWeekYear > 0 {
		year = ps.isoWeekYear
	}

	month := time.Month(ps.month)
	if ps.month < 0 {
		month = time.January
	}
	day := ps.day
	if day < 0 {
		day = 1
	}
	hour := max(ps.hour, 0)
	min := max(ps.min, 0)
	sec := max(ps.sec, 0)

	if ps.use12h {
		switch ps.ampm {
		case 2:
			if hour != 12 {
				hour += 12
			}
		case 1:
			if hour == 12 {
				hour = 0
			}
		}
	}

	if ps.hasIsoWeek && ps.isoWeek > 0 && ps.hasWeekday {
		wy := year
		if ps.isoWeekYear > 0 {
			wy = ps.isoWeekYear
		}
		t0 := isoWeekDate(wy, ps.isoWeek, ps.weekday)
		return time.Date(t0.Year(), t0.Month(), t0.Day(), hour, min, sec, ps.nano, loc), nil
	}

	if ps.hasDayOfYear && ps.month < 0 {
		t0 := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
		t0 = t0.AddDate(0, 0, ps.dayOfYear-1)
		month = t0.Month()
		day = t0.Day()
	}

	if ps.hasWeekNumber && ps.hasWeekday && ps.year >= 0 {
		t0 := rubyWeekToDate(year, ps.weekNumber, ps.weekBasis, ps.weekday)
		return time.Date(t0.Year(), t0.Month(), t0.Day(), hour, min, sec, ps.nano, loc), nil
	}

	return time.Date(year, month, day, hour, min, sec, ps.nano, loc), nil
}

// rubyWeekToDate computes the date from a year, week number, week basis, and weekday.
// weekBasis 0 = Sunday-based (%U), 1 = Monday-based (%W).
// weekday is Go convention: 0=Sun … 6=Sat.
func rubyWeekToDate(year, week, basis, weekday int) time.Time {
	jan1 := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	jan1Wday := int(jan1.Weekday())

	var offset int
	if basis == 0 {
		// %U: Sunday-based. Week 1 contains the first Sunday.
		// Days before the first Sunday are in week 0.
		offset = (7-jan1Wday)%7 + weekday
	} else {
		// %W: Monday-based. Week 1 contains the first Monday.
		// Days before the first Monday are in week 0.
		mondayOffset := (8 - jan1Wday) % 7
		if mondayOffset == 7 {
			mondayOffset = 0
		}
		offset = mondayOffset - 1 + weekday
		if weekday == 0 {
			offset += 7
		}
		offset = (7-jan1Wday)%7 - 1 + weekday
		if offset < 0 {
			offset += 7
		}
	}

	if basis == 0 {
		sundayOffset := (7 - jan1Wday) % 7
		firstSunday := jan1.AddDate(0, 0, sundayOffset)
		return firstSunday.AddDate(0, 0, (week-1)*7+weekday)
	}

	mondayOffset := (8 - jan1Wday) % 7
	if mondayOffset == 7 {
		mondayOffset = 0
	}
	firstMonday := jan1.AddDate(0, 0, mondayOffset)
	wd := weekday
	if wd == 0 {
		wd = 7
	}
	return firstMonday.AddDate(0, 0, (week-1)*7+(wd-1))
}

// --- integer helpers ---

func rubyParseFixedInt(rest []rune, dst *int, width int) (int, error) {
	if len(rest) < width {
		return 0, fmt.Errorf("parse: expected %d-digit number", width)
	}
	n := 0
	for i := range width {
		if !isDigit(rest[i]) {
			return 0, fmt.Errorf("parse: expected digit at position %d, got %q", i, string(rest[i]))
		}
		n = n*10 + int(rest[i]-'0')
	}
	*dst = n
	return width, nil
}

func rubyParseVarInt(rest []rune, dst *int, minWidth, maxWidth int) (int, error) {
	i := 0
	for i < len(rest) && i < maxWidth && isDigit(rest[i]) {
		i++
	}
	if i < minWidth {
		return 0, fmt.Errorf("parse: expected at least %d digit(s)", minWidth)
	}
	n := 0
	for j := 0; j < i; j++ {
		n = n*10 + int(rest[j]-'0')
	}
	*dst = n
	return i, nil
}

func rubyParsePaddedInt(rest []rune, dst *int, minVal, maxVal int) (int, error) {
	if len(rest) < 2 || !isDigit(rest[0]) || !isDigit(rest[1]) {
		s := ""
		if len(rest) > 0 {
			n := min(4, len(rest))
			s = string(rest[:n])
		}
		return 0, fmt.Errorf("parse: expected 2-digit number, got %q", s)
	}
	n := int(rest[0]-'0')*10 + int(rest[1]-'0')
	if n < minVal || n > maxVal {
		return 0, fmt.Errorf("parse: value %d out of range [%d,%d]", n, minVal, maxVal)
	}
	*dst = n
	return 2, nil
}

func rubyParseSpacePaddedInt(rest []rune, dst *int, minVal, maxVal int) (int, error) {
	if len(rest) == 0 {
		return 0, fmt.Errorf("parse: expected number")
	}
	i := 0
	if rest[i] == ' ' {
		i++
	}
	if i >= len(rest) || !isDigit(rest[i]) {
		return 0, fmt.Errorf("parse: expected digit")
	}
	n := int(rest[i] - '0')
	i++
	if i < len(rest) && isDigit(rest[i]) {
		n = n*10 + int(rest[i]-'0')
		i++
	}
	if n < minVal || n > maxVal {
		return 0, fmt.Errorf("parse: value %d out of range [%d,%d]", n, minVal, maxVal)
	}
	*dst = n
	return i, nil
}

// --- string token helpers ---

var (
	rubyWeekdayAbbrs = [7]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	rubyWeekdayFulls = [7]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	rubyMonthAbbrs   = [13]string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	rubyMonthFulls   = [13]string{"", "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
)

func rubyParseWeekdayAbbr(rest []rune, dst *int) (int, error) {
	if len(rest) < 3 {
		return 0, fmt.Errorf("parse: expected weekday abbreviation")
	}
	s := strings.ToUpper(string(rest[:3]))
	for i, name := range rubyWeekdayAbbrs {
		if s == strings.ToUpper(name) {
			*dst = i
			return 3, nil
		}
	}
	return 0, fmt.Errorf("parse: unknown weekday abbreviation %q", string(rest[:3]))
}

func rubyParseWeekdayFull(rest []rune, dst *int) (int, error) {
	for i, name := range rubyWeekdayFulls {
		r := []rune(name)
		if len(rest) >= len(r) && strings.EqualFold(string(rest[:len(r)]), name) {
			*dst = i
			return len(r), nil
		}
	}
	return 0, fmt.Errorf("parse: expected full weekday name")
}

func rubyParseMonthAbbr(rest []rune, dst *int) (int, error) {
	if len(rest) < 3 {
		return 0, fmt.Errorf("parse: expected month abbreviation")
	}
	s := strings.ToUpper(string(rest[:3]))
	for i := 1; i <= 12; i++ {
		if s == strings.ToUpper(rubyMonthAbbrs[i]) {
			*dst = i
			return 3, nil
		}
	}
	return 0, fmt.Errorf("parse: unknown month abbreviation %q", string(rest[:3]))
}

func rubyParseMonthFull(rest []rune, dst *int) (int, error) {
	for i := 1; i <= 12; i++ {
		name := rubyMonthFulls[i]
		r := []rune(name)
		if len(rest) >= len(r) && strings.EqualFold(string(rest[:len(r)]), name) {
			*dst = i
			return len(r), nil
		}
	}
	return 0, fmt.Errorf("parse: expected full month name")
}

func rubyParseAmPm(rest []rune, upper bool, dst *int) (int, error) {
	if len(rest) < 2 {
		return 0, fmt.Errorf("parse: expected AM/PM")
	}
	s := string(rest[:2])
	if upper {
		switch s {
		case "AM":
			*dst = 1
			return 2, nil
		case "PM":
			*dst = 2
			return 2, nil
		}
	} else {
		switch s {
		case "am":
			*dst = 1
			return 2, nil
		case "pm":
			*dst = 2
			return 2, nil
		}
	}
	return 0, fmt.Errorf("parse: expected AM/PM, got %q", s)
}

// --- timezone helpers ---

func rubyParseTzOffset(rest []rune, colonCount int, dst **time.Location) (int, error) {
	if len(rest) == 0 {
		return 0, fmt.Errorf("parse: expected timezone offset")
	}
	sign := 1
	i := 0
	switch rest[i] {
	case '+':
		i++
	case '-':
		sign = -1
		i++
	default:
		return 0, fmt.Errorf("parse: expected '+' or '-' in timezone offset")
	}

	switch colonCount {
	case 0:
		// +HHMM
		if i+4 > len(rest) {
			return 0, fmt.Errorf("parse: expected ±HHMM timezone offset")
		}
		for j := i; j < i+4; j++ {
			if !isDigit(rest[j]) {
				return 0, fmt.Errorf("parse: expected digit in timezone offset")
			}
		}
		hh := int(rest[i]-'0')*10 + int(rest[i+1]-'0')
		mm := int(rest[i+2]-'0')*10 + int(rest[i+3]-'0')
		*dst = time.FixedZone("", sign*(hh*3600+mm*60))
		return i + 4, nil
	case 1:
		// +HH:MM
		if i+5 > len(rest) {
			return 0, fmt.Errorf("parse: expected ±HH:MM timezone offset")
		}
		if !isDigit(rest[i]) || !isDigit(rest[i+1]) || rest[i+2] != ':' || !isDigit(rest[i+3]) || !isDigit(rest[i+4]) {
			return 0, fmt.Errorf("parse: invalid ±HH:MM timezone offset")
		}
		hh := int(rest[i]-'0')*10 + int(rest[i+1]-'0')
		mm := int(rest[i+3]-'0')*10 + int(rest[i+4]-'0')
		*dst = time.FixedZone("", sign*(hh*3600+mm*60))
		return i + 5, nil
	default:
		// +HH:MM:SS
		if i+8 > len(rest) {
			return 0, fmt.Errorf("parse: expected ±HH:MM:SS timezone offset")
		}
		if !isDigit(rest[i]) || !isDigit(rest[i+1]) || rest[i+2] != ':' ||
			!isDigit(rest[i+3]) || !isDigit(rest[i+4]) || rest[i+5] != ':' ||
			!isDigit(rest[i+6]) || !isDigit(rest[i+7]) {
			return 0, fmt.Errorf("parse: invalid ±HH:MM:SS timezone offset")
		}
		hh := int(rest[i]-'0')*10 + int(rest[i+1]-'0')
		mm := int(rest[i+3]-'0')*10 + int(rest[i+4]-'0')
		ss := int(rest[i+6]-'0')*10 + int(rest[i+7]-'0')
		*dst = time.FixedZone("", sign*(hh*3600+mm*60+ss))
		return i + 8, nil
	}
}

func rubyParseTzAbbr(rest []rune, dst **time.Location) (int, error) {
	i := 0
	for i < len(rest) && isLetter(rest[i]) {
		i++
	}
	if i == 0 {
		return 0, fmt.Errorf("parse: expected timezone abbreviation")
	}
	name := string(rest[:i])
	t, err := time.Parse("MST", name)
	if err != nil {
		*dst = time.FixedZone(name, 0)
	} else {
		*dst = t.Location()
	}
	return i, nil
}

func rubyParseUnix(rest []rune, dst *int64) (int, error) {
	i := 0
	if i < len(rest) && (rest[i] == '-' || rest[i] == '+') {
		i++
	}
	start := i
	for i < len(rest) && isDigit(rest[i]) {
		i++
	}
	if i == start {
		return 0, fmt.Errorf("parse: expected Unix timestamp")
	}
	n, err := strconv.ParseInt(string(rest[:i]), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse: invalid Unix timestamp: %w", err)
	}
	*dst = n
	return i, nil
}

// --- utilities ---

func rubyTwoDigitYear(y int) int {
	if y >= 69 {
		return 1900 + y
	}
	return 2000 + y
}

func isoWeekDate(year, week, weekday int) time.Time {
	jan4 := time.Date(year, time.January, 4, 0, 0, 0, 0, time.UTC)
	jan4ISO := int(jan4.Weekday())
	if jan4ISO == 0 {
		jan4ISO = 7
	}
	week1Mon := jan4.AddDate(0, 0, 1-jan4ISO)
	targetISO := weekday
	if targetISO == 0 {
		targetISO = 7
	}
	return week1Mon.AddDate(0, 0, (week-1)*7+(targetISO-1))
}

func isDigit(r rune) bool  { return r >= '0' && r <= '9' }
func isLetter(r rune) bool { return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') }
