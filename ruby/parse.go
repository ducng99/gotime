package ruby

import (
	"fmt"
	"strings"
	"time"
)

// Parse parses value using Ruby's Time#strftime format string, the inverse of Format.
// Directives that produce non-reversible output (%C, %j, %s, %::z, %u, %w, %U, %W, %V, %G, %g, %v)
// return an error.
func Parse(format, value string) (Time, error) {
	layout, err := rubyFormatToGoLayout(format)
	if err != nil {
		return Time{}, err
	}
	t, err := time.Parse(layout, value)
	if err != nil {
		return Time{}, err
	}
	return Time(t), nil
}

func rubyFormatToGoLayout(format string) (string, error) {
	var sb strings.Builder
	runes := []rune(format)
	i := 0
	for i < len(runes) {
		ch := runes[i]
		if ch != '%' {
			sb.WriteRune(ch)
			i++
			continue
		}
		i++ // consume '%'
		if i >= len(runes) {
			sb.WriteByte('%')
			break
		}

		// Parse flags (same set as Format, case modifiers ignored for parsing).
		var nopad, spacePad, zeroPad bool
		colonCount := 0
		flagsDone := false
		for i < len(runes) && !flagsDone {
			switch runes[i] {
			case '-':
				nopad = true
				i++
			case '_':
				spacePad = true
				i++
			case '0':
				zeroPad = true
				i++
			case '^', '#':
				i++ // case modifiers don't affect parsing
			case ':':
				colonCount++
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
			hasWidth = true
			i++
		}

		if i >= len(runes) {
			sb.WriteByte('%')
			break
		}

		directive := runes[i]
		i++

		if directive == '%' {
			sb.WriteByte('%')
			continue
		}

		token, err := rubyDirectiveToGoLayout(directive, colonCount, hasWidth, width, nopad, spacePad, zeroPad)
		if err != nil {
			return "", err
		}
		sb.WriteString(token)
	}
	return sb.String(), nil
}

func rubyDirectiveToGoLayout(dir rune, colonCount int, hasWidth bool, width int, nopad, spacePad, zeroPad bool) (string, error) {
	switch dir {
	// Year
	case 'Y':
		return "2006", nil
	case 'y':
		return "06", nil
	case 'C':
		return "", fmt.Errorf("parse: unsupported directive %%C")

	// Month
	case 'm':
		if nopad {
			return "1", nil
		}
		return "01", nil
	case 'B':
		return "January", nil
	case 'b', 'h':
		return "Jan", nil

	// Day
	case 'd':
		if nopad {
			return "2", nil
		}
		return "02", nil
	case 'e':
		if nopad || zeroPad {
			return "2", nil
		}
		return "_2", nil // default space-padded
	case 'j':
		return "", fmt.Errorf("parse: unsupported directive %%j")

	// Hour
	case 'H', 'k':
		return "15", nil
	case 'I':
		if nopad {
			return "3", nil
		}
		return "03", nil
	case 'l':
		return "3", nil // default space-padded 12h; '3' is flexible

	// AM/PM
	case 'P':
		return "pm", nil
	case 'p':
		return "PM", nil

	// Minute, second
	case 'M':
		return "04", nil
	case 'S':
		return "05", nil

	// Subsecond
	case 'L':
		return "000", nil // 3 ms digits; relies on preceding '.' in format → ".000"
	case 'N':
		prec := 9
		if hasWidth && width > 0 {
			prec = width
		}
		if prec > 9 {
			prec = 9
		}
		return strings.Repeat("0", prec), nil

	// Unix timestamp
	case 's':
		return "", fmt.Errorf("parse: unsupported directive %%s")

	// Timezone
	case 'z':
		switch colonCount {
		case 0:
			return "-0700", nil
		case 1:
			return "-07:00", nil
		default:
			return "", fmt.Errorf("parse: unsupported directive %%::z")
		}
	case 'Z':
		return "MST", nil

	// Weekday
	case 'A':
		return "Monday", nil
	case 'a':
		return "Mon", nil
	case 'u', 'w':
		return "", fmt.Errorf("parse: unsupported directive %%%c", dir)

	// Week numbers / ISO week year
	case 'U', 'W', 'V', 'G', 'g':
		return "", fmt.Errorf("parse: unsupported directive %%%c", dir)

	// Literals
	case 'n':
		return "\n", nil
	case 't':
		return "\t", nil

	// Composite directives (expanded to equivalent Go layout).
	case 'c':
		return "Mon Jan _2 15:04:05 2006", nil
	case 'D', 'x':
		return "01/02/06", nil
	case 'F':
		return "2006-01-02", nil
	case 'v':
		return "", fmt.Errorf("parse: unsupported directive %%v")
	case 'X', 'T':
		return "15:04:05", nil
	case 'r':
		return "03:04:05 PM", nil
	case 'R':
		return "15:04", nil

	default:
		return "", fmt.Errorf("parse: unsupported directive %%%c", dir)
	}
}
