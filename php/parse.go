package php

import (
	"fmt"
	"strings"
	"time"
)

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
