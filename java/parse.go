package java

import (
	"fmt"
	"strconv"
	"time"
)

// Parse parses value using Java's SimpleDateFormat/DateTimeFormatter pattern, the inverse of Format.
func Parse(format, value string) (Time, error) {
	t, err := javaParse(format, value)
	if err != nil {
		return Time{}, err
	}
	return Time(t), nil
}

type javaParseState struct {
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
	hourK  bool // hour came from K (0-11), not h (1-12)

	hasDayOfYear bool
	dayOfYear    int // 1-based (Java D is 1-based)

	hasIsoWeek  bool
	isoWeekYear int
	isoWeek     int

	hasWeekday bool
	weekday    int // Go convention: 0=Sun…6=Sat

	hasMsOfDay bool
	msOfDay    int64

	hasNanoOfDay bool
	nanoOfDay    int64
}

func newJavaParseState() *javaParseState {
	return &javaParseState{year: -1, month: -1, day: -1, hour: -1, min: -1, sec: -1}
}

func (ps *javaParseState) build() (time.Time, error) {
	loc := ps.loc
	if loc == nil {
		loc = time.UTC
	}

	// Nano-of-day overrides all time fields.
	if ps.hasNanoOfDay {
		n := ps.nanoOfDay
		h := int(n / 3_600_000_000_000)
		n %= 3_600_000_000_000
		m := int(n / 60_000_000_000)
		n %= 60_000_000_000
		s := int(n / 1_000_000_000)
		ns := int(n % 1_000_000_000)
		ps.hour, ps.min, ps.sec, ps.nano = h, m, s, ns
	} else if ps.hasMsOfDay {
		ms := ps.msOfDay
		h := int(ms / 3_600_000)
		ms %= 3_600_000
		m := int(ms / 60_000)
		ms %= 60_000
		s := int(ms / 1000)
		ns := int((ms % 1000) * 1_000_000)
		ps.hour, ps.min, ps.sec, ps.nano = h, m, s, ns
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
		if ps.hourK {
			// K is 0-11; PM adds 12.
			if ps.ampm == 2 {
				hour += 12
			}
		} else {
			// h is 1-12; 12 AM → 0, anything PM → +12.
			hour = hour % 12
			if ps.ampm == 2 {
				hour += 12
			}
		}
	}

	// ISO week date: Y+w+weekday → full calendar date.
	if ps.hasIsoWeek && ps.isoWeek > 0 && ps.hasWeekday {
		wy := year
		if ps.isoWeekYear > 0 {
			wy = ps.isoWeekYear
		}
		t0 := isoWeekDate(wy, ps.isoWeek, ps.weekday)
		return time.Date(t0.Year(), t0.Month(), t0.Day(), hour, min, sec, ps.nano, loc), nil
	}
	// Fallback: Y alone → use as calendar year.
	if ps.year < 0 && ps.isoWeekYear > 0 {
		year = ps.isoWeekYear
	}

	// Day of year: D+year → full calendar date.
	if ps.hasDayOfYear {
		return time.Date(year, time.January, ps.dayOfYear, hour, min, sec, ps.nano, loc), nil
	}

	return time.Date(year, month, day, hour, min, sec, ps.nano, loc), nil
}

func javaParse(format, value string) (time.Time, error) {
	fRunes := []rune(format)
	vRunes := []rune(value)
	vi := 0
	ps := newJavaParseState()

	for fi := 0; fi < len(fRunes); fi++ {
		ch := fRunes[fi]

		// Quoted literal: '' = literal single-quote; 'text' = literal text.
		if ch == '\'' {
			if fi+1 < len(fRunes) && fRunes[fi+1] == '\'' {
				if vi >= len(vRunes) || vRunes[vi] != '\'' {
					return time.Time{}, fmt.Errorf("parse: expected \"'\" at position %d", vi)
				}
				vi++
				fi++
				continue
			}
			fi++
			for fi < len(fRunes) {
				if fRunes[fi] == '\'' {
					if fi+1 < len(fRunes) && fRunes[fi+1] == '\'' {
						if vi >= len(vRunes) || vRunes[vi] != '\'' {
							return time.Time{}, fmt.Errorf("parse: expected \"'\" at position %d", vi)
						}
						vi++
						fi += 2
					} else {
						fi++
						break
					}
				} else {
					if vi >= len(vRunes) || vRunes[vi] != fRunes[fi] {
						return time.Time{}, fmt.Errorf("parse: expected %q at position %d", string(fRunes[fi]), vi)
					}
					vi++
					fi++
				}
			}
			fi-- // outer loop will fi++
			continue
		}

		// Find run of identical pattern letters.
		j := fi + 1
		for j < len(fRunes) && fRunes[j] == ch {
			j++
		}
		count := j - fi

		var (
			consumed int
			err      error
		)
		rest := vRunes[vi:]

		switch ch {
		case 'G': // Era — consume and discard.
			consumed, err = javaParseEra(rest, count)

		case 'y': // Calendar year.
			if count == 2 {
				var n int
				consumed, err = parseJavaFixedInt(rest, &n, 2)
				if err == nil {
					ps.year = twoDigitYear(n)
				}
			} else {
				consumed, err = parseJavaVarInt(rest, &ps.year, 1, max(count, 4))
			}

		case 'u': // Extended year (same as calendar year).
			consumed, err = parseJavaVarInt(rest, &ps.year, 1, max(count, 4))

		case 'Y': // ISO week year.
			if count == 2 {
				var n int
				consumed, err = parseJavaFixedInt(rest, &n, 2)
				if err == nil {
					ps.isoWeekYear = twoDigitYear(n)
					ps.hasIsoWeek = true
				}
			} else {
				consumed, err = parseJavaVarInt(rest, &ps.isoWeekYear, 1, max(count, 4))
				if err == nil {
					ps.hasIsoWeek = true
				}
			}

		case 'Q', 'q': // Quarter — consume and discard.
			consumed, err = javaParseQuarter(rest, count)

		case 'M', 'L': // Month.
			switch count {
			case 1:
				consumed, err = parseJavaVarInt(rest, &ps.month, 1, 2)
			case 2:
				consumed, err = parseJavaFixedInt(rest, &ps.month, 2)
			case 3:
				consumed, err = parseJavaMonthAbbr(rest, &ps.month)
			default:
				consumed, err = parseJavaMonthFull(rest, &ps.month)
			}

		case 'w': // ISO week number.
			var n int
			if count == 1 {
				consumed, err = parseJavaVarInt(rest, &n, 1, 2)
			} else {
				consumed, err = parseJavaFixedInt(rest, &n, 2)
			}
			if err == nil {
				ps.isoWeek = n
				ps.hasIsoWeek = true
			}

		case 'W': // Week of month — consume and discard.
			var dummy int
			consumed, err = parseJavaVarInt(rest, &dummy, 1, 1)

		case 'd': // Day of month.
			if count == 1 {
				consumed, err = parseJavaVarInt(rest, &ps.day, 1, 2)
			} else {
				consumed, err = parseJavaFixedInt(rest, &ps.day, 2)
			}

		case 'D': // Day of year (1-based).
			var n int
			if count == 1 {
				consumed, err = parseJavaVarInt(rest, &n, 1, 3)
			} else {
				w := count
				if w > 3 {
					w = 3
				}
				consumed, err = parseJavaFixedInt(rest, &n, w)
			}
			if err == nil {
				ps.dayOfYear = n
				ps.hasDayOfYear = true
			}

		case 'F': // Day of week in month — consume and discard.
			var dummy int
			consumed, err = parseJavaVarInt(rest, &dummy, 1, 1)

		case 'E': // Weekday name.
			if count >= 5 {
				// Single letter — ambiguous, consume and discard.
				if len(rest) == 0 || !isLetter(rest[0]) {
					err = fmt.Errorf("parse: expected weekday letter")
				} else {
					consumed = 1
				}
			} else if count == 4 {
				consumed, err = parseJavaWeekdayFull(rest, &ps.weekday)
				if err == nil {
					ps.hasWeekday = true
				}
			} else {
				consumed, err = parseJavaWeekdayAbbr(rest, &ps.weekday)
				if err == nil {
					ps.hasWeekday = true
				}
			}

		case 'e': // Local day of week, numeric (Java: Sun=1…Sat=7).
			var n int
			if count == 1 {
				consumed, err = parseJavaVarInt(rest, &n, 1, 1)
			} else {
				consumed, err = parseJavaFixedInt(rest, &n, 2)
			}
			if err == nil {
				ps.weekday = n - 1 // Java Sun=1 → Go Sun=0
				ps.hasWeekday = true
			}

		case 'c': // Stand-alone day of week.
			switch count {
			case 1:
				var n int
				consumed, err = parseJavaVarInt(rest, &n, 1, 1)
				if err == nil {
					ps.weekday = n - 1
					ps.hasWeekday = true
				}
			case 2:
				var n int
				consumed, err = parseJavaFixedInt(rest, &n, 2)
				if err == nil {
					ps.weekday = n - 1
					ps.hasWeekday = true
				}
			case 3:
				consumed, err = parseJavaWeekdayAbbr(rest, &ps.weekday)
				if err == nil {
					ps.hasWeekday = true
				}
			default:
				consumed, err = parseJavaWeekdayFull(rest, &ps.weekday)
				if err == nil {
					ps.hasWeekday = true
				}
			}

		case 'a': // AM/PM.
			consumed, err = parseJavaAmPm(rest, &ps.ampm)

		case 'H': // Hour 0-23.
			if count == 1 {
				consumed, err = parseJavaVarInt(rest, &ps.hour, 1, 2)
			} else {
				consumed, err = parseJavaFixedInt(rest, &ps.hour, 2)
			}

		case 'k': // Hour 1-24 (24 = midnight).
			var n int
			if count == 1 {
				consumed, err = parseJavaVarInt(rest, &n, 1, 2)
			} else {
				consumed, err = parseJavaFixedInt(rest, &n, 2)
			}
			if err == nil {
				ps.hour = n % 24 // 24 → 0
			}

		case 'K': // Hour 0-11 (needs AM/PM).
			if count == 1 {
				consumed, err = parseJavaVarInt(rest, &ps.hour, 1, 2)
			} else {
				consumed, err = parseJavaFixedInt(rest, &ps.hour, 2)
			}
			if err == nil {
				ps.use12h = true
				ps.hourK = true
			}

		case 'h': // Hour 1-12 (needs AM/PM).
			if count == 1 {
				consumed, err = parseJavaVarInt(rest, &ps.hour, 1, 2)
			} else {
				consumed, err = parseJavaFixedInt(rest, &ps.hour, 2)
			}
			if err == nil {
				ps.use12h = true
			}

		case 'm': // Minute.
			if count == 1 {
				consumed, err = parseJavaVarInt(rest, &ps.min, 1, 2)
			} else {
				consumed, err = parseJavaFixedInt(rest, &ps.min, 2)
			}

		case 's': // Second.
			if count == 1 {
				consumed, err = parseJavaVarInt(rest, &ps.sec, 1, 2)
			} else {
				consumed, err = parseJavaFixedInt(rest, &ps.sec, 2)
			}

		case 'S': // Fractional second (count = significant digits).
			prec := count
			if prec > 9 {
				prec = 9
			}
			consumed, err = parseJavaFracSec(rest, prec, &ps.nano)

		case 'A': // Millisecond of day.
			consumed, err = parseJavaInt64(rest, &ps.msOfDay)
			if err == nil {
				ps.hasMsOfDay = true
			}

		case 'N': // Nano of day.
			consumed, err = parseJavaInt64(rest, &ps.nanoOfDay)
			if err == nil {
				ps.hasNanoOfDay = true
			}

		case 'n': // Nano of second.
			consumed, err = parseJavaVarInt(rest, &ps.nano, 1, 9)

		case 'z', 'v': // Timezone abbreviation name.
			consumed, err = parseJavaTzAbbr(rest, &ps.loc)

		case 'Z': // Numeric timezone.
			if count >= 4 {
				consumed, err = parseJavaTzGMT(rest, &ps.loc)
			} else {
				consumed, err = parseJavaTzNumeric(rest, &ps.loc)
			}

		case 'X': // ISO 8601 timezone (Z for zero offset).
			consumed, err = parseJavaTzISO(rest, true, &ps.loc)

		case 'x': // ISO 8601 timezone (never Z).
			consumed, err = parseJavaTzISO(rest, false, &ps.loc)

		case 'O': // Localized GMT offset.
			consumed, err = parseJavaTzGMT(rest, &ps.loc)

		case 'V': // Timezone ID — consume and discard; offset tokens already set loc.
			consumed = javaConsumeZoneID(rest)

		default:
			// Non-letter: match literally.
			if len(rest) == 0 || rest[0] != ch {
				err = fmt.Errorf("parse: expected %q at position %d", string(ch), vi)
			} else {
				consumed = 1
			}
		}

		if err != nil {
			return time.Time{}, err
		}
		vi += consumed
		fi = j - 1 // outer loop will fi++
	}

	if vi < len(vRunes) {
		return time.Time{}, fmt.Errorf("parse: unexpected trailing text: %q", string(vRunes[vi:]))
	}

	return ps.build()
}

// --- integer helpers ---

func parseJavaFixedInt(rest []rune, dst *int, width int) (int, error) {
	if len(rest) < width {
		return 0, fmt.Errorf("parse: expected %d-digit number", width)
	}
	n := 0
	for i := range width {
		if !isDigit(rest[i]) {
			return 0, fmt.Errorf("parse: expected digit, got %q", string(rest[i]))
		}
		n = n*10 + int(rest[i]-'0')
	}
	*dst = n
	return width, nil
}

func parseJavaVarInt(rest []rune, dst *int, minWidth, maxWidth int) (int, error) {
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

func parseJavaInt64(rest []rune, dst *int64) (int, error) {
	i := 0
	for i < len(rest) && isDigit(rest[i]) {
		i++
	}
	if i == 0 {
		return 0, fmt.Errorf("parse: expected digits")
	}
	n, err := strconv.ParseInt(string(rest[:i]), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse: invalid integer: %w", err)
	}
	*dst = n
	return i, nil
}

func parseJavaFracSec(rest []rune, prec int, dst *int) (int, error) {
	if len(rest) < prec {
		return 0, fmt.Errorf("parse: expected %d fraction digits", prec)
	}
	n := 0
	for i := range prec {
		if !isDigit(rest[i]) {
			return 0, fmt.Errorf("parse: expected digit, got %q", string(rest[i]))
		}
		n = n*10 + int(rest[i]-'0')
	}
	scale := 1
	for range 9 - prec {
		scale *= 10
	}
	*dst = n * scale
	return prec, nil
}

// --- string token helpers ---

var (
	javaWeekdayAbbrs = [7]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	javaWeekdayFulls = [7]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	javaMonthAbbrs   = [13]string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	javaMonthFulls   = [13]string{"", "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
)

func javaParseEra(rest []rune, _ int) (int, error) {
	for _, c := range []string{"Anno Domini", "Before Christ", "AD", "BC"} {
		cr := []rune(c)
		if len(rest) >= len(cr) && string(rest[:len(cr)]) == c {
			return len(cr), nil
		}
	}
	return 0, fmt.Errorf("parse: expected era designator")
}

func javaParseQuarter(rest []rune, count int) (int, error) {
	switch count {
	case 1:
		if len(rest) == 0 || rest[0] < '1' || rest[0] > '4' {
			return 0, fmt.Errorf("parse: expected quarter 1-4")
		}
		return 1, nil
	case 2:
		if len(rest) < 2 || rest[0] != '0' || rest[1] < '1' || rest[1] > '4' {
			return 0, fmt.Errorf("parse: expected 2-digit quarter 01-04")
		}
		return 2, nil
	case 3:
		if len(rest) < 2 || rest[0] != 'Q' || rest[1] < '1' || rest[1] > '4' {
			return 0, fmt.Errorf("parse: expected quarter like \"Q1\"")
		}
		return 2, nil
	default:
		for _, c := range []string{"1st quarter", "2nd quarter", "3rd quarter", "4th quarter"} {
			cr := []rune(c)
			if len(rest) >= len(cr) && string(rest[:len(cr)]) == c {
				return len(cr), nil
			}
		}
		return 0, fmt.Errorf("parse: expected quarter name like \"1st quarter\"")
	}
}

func parseJavaWeekdayAbbr(rest []rune, dst *int) (int, error) {
	if len(rest) < 3 {
		return 0, fmt.Errorf("parse: expected weekday abbreviation")
	}
	s := string(rest[:3])
	for i, name := range javaWeekdayAbbrs {
		if s == name {
			*dst = i
			return 3, nil
		}
	}
	return 0, fmt.Errorf("parse: unknown weekday abbreviation %q", s)
}

func parseJavaWeekdayFull(rest []rune, dst *int) (int, error) {
	for i, name := range javaWeekdayFulls {
		r := []rune(name)
		if len(rest) >= len(r) && string(rest[:len(r)]) == name {
			*dst = i
			return len(r), nil
		}
	}
	return 0, fmt.Errorf("parse: expected full weekday name")
}

func parseJavaMonthAbbr(rest []rune, dst *int) (int, error) {
	if len(rest) < 3 {
		return 0, fmt.Errorf("parse: expected month abbreviation")
	}
	s := string(rest[:3])
	for i := 1; i <= 12; i++ {
		if s == javaMonthAbbrs[i] {
			*dst = i
			return 3, nil
		}
	}
	return 0, fmt.Errorf("parse: unknown month abbreviation %q", s)
}

func parseJavaMonthFull(rest []rune, dst *int) (int, error) {
	for i := 1; i <= 12; i++ {
		name := javaMonthFulls[i]
		r := []rune(name)
		if len(rest) >= len(r) && string(rest[:len(r)]) == name {
			*dst = i
			return len(r), nil
		}
	}
	return 0, fmt.Errorf("parse: expected full month name")
}

func parseJavaAmPm(rest []rune, dst *int) (int, error) {
	if len(rest) < 2 {
		return 0, fmt.Errorf("parse: expected AM/PM")
	}
	switch string(rest[:2]) {
	case "AM":
		*dst = 1
		return 2, nil
	case "PM":
		*dst = 2
		return 2, nil
	}
	return 0, fmt.Errorf("parse: expected AM/PM, got %q", string(rest[:2]))
}

// --- timezone helpers ---

func parseJavaTzNumeric(rest []rune, dst **time.Location) (int, error) {
	// ±HHmm
	if len(rest) < 5 {
		return 0, fmt.Errorf("parse: expected ±HHmm timezone offset")
	}
	sign, err := javaParseSign(rest[0])
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

func parseJavaTzGMT(rest []rune, dst **time.Location) (int, error) {
	// "GMT" optionally followed by ±h[:mm] or ±HH:mm
	if len(rest) < 3 || string(rest[:3]) != "GMT" {
		return 0, fmt.Errorf("parse: expected GMT timezone")
	}
	if len(rest) == 3 || (rest[3] != '+' && rest[3] != '-') {
		*dst = time.UTC
		return 3, nil
	}
	sign := 1
	if rest[3] == '-' {
		sign = -1
	}
	i := 4
	if i >= len(rest) || !isDigit(rest[i]) {
		return 0, fmt.Errorf("parse: expected hour after GMT sign")
	}
	h := int(rest[i] - '0')
	i++
	if i < len(rest) && isDigit(rest[i]) {
		h = h*10 + int(rest[i]-'0')
		i++
	}
	m := 0
	if i < len(rest) && rest[i] == ':' {
		i++
		if i+1 >= len(rest) || !isDigit(rest[i]) || !isDigit(rest[i+1]) {
			return 0, fmt.Errorf("parse: expected mm after GMT±h:")
		}
		m = int(rest[i]-'0')*10 + int(rest[i+1]-'0')
		i += 2
	}
	*dst = time.FixedZone("", sign*(h*3600+m*60))
	return i, nil
}

func parseJavaTzISO(rest []rune, allowZ bool, dst **time.Location) (int, error) {
	if len(rest) == 0 {
		return 0, fmt.Errorf("parse: expected timezone")
	}
	if allowZ && rest[0] == 'Z' {
		*dst = time.UTC
		return 1, nil
	}
	if len(rest) < 3 {
		return 0, fmt.Errorf("parse: expected timezone offset")
	}
	sign, err := javaParseSign(rest[0])
	if err != nil {
		return 0, err
	}
	if !isDigit(rest[1]) || !isDigit(rest[2]) {
		return 0, fmt.Errorf("parse: expected HH in timezone offset")
	}
	h := int(rest[1]-'0')*10 + int(rest[2]-'0')
	i := 3
	m := 0
	if i < len(rest) && rest[i] == ':' {
		if i+2 >= len(rest) || !isDigit(rest[i+1]) || !isDigit(rest[i+2]) {
			return 0, fmt.Errorf("parse: expected mm after ±HH:")
		}
		m = int(rest[i+1]-'0')*10 + int(rest[i+2]-'0')
		i += 3
	} else if i+1 < len(rest) && isDigit(rest[i]) && isDigit(rest[i+1]) {
		m = int(rest[i]-'0')*10 + int(rest[i+1]-'0')
		i += 2
	}
	*dst = time.FixedZone("", sign*(h*3600+m*60))
	return i, nil
}

func parseJavaTzAbbr(rest []rune, dst **time.Location) (int, error) {
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

func javaConsumeZoneID(rest []rune) int {
	i := 0
	for i < len(rest) && (isLetter(rest[i]) || isDigit(rest[i]) || rest[i] == '/' || rest[i] == '_' || rest[i] == '-' || rest[i] == '+') {
		i++
	}
	return i
}

// --- utilities ---

func javaParseSign(r rune) (int, error) {
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
