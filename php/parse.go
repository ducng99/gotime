package php

import (
	"fmt"
	"strconv"
	"time"
)

// Parse parses value using PHP's date() format string, the inverse of Format.
// Tokens that produce non-reversible output (X) return an error.
func Parse(format, value string) (Time, error) {
	t, err := phpParse(expandComposites(format), value)
	if err != nil {
		return Time{}, err
	}
	return Time(t), nil
}

// expandComposites replaces composite tokens c and r with their equivalent formats.
func expandComposites(format string) string {
	runes := []rune(format)
	var out []rune
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\\' && i+1 < len(runes) {
			out = append(out, runes[i], runes[i+1])
			i++
			continue
		}
		switch runes[i] {
		case 'c':
			out = append(out, []rune(`Y-m-d\TH:i:sP`)...)
		case 'r':
			out = append(out, []rune(`D, d M Y H:i:s O`)...)
		default:
			out = append(out, runes[i])
		}
	}
	return string(out)
}

type phpParseState struct {
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

	hasUnix  bool
	unixSecs int64

	hasDayOfYear bool
	dayOfYear    int // PHP z is 0-based

	hasIsoWeek  bool
	isoWeekYear int
	isoWeek     int

	hasWeekday bool
	weekday    int // 0=Sun … 6=Sat

	hasBeat bool
	beat    int
}

func newParseState() *phpParseState {
	return &phpParseState{year: -1, month: -1, day: -1, hour: -1, min: -1, sec: -1}
}

func phpParse(format, value string) (time.Time, error) {
	fRunes := []rune(format)
	vRunes := []rune(value)
	vi := 0

	ps := newParseState()

	for fi := 0; fi < len(fRunes); fi++ {
		ch := fRunes[fi]

		if ch == '\\' && fi+1 < len(fRunes) {
			fi++
			if vi >= len(vRunes) || vRunes[vi] != fRunes[fi] {
				return time.Time{}, fmt.Errorf("parse: expected literal %q", string(fRunes[fi]))
			}
			vi++
			continue
		}

		var (
			consumed int
			err      error
		)
		rest := vRunes[vi:]

		switch ch {
		case 'd':
			consumed, err = parsePaddedInt(rest, &ps.day, 1, 31)
		case 'j':
			consumed, err = parseVarInt(rest, &ps.day, 1, 2, 1, 31)
		case 'D':
			consumed, err = parseWeekdayAbbr(rest, &ps.weekday)
			if err == nil {
				ps.hasWeekday = true
			}
		case 'l':
			consumed, err = parseWeekdayFull(rest, &ps.weekday)
			if err == nil {
				ps.hasWeekday = true
			}
		case 'N': // ISO weekday 1=Mon…7=Sun
			var n int
			consumed, err = parseVarInt(rest, &n, 1, 1, 1, 7)
			if err == nil {
				ps.weekday = n % 7 // 7→0 (Sun), 1→1 (Mon) …
				ps.hasWeekday = true
			}
		case 'w': // weekday 0=Sun…6=Sat
			var n int
			consumed, err = parseVarInt(rest, &n, 1, 1, 0, 6)
			if err == nil {
				ps.weekday = n
				ps.hasWeekday = true
			}
		case 'z': // day of year 0-based, 1–3 digits
			consumed, err = parseVarInt(rest, &ps.dayOfYear, 1, 3, 0, 365)
			if err == nil {
				ps.hasDayOfYear = true
			}
		case 'S': // ordinal suffix st/nd/rd/th — consume and discard
			consumed, err = parseOrdinal(rest)
		case 't': // days in month — consume and discard
			var dummy int
			consumed, err = parseFixedInt(rest, &dummy, 2)
		case 'L': // leap year 0/1 — consume and discard
			consumed, err = consumeDigit(rest)
		case 'W': // ISO week number, 2 digits
			consumed, err = parseFixedInt(rest, &ps.isoWeek, 2)
			if err == nil {
				ps.hasIsoWeek = true
			}
		case 'm':
			consumed, err = parsePaddedInt(rest, &ps.month, 1, 12)
		case 'n':
			consumed, err = parseVarInt(rest, &ps.month, 1, 2, 1, 12)
		case 'F':
			consumed, err = parseMonthFull(rest, &ps.month)
		case 'M':
			consumed, err = parseMonthAbbr(rest, &ps.month)
		case 'Y':
			consumed, err = parseFixedInt(rest, &ps.year, 4)
		case 'y':
			consumed, err = parseFixedInt(rest, &ps.year, 2)
			if err == nil {
				ps.year = twoDigitYear(ps.year)
			}
		case 'o': // ISO week year, 4 digits
			consumed, err = parseFixedInt(rest, &ps.isoWeekYear, 4)
			if err == nil {
				ps.hasIsoWeek = true
			}
		case 'X': // expanded year with sign: +YYYY or -YYYY
			consumed, err = parseExpandedYear(rest, &ps.year)
		case 'G':
			consumed, err = parseVarInt(rest, &ps.hour, 1, 2, 0, 23)
		case 'H':
			consumed, err = parsePaddedInt(rest, &ps.hour, 0, 23)
		case 'g':
			consumed, err = parseVarInt(rest, &ps.hour, 1, 2, 1, 12)
			if err == nil {
				ps.use12h = true
			}
		case 'h':
			consumed, err = parsePaddedInt(rest, &ps.hour, 1, 12)
			if err == nil {
				ps.use12h = true
			}
		case 'a':
			consumed, err = parseAmPm(rest, false, &ps.ampm)
		case 'A':
			consumed, err = parseAmPm(rest, true, &ps.ampm)
		case 'B': // Swatch Internet Time 000–999
			consumed, err = parseFixedInt(rest, &ps.beat, 3)
			if err == nil {
				ps.hasBeat = true
			}
		case 'i':
			consumed, err = parsePaddedInt(rest, &ps.min, 0, 59)
		case 's':
			consumed, err = parsePaddedInt(rest, &ps.sec, 0, 59)
		case 'v': // milliseconds, 3 digits
			var ms int
			consumed, err = parseFixedInt(rest, &ms, 3)
			if err == nil {
				ps.nano = ms * 1_000_000
			}
		case 'u': // microseconds, 6 digits
			var us int
			consumed, err = parseFixedInt(rest, &us, 6)
			if err == nil {
				ps.nano = us * 1_000
			}
		case 'e':
			consumed, err = parseTzIdentifier(rest, &ps.loc)
		case 'T':
			consumed, err = parseTzAbbr(rest, &ps.loc)
		case 'O':
			consumed, err = parseTzOffsetHHMM(rest, &ps.loc)
		case 'P':
			consumed, err = parseTzOffsetColon(rest, &ps.loc)
		case 'p':
			consumed, err = parseTzOffsetZ(rest, &ps.loc)
		case 'Z': // timezone offset in seconds
			consumed, err = parseTzSeconds(rest, &ps.loc)
		case 'I': // DST indicator 0/1 — consume and discard
			consumed, err = consumeDigit(rest)
		case 'U':
			consumed, err = parseUnix(rest, &ps.unixSecs)
			if err == nil {
				ps.hasUnix = true
			}
		default:
			if len(rest) == 0 || rest[0] != ch {
				err = fmt.Errorf("parse: expected %q", string(ch))
			} else {
				consumed = 1
			}
		}

		if err != nil {
			return time.Time{}, err
		}
		vi += consumed
	}

	if vi < len(vRunes) {
		return time.Time{}, fmt.Errorf("parse: unexpected trailing text: %q", string(vRunes[vi:]))
	}

	return ps.build()
}

func (ps *phpParseState) build() (time.Time, error) {
	loc := ps.loc
	if loc == nil {
		loc = time.UTC
	}

	if ps.hasUnix {
		return time.Unix(ps.unixSecs, 0).In(loc), nil
	}

	year := max(ps.year, 0)
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

	// Day of year overrides month/day when those weren't supplied
	if ps.hasDayOfYear && ps.month < 0 {
		t0 := time.Date(year, time.January, 1+ps.dayOfYear, 0, 0, 0, 0, time.UTC)
		month = t0.Month()
		day = t0.Day()
	}

	// ISO week date (o + W + weekday)
	if ps.hasIsoWeek && ps.isoWeek > 0 && ps.hasWeekday {
		wy := year
		if ps.isoWeekYear > 0 {
			wy = ps.isoWeekYear
		}
		t0 := isoWeekDate(wy, ps.isoWeek, ps.weekday)
		year = t0.Year()
		month = t0.Month()
		day = t0.Day()
	} else if ps.isoWeekYear > 0 && ps.year < 0 {
		year = ps.isoWeekYear
	}

	// Swatch Internet Time: beat * 86.4 seconds from BMT midnight (UTC+1)
	if ps.hasBeat {
		beatTenths := ps.beat * 864 // beat * 86.4 * 10 (integer arithmetic)
		beatSec := beatTenths / 10
		beatNano := (beatTenths % 10) * 100_000_000
		bmt := time.FixedZone("BMT", 3600)
		return time.Date(year, month, day, 0, 0, beatSec, beatNano, bmt).In(loc), nil
	}

	return time.Date(year, month, day, hour, min, sec, ps.nano, loc), nil
}

// --- integer helpers ---

func parseExpandedYear(rest []rune, dst *int) (int, error) {
	if len(rest) < 5 {
		return 0, fmt.Errorf("parse: expected expanded year (+YYYY or -YYYY)")
	}
	sign, err := parseSign(rest[0])
	if err != nil {
		return 0, fmt.Errorf("parse: expected expanded year (+YYYY or -YYYY)")
	}
	i := 1
	for i < len(rest) && isDigit(rest[i]) {
		i++
	}
	if i < 5 {
		return 0, fmt.Errorf("parse: expected at least 4 digits in expanded year")
	}
	n := 0
	for j := 1; j < i; j++ {
		n = n*10 + int(rest[j]-'0')
	}
	*dst = sign * n
	return i, nil
}

func parsePaddedInt(rest []rune, dst *int, minVal, maxVal int) (int, error) {
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

func parseFixedInt(rest []rune, dst *int, width int) (int, error) {
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

func parseVarInt(rest []rune, dst *int, minWidth, maxWidth, minVal, maxVal int) (int, error) {
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
	if n < minVal || n > maxVal {
		return 0, fmt.Errorf("parse: value %d out of range [%d,%d]", n, minVal, maxVal)
	}
	*dst = n
	return i, nil
}

func consumeDigit(rest []rune) (int, error) {
	if len(rest) == 0 || !isDigit(rest[0]) {
		return 0, fmt.Errorf("parse: expected digit")
	}
	return 1, nil
}

// --- string token helpers ---

var (
	weekdayAbbrs = [7]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	weekdayFulls = [7]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	monthAbbrs   = [13]string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	monthFulls   = [13]string{"", "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
)

func parseWeekdayAbbr(rest []rune, dst *int) (int, error) {
	if len(rest) < 3 {
		return 0, fmt.Errorf("parse: expected weekday abbreviation")
	}
	s := string(rest[:3])
	for i, name := range weekdayAbbrs {
		if s == name {
			*dst = i
			return 3, nil
		}
	}
	return 0, fmt.Errorf("parse: unknown weekday abbreviation %q", s)
}

func parseWeekdayFull(rest []rune, dst *int) (int, error) {
	for i, name := range weekdayFulls {
		r := []rune(name)
		if len(rest) >= len(r) && string(rest[:len(r)]) == name {
			*dst = i
			return len(r), nil
		}
	}
	return 0, fmt.Errorf("parse: expected full weekday name")
}

func parseMonthAbbr(rest []rune, dst *int) (int, error) {
	if len(rest) < 3 {
		return 0, fmt.Errorf("parse: expected month abbreviation")
	}
	s := string(rest[:3])
	for i := 1; i <= 12; i++ {
		if s == monthAbbrs[i] {
			*dst = i
			return 3, nil
		}
	}
	return 0, fmt.Errorf("parse: unknown month abbreviation %q", s)
}

func parseMonthFull(rest []rune, dst *int) (int, error) {
	for i := 1; i <= 12; i++ {
		name := monthFulls[i]
		r := []rune(name)
		if len(rest) >= len(r) && string(rest[:len(r)]) == name {
			*dst = i
			return len(r), nil
		}
	}
	return 0, fmt.Errorf("parse: expected full month name")
}

func parseOrdinal(rest []rune) (int, error) {
	if len(rest) < 2 {
		return 0, fmt.Errorf("parse: expected ordinal suffix")
	}
	s := string(rest[:2])
	switch s {
	case "st", "nd", "rd", "th":
		return 2, nil
	}
	return 0, fmt.Errorf("parse: expected ordinal suffix (st/nd/rd/th), got %q", s)
}

func parseAmPm(rest []rune, upper bool, dst *int) (int, error) {
	if len(rest) < 2 {
		return 0, fmt.Errorf("parse: expected am/pm")
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

func parseTzOffsetHHMM(rest []rune, dst **time.Location) (int, error) {
	if len(rest) < 5 {
		return 0, fmt.Errorf("parse: expected ±HHMM timezone offset")
	}
	sign, err := parseSign(rest[0])
	if err != nil {
		return 0, err
	}
	for i := 1; i <= 4; i++ {
		if !isDigit(rest[i]) {
			return 0, fmt.Errorf("parse: expected digit in timezone offset")
		}
	}
	hh := int(rest[1]-'0')*10 + int(rest[2]-'0')
	mm := int(rest[3]-'0')*10 + int(rest[4]-'0')
	*dst = time.FixedZone("", sign*(hh*3600+mm*60))
	return 5, nil
}

func parseTzOffsetColon(rest []rune, dst **time.Location) (int, error) {
	if len(rest) < 6 {
		return 0, fmt.Errorf("parse: expected ±HH:MM timezone offset")
	}
	sign, err := parseSign(rest[0])
	if err != nil {
		return 0, err
	}
	if !isDigit(rest[1]) || !isDigit(rest[2]) || rest[3] != ':' || !isDigit(rest[4]) || !isDigit(rest[5]) {
		return 0, fmt.Errorf("parse: invalid ±HH:MM timezone offset")
	}
	hh := int(rest[1]-'0')*10 + int(rest[2]-'0')
	mm := int(rest[4]-'0')*10 + int(rest[5]-'0')
	*dst = time.FixedZone("", sign*(hh*3600+mm*60))
	return 6, nil
}

func parseTzOffsetZ(rest []rune, dst **time.Location) (int, error) {
	if len(rest) > 0 && rest[0] == 'Z' {
		*dst = time.UTC
		return 1, nil
	}
	return parseTzOffsetColon(rest, dst)
}

func parseTzSeconds(rest []rune, dst **time.Location) (int, error) {
	i := 0
	sign := 1
	if i < len(rest) && rest[i] == '-' {
		sign = -1
		i++
	} else if i < len(rest) && rest[i] == '+' {
		i++
	}
	start := i
	for i < len(rest) && isDigit(rest[i]) {
		i++
	}
	if i == start {
		return 0, fmt.Errorf("parse: expected integer for timezone offset in seconds")
	}
	n, _ := strconv.Atoi(string(rest[start:i]))
	*dst = time.FixedZone("", sign*n)
	return i, nil
}

func parseTzAbbr(rest []rune, dst **time.Location) (int, error) {
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

func parseTzIdentifier(rest []rune, dst **time.Location) (int, error) {
	i := 0
	for i < len(rest) && (isLetter(rest[i]) || isDigit(rest[i]) || rest[i] == '/' || rest[i] == '_' || rest[i] == '-' || rest[i] == '+') {
		i++
	}
	if i == 0 {
		return 0, fmt.Errorf("parse: expected timezone identifier")
	}
	name := string(rest[:i])
	loc, err := time.LoadLocation(name)
	if err != nil {
		return 0, fmt.Errorf("parse: unknown timezone %q", name)
	}
	*dst = loc
	return i, nil
}

// --- Unix timestamp ---

func parseUnix(rest []rune, dst *int64) (int, error) {
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

func parseSign(r rune) (int, error) {
	switch r {
	case '+':
		return 1, nil
	case '-':
		return -1, nil
	}
	return 0, fmt.Errorf("parse: expected '+' or '-', got %q", string(r))
}

func twoDigitYear(y int) int {
	if y >= 69 {
		return 1900 + y
	}
	return 2000 + y
}

// isoWeekDate returns the date for the given ISO year, week, and weekday (0=Sun…6=Sat).
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
