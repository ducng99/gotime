package java

import (
	"fmt"
	"strings"
	"time"
)

// Parse parses value using Java's SimpleDateFormat/DateTimeFormatter pattern, the inverse of Format.
// Patterns that produce non-reversible output (G, Y, Q, q, w, W, D, F, e, c, k, K, A, N, n, z, O, V, v)
// return an error.
func Parse(format, value string) (Time, error) {
	layout, err := javaFormatToGoLayout(format)
	if err != nil {
		return Time{}, err
	}
	t, err := time.Parse(layout, value)
	if err != nil {
		return Time{}, err
	}
	return Time(t), nil
}

func javaFormatToGoLayout(format string) (string, error) {
	var sb strings.Builder
	runes := []rune(format)
	i := 0
	for i < len(runes) {
		ch := runes[i]

		// Handle quoted text.
		if ch == '\'' {
			if i+1 < len(runes) && runes[i+1] == '\'' {
				sb.WriteByte('\'')
				i += 2
				continue
			}
			i++
			for i < len(runes) {
				if runes[i] == '\'' {
					if i+1 < len(runes) && runes[i+1] == '\'' {
						sb.WriteByte('\'')
						i += 2
					} else {
						i++
						break
					}
				} else {
					sb.WriteRune(runes[i])
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

		token, err := javaLetterToGoLayout(ch, count)
		if err != nil {
			return "", err
		}
		if token == "" {
			// Non-letter or unrecognised: emit literally.
			for k := i; k < j; k++ {
				sb.WriteRune(runes[k])
			}
		} else {
			sb.WriteString(token)
		}
		i = j
	}
	return sb.String(), nil
}

func javaLetterToGoLayout(ch rune, count int) (string, error) {
	switch ch {
	// Era
	case 'G':
		return "", fmt.Errorf("parse: unsupported Java pattern 'G'")

	// Year
	case 'y', 'u':
		if count == 2 {
			return "06", nil
		}
		return "2006", nil
	case 'Y':
		return "", fmt.Errorf("parse: unsupported Java pattern 'Y' (week year)")

	// Quarter
	case 'Q', 'q':
		return "", fmt.Errorf("parse: unsupported Java pattern %q", string(ch))

	// Month
	case 'M', 'L':
		switch count {
		case 1:
			return "1", nil
		case 2:
			return "01", nil
		case 3:
			return "Jan", nil
		default:
			return "January", nil
		}

	// Week
	case 'w', 'W':
		return "", fmt.Errorf("parse: unsupported Java pattern %q", string(ch))

	// Day of month
	case 'd':
		if count == 1 {
			return "2", nil
		}
		return "02", nil

	// Day of year / day of week in month
	case 'D', 'F':
		return "", fmt.Errorf("parse: unsupported Java pattern %q", string(ch))

	// Weekday
	case 'E':
		if count >= 4 {
			return "Monday", nil
		}
		return "Mon", nil
	case 'e', 'c':
		return "", fmt.Errorf("parse: unsupported Java pattern %q", string(ch))

	// AM/PM
	case 'a':
		return "PM", nil

	// Hour
	case 'H':
		return "15", nil // 0-23; Go '15' is flexible
	case 'h':
		if count == 1 {
			return "3", nil
		}
		return "03", nil
	case 'k', 'K':
		return "", fmt.Errorf("parse: unsupported Java pattern %q", string(ch))

	// Minute
	case 'm':
		return "04", nil

	// Second
	case 's':
		return "05", nil

	// Fractional second (each S = one digit, from most significant)
	case 'S':
		prec := count
		if prec > 9 {
			prec = 9
		}
		return strings.Repeat("0", prec), nil

	// Millisecond/nanosecond of day, nanosecond of second
	case 'A', 'N', 'n':
		return "", fmt.Errorf("parse: unsupported Java pattern %q", string(ch))

	// Timezone
	case 'z':
		return "MST", nil
	case 'Z':
		if count >= 4 {
			return "", fmt.Errorf("parse: unsupported Java pattern 'ZZZZ' (GMT+hh:mm)")
		}
		return "-0700", nil
	case 'O':
		return "", fmt.Errorf("parse: unsupported Java pattern 'O'")
	case 'X':
		if count == 3 || count == 5 {
			return "Z07:00", nil
		}
		return "Z0700", nil
	case 'x':
		if count == 3 || count == 5 {
			return "-07:00", nil
		}
		return "-0700", nil
	case 'V', 'v':
		return "", fmt.Errorf("parse: unsupported Java pattern %q", string(ch))

	default:
		return "", nil // non-letter / unrecognised: pass through literally
	}
}
