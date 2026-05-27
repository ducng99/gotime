package java

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Time time.Time

// DateTimeFormatter constants - Java predefined format patterns
// See https://docs.oracle.com/en/java/javase/17/docs/api/java.base/java/time/format/DateTimeFormatter.html
// Note: Java's predefined formatters are DateTimeFormatter objects built with DateTimeFormatterBuilder,
// not simple pattern strings. These are the closest pattern equivalents.
const (
	DateTimeBasicIsoDate      = `yyyyMMdd`
	DateTimeIsoLocalDate      = `yyyy-MM-dd`
	DateTimeIsoOffsetDate     = `yyyy-MM-ddXXX`
	DateTimeIsoDate           = `yyyy-MM-ddXXX`
	DateTimeIsoLocalTime      = `HH:mm:ss`
	DateTimeIsoOffsetTime     = `HH:mm:ssXXX`
	DateTimeIsoTime           = `HH:mm:ssXXX`
	DateTimeIsoLocalDateTime  = `yyyy-MM-dd'T'HH:mm:ss`
	DateTimeIsoOffsetDateTime = `yyyy-MM-dd'T'HH:mm:ssXXX`
	DateTimeIsoZonedDateTime  = `yyyy-MM-dd'T'HH:mm:ssXXX'['VV']'`
	DateTimeIsoDateTime       = `yyyy-MM-dd'T'HH:mm:ssXXX'['VV']'`
	DateTimeIsoOrdinalDate    = `yyyy-DDD`
	DateTimeIsoWeekDate       = `YYYY-'W'ww-e`
	DateTimeIsoInstant        = `yyyy-MM-dd'T'HH:mm:ss'Z'`
	DateTimeRfc1123           = `EEE, d MMM yyyy HH:mm:ss z`
)

// Format formats using Java's SimpleDateFormat/DateTimeFormatter pattern.
// See https://docs.oracle.com/javase/8/docs/api/java/text/SimpleDateFormat.html
// Literal text is enclosed in single quotes. A doubled single quote (”)
// produces a literal single quote. Consecutive identical pattern letters
// form a specifier; the count changes the output format.
func (t *Time) Format(format string) string {
	goTime := time.Time(*t)
	var out strings.Builder

	runes := []rune(format)
	i := 0
	for i < len(runes) {
		ch := runes[i]

		// Handle quoted text or escaped quote.
		if ch == '\'' {
			// Escaped quote: '' produces a literal '.
			if i+1 < len(runes) && runes[i+1] == '\'' {
				out.WriteByte('\'')
				i += 2
				continue
			}
			// Start of quoted text.
			i++
			for i < len(runes) {
				if runes[i] == '\'' {
					if i+1 < len(runes) && runes[i+1] == '\'' {
						out.WriteByte('\'')
						i += 2
					} else {
						i++
						break
					}
				} else {
					out.WriteRune(runes[i])
					i++
				}
			}
			continue
		}

		// Find the run of identical letters.
		j := i + 1
		for j < len(runes) && runes[j] == ch {
			j++
		}
		count := j - i

		value := javaValue(goTime, ch, count)
		if value == "" {
			// Unknown pattern letter: emit literally.
			for k := i; k < j; k++ {
				out.WriteRune(runes[k])
			}
		} else {
			out.WriteString(value)
		}

		i = j
	}

	return out.String()
}

func javaValue(t time.Time, ch rune, count int) string {
	switch ch {
	// Era
	case 'G':
		if count >= 4 {
			year := t.Year()
			if year >= 1 {
				return "Anno Domini"
			}
			return "Before Christ"
		}
		year := t.Year()
		if year >= 1 {
			return "AD"
		}
		return "BC"

	// Year
	case 'y':
		year := t.Year()
		if year < 0 {
			year = -year
		}
		switch count {
		case 1:
			return strconv.Itoa(year)
		case 2:
			return fmt.Sprintf("%02d", year%100)
		default:
			return fmt.Sprintf("%0*d", count, year)
		}
	case 'Y':
		// Week year (same as calendar year for most cases).
		isoYear, _ := t.ISOWeek()
		switch count {
		case 1:
			return strconv.Itoa(isoYear)
		case 2:
			return fmt.Sprintf("%02d", isoYear%100)
		default:
			return fmt.Sprintf("%0*d", count, isoYear)
		}
	case 'u':
		// Extended year (same as year for AD).
		return strconv.Itoa(t.Year())

	// Quarter
	case 'Q':
		month := int(t.Month())
		quarter := (month-1)/3 + 1
		switch count {
		case 1:
			return strconv.Itoa(quarter)
		case 2:
			return fmt.Sprintf("%02d", quarter)
		case 3:
			return fmt.Sprintf("Q%d", quarter)
		case 4:
			return []string{"1st quarter", "2nd quarter", "3rd quarter", "4th quarter"}[quarter-1]
		default:
			return fmt.Sprintf("Q%d", quarter)
		}
	case 'q':
		// Stand-alone quarter (same as Q for English).
		month := int(t.Month())
		quarter := (month-1)/3 + 1
		switch count {
		case 1:
			return strconv.Itoa(quarter)
		case 2:
			return fmt.Sprintf("%02d", quarter)
		case 3:
			return fmt.Sprintf("Q%d", quarter)
		case 4:
			return []string{"1st quarter", "2nd quarter", "3rd quarter", "4th quarter"}[quarter-1]
		default:
			return fmt.Sprintf("Q%d", quarter)
		}

	// Month
	case 'M':
		return monthValue(t, count)
	case 'L':
		// Stand-alone month (same as M for English).
		return monthValue(t, count)

	// Week
	case 'w':
		_, week := t.ISOWeek()
		if count == 1 {
			return strconv.Itoa(week)
		}
		return fmt.Sprintf("%02d", week)
	case 'W':
		// Week of month.
		day := t.Day()
		weekOfMonth := (day-1)/7 + 1
		return strconv.Itoa(weekOfMonth)

	// Day
	case 'd':
		switch count {
		case 1:
			return strconv.Itoa(t.Day())
		case 2:
			return fmt.Sprintf("%02d", t.Day())
		default:
			return fmt.Sprintf("%0*d", count, t.Day())
		}
	case 'D':
		// Day of year.
		switch count {
		case 1:
			return strconv.Itoa(t.YearDay())
		case 2:
			return fmt.Sprintf("%02d", t.YearDay())
		case 3:
			return fmt.Sprintf("%03d", t.YearDay())
		default:
			return fmt.Sprintf("%0*d", count, t.YearDay())
		}
	case 'F':
		// Day of week in month (DateTimeFormatter style: cycles 1-7).
		return strconv.Itoa((t.Day()-1)%7 + 1)

	// Day of week
	case 'E':
		switch count {
		case 1, 2, 3:
			return t.Format("Mon")
		case 4:
			return t.Format("Monday")
		case 5:
			return t.Format("Monday")[:1]
		default:
			return t.Format("Monday")
		}
	case 'e':
		// Local day of week (1=Sunday in Java default locale, but we use 1=Monday ISO).
		d := int(t.Weekday())
		// Java default locale: Sunday=1, Monday=2, ..., Saturday=7
		javaDay := d + 1
		if javaDay > 7 {
			javaDay = 1
		}
		if count == 1 {
			return strconv.Itoa(javaDay)
		}
		return fmt.Sprintf("%02d", javaDay)
	case 'c':
		// Stand-alone day of week (Java: Sunday=1, Monday=2, ..., Saturday=7).
		d := int(t.Weekday()) + 1
		if d > 7 {
			d = 1
		}
		switch count {
		case 1:
			return strconv.Itoa(d)
		case 2:
			return fmt.Sprintf("%02d", d)
		case 3:
			return t.Format("Mon")
		default:
			return t.Format("Monday")
		}

	// AM/PM
	case 'a':
		return t.Format("PM")

	// Hour
	case 'H':
		// Hour of day (0-23).
		switch count {
		case 1:
			return strconv.Itoa(t.Hour())
		case 2:
			return fmt.Sprintf("%02d", t.Hour())
		default:
			return fmt.Sprintf("%0*d", count, t.Hour())
		}
	case 'k':
		// Hour of day (1-24).
		h := t.Hour()
		if h == 0 {
			h = 24
		}
		switch count {
		case 1:
			return strconv.Itoa(h)
		case 2:
			return fmt.Sprintf("%02d", h)
		default:
			return fmt.Sprintf("%0*d", count, h)
		}
	case 'K':
		// Hour in am/pm (0-11).
		h := t.Hour() % 12
		switch count {
		case 1:
			return strconv.Itoa(h)
		case 2:
			return fmt.Sprintf("%02d", h)
		default:
			return fmt.Sprintf("%0*d", count, h)
		}
	case 'h':
		// Hour in am/pm (1-12).
		h := t.Hour() % 12
		if h == 0 {
			h = 12
		}
		switch count {
		case 1:
			return strconv.Itoa(h)
		case 2:
			return fmt.Sprintf("%02d", h)
		default:
			return fmt.Sprintf("%0*d", count, h)
		}

	// Minute
	case 'm':
		switch count {
		case 1:
			return strconv.Itoa(t.Minute())
		case 2:
			return fmt.Sprintf("%02d", t.Minute())
		default:
			return fmt.Sprintf("%0*d", count, t.Minute())
		}

	// Second
	case 's':
		switch count {
		case 1:
			return strconv.Itoa(t.Second())
		case 2:
			return fmt.Sprintf("%02d", t.Second())
		default:
			return fmt.Sprintf("%0*d", count, t.Second())
		}

	// Fractional second
	case 'S':
		// Each S is a digit of fractional second.
		nano := t.Nanosecond()
		switch count {
		case 1:
			return fmt.Sprintf("%d", nano/100_000_000)
		case 2:
			return fmt.Sprintf("%02d", nano/10_000_000)
		case 3:
			return fmt.Sprintf("%03d", nano/1_000_000)
		default:
			s := fmt.Sprintf("%09d", nano)
			if count <= 9 {
				return s[:count]
			}
			return s + strings.Repeat("0", count-9)
		}

	// Millisecond of day
	case 'A':
		millisOfDay := (t.Hour()*3600+t.Minute()*60+t.Second())*1000 + t.Nanosecond()/1_000_000
		return strconv.Itoa(millisOfDay)

	// Nano of day
	case 'N':
		nanoOfDay := int64(t.Hour())*3_600_000_000_000 +
			int64(t.Minute())*60_000_000_000 +
			int64(t.Second())*1_000_000_000 +
			int64(t.Nanosecond())
		return strconv.FormatInt(nanoOfDay, 10)

	// Nano of second
	case 'n':
		return strconv.Itoa(t.Nanosecond())

	// Timezone
	case 'z':
		if count >= 4 {
			name, _ := t.Zone()
			// Try to produce a long name.
			return name
		}
		name, _ := t.Zone()
		return name
	case 'Z':
		_, offset := t.Zone()
		if count >= 4 {
			// GMT prefix. For zero offset, just "GMT".
			if offset == 0 {
				return "GMT"
			}
			sign := "+"
			if offset < 0 {
				sign = "-"
				offset = -offset
			}
			h, m := offset/3600, (offset%3600)/60
			return fmt.Sprintf("GMT%s%02d:%02d", sign, h, m)
		}
		// ±HHmm
		sign := "+"
		if offset < 0 {
			sign = "-"
			offset = -offset
		}
		h, m := offset/3600, (offset%3600)/60
		return fmt.Sprintf("%s%02d%02d", sign, h, m)
	case 'O':
		// Localized GMT offset. O uses short format (single-digit hour), OOOO uses long (zero-padded).
		_, offset := t.Zone()
		if offset == 0 {
			return "GMT"
		}
		sign := "+"
		if offset < 0 {
			sign = "-"
			offset = -offset
		}
		h, m := offset/3600, (offset%3600)/60
		if count >= 4 {
			return fmt.Sprintf("GMT%s%02d:%02d", sign, h, m)
		}
		if m == 0 {
			return fmt.Sprintf("GMT%s%d", sign, h)
		}
		return fmt.Sprintf("GMT%s%d:%02d", sign, h, m)
	case 'X':
		// ISO 8601 timezone. For zero offset, all counts output "Z".
		_, offset := t.Zone()
		if offset == 0 {
			return "Z"
		}
		sign := "+"
		if offset < 0 {
			sign = "-"
			offset = -offset
		}
		h, m := offset/3600, (offset%3600)/60
		switch count {
		case 1:
			if m == 0 {
				return fmt.Sprintf("%s%02d", sign, h)
			}
			return fmt.Sprintf("%s%02d%02d", sign, h, m)
		case 2:
			return fmt.Sprintf("%s%02d%02d", sign, h, m)
		case 3:
			return fmt.Sprintf("%s%02d:%02d", sign, h, m)
		case 4:
			return fmt.Sprintf("%s%02d%02d", sign, h, m)
		default:
			return fmt.Sprintf("%s%02d:%02d", sign, h, m)
		}
	case 'x':
		// ISO 8601 timezone (never outputs Z).
		_, offset := t.Zone()
		sign := "+"
		if offset < 0 {
			sign = "-"
			offset = -offset
		}
		h, m := offset/3600, (offset%3600)/60
		switch count {
		case 1:
			if m == 0 {
				return fmt.Sprintf("%s%02d", sign, h)
			}
			return fmt.Sprintf("%s%02d%02d", sign, h, m)
		case 2:
			return fmt.Sprintf("%s%02d%02d", sign, h, m)
		case 3:
			return fmt.Sprintf("%s%02d:%02d", sign, h, m)
		default:
			return fmt.Sprintf("%s%02d:%02d", sign, h, m)
		}

	// Full date/time patterns (commonly used).
	case 'v':
		// Generic timezone (same as z for most cases).
		name, _ := t.Zone()
		return name
	case 'V':
		// Timezone ID. Requires count=2 in Java.
		if count == 2 {
			return t.Location().String()
		}
		return ""

	default:
		return ""
	}
}

func monthValue(t time.Time, count int) string {
	switch count {
	case 1:
		return strconv.Itoa(int(t.Month()))
	case 2:
		return fmt.Sprintf("%02d", int(t.Month()))
	case 3:
		return t.Format("Jan")
	case 4:
		return t.Format("January")
	case 5:
		return t.Format("January")[:1]
	default:
		return t.Format("January")
	}
}
