package php

import (
	"testing"
	"time"
)

func TestFormatPhp(t *testing.T) {
	// Wednesday 2023-03-15 14:30:45.123456789 UTC
	ref := Time(time.Date(2023, time.March, 15, 14, 30, 45, 123456789, time.UTC))

	ist := time.FixedZone("IST", 5*3600+30*60)   // +05:30
	istRef := Time(time.Date(2023, time.March, 15, 14, 30, 45, 0, ist))
	est := time.FixedZone("EST", -5*3600)         // -05:00
	estRef := Time(time.Date(2023, time.March, 15, 14, 30, 45, 0, est))

	day := func(d int) Time {
		return Time(time.Date(2023, time.January, d, 0, 0, 0, 0, time.UTC))
	}
	hour := func(h int) Time {
		return Time(time.Date(2023, time.March, 15, h, 0, 0, 0, time.UTC))
	}

	tests := []struct {
		name   string
		t      Time
		format string
		want   string
	}{
		// --- Day ---
		{"d leading zero", ref, "d", "15"},
		{"D short name", ref, "D", "Wed"},
		{"j no leading zero", ref, "j", "15"},
		{"l full name", ref, "l", "Wednesday"},
		{"N iso weekday Wed=3", ref, "N", "3"},
		{"N iso weekday Sun=7", Time(time.Date(2023, time.March, 19, 0, 0, 0, 0, time.UTC)), "N", "7"},
		{"S th", ref, "S", "th"},
		{"S st day 1", day(1), "S", "st"},
		{"S nd day 2", day(2), "S", "nd"},
		{"S rd day 3", day(3), "S", "rd"},
		{"S th day 4", day(4), "S", "th"},
		{"S th day 11", day(11), "S", "th"},
		{"S th day 12", day(12), "S", "th"},
		{"S th day 13", day(13), "S", "th"},
		{"S st day 21", day(21), "S", "st"},
		{"S nd day 22", day(22), "S", "nd"},
		{"S rd day 23", day(23), "S", "rd"},
		{"S st day 31", day(31), "S", "st"},
		{"w Sun=0", Time(time.Date(2023, time.March, 19, 0, 0, 0, 0, time.UTC)), "w", "0"},
		{"w Wed=3", ref, "w", "3"},
		{"z day of year 0-indexed", ref, "z", "73"},
		{"z Jan 1 is 0", day(1), "z", "0"},

		// --- Week ---
		{"W iso week 11", ref, "W", "11"},
		{"W padded single digit", Time(time.Date(2023, time.January, 4, 0, 0, 0, 0, time.UTC)), "W", "01"},

		// --- Month ---
		{"F full month", ref, "F", "March"},
		{"m with leading zero", ref, "m", "03"},
		{"M short month", ref, "M", "Mar"},
		{"n no leading zero", ref, "n", "3"},
		{"t March has 31", ref, "t", "31"},
		{"t April has 30", Time(time.Date(2023, time.April, 1, 0, 0, 0, 0, time.UTC)), "t", "30"},
		{"t Feb non-leap has 28", Time(time.Date(2023, time.February, 1, 0, 0, 0, 0, time.UTC)), "t", "28"},
		{"t Feb leap has 29", Time(time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC)), "t", "29"},

		// --- Year ---
		{"L not leap 2023", ref, "L", "0"},
		{"L leap 2024", Time(time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)), "L", "1"},
		{"L leap 2000 div400", Time(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)), "L", "1"},
		{"L not leap 1900 div100", Time(time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)), "L", "0"},
		{"o iso year same as calendar", ref, "o", "2023"},
		{"o iso year differs from calendar", Time(time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC)), "o", "2015"},
		{"Y 4-digit year", ref, "Y", "2023"},
		{"y 2-digit year", ref, "y", "23"},

		// --- Time ---
		{"a pm", ref, "a", "pm"},
		{"a am", hour(9), "a", "am"},
		{"A PM", ref, "A", "PM"},
		{"A AM", hour(9), "A", "AM"},
		{"B swatch time", ref, "B", "646"},
		{"B padded", hour(0), "B", "041"},
		{"g 12-hour no pad", ref, "g", "2"},
		{"g midnight is 12", hour(0), "g", "12"},
		{"g noon is 12", hour(12), "g", "12"},
		{"G 24-hour no pad", ref, "G", "14"},
		{"G midnight is 0", hour(0), "G", "0"},
		{"h 12-hour padded", ref, "h", "02"},
		{"h midnight padded", hour(0), "h", "12"},
		{"H 24-hour padded", ref, "H", "14"},
		{"H midnight padded", hour(0), "H", "00"},
		{"i minutes", ref, "i", "30"},
		{"s seconds", ref, "s", "45"},
		{"u microseconds", ref, "u", "123456"},
		{"v milliseconds", ref, "v", "123"},

		// --- Timezone (UTC) ---
		{"e UTC identifier", ref, "e", "UTC"},
		{"I no DST UTC", ref, "I", "0"},
		{"O UTC offset", ref, "O", "+0000"},
		{"P UTC offset with colon", ref, "P", "+00:00"},
		{"p Z for UTC", ref, "p", "Z"},
		{"T UTC abbreviation", ref, "T", "UTC"},
		{"Z UTC seconds offset", ref, "Z", "0"},

		// --- Timezone (fixed positive offset +05:30) ---
		{"e IST identifier", istRef, "e", "IST"},
		{"O positive offset", istRef, "O", "+0530"},
		{"P positive offset with colon", istRef, "P", "+05:30"},
		{"p positive offset not UTC", istRef, "p", "+05:30"},
		{"Z positive offset seconds", istRef, "Z", "19800"},

		// --- Timezone (fixed negative offset -05:00) ---
		{"O negative offset", estRef, "O", "-0500"},
		{"P negative offset with colon", estRef, "P", "-05:00"},
		{"Z negative offset seconds", estRef, "Z", "-18000"},

		// --- Full date/time ---
		{"c ISO 8601", ref, "c", "2023-03-15T14:30:45+00:00"},
		{"r RFC 2822", ref, "r", "Wed, 15 Mar 2023 14:30:45 +0000"},
		{"U unix timestamp", ref, "U", "1678890645"},

		// --- Escaping and literals ---
		{"backslash escapes format char", ref, `\Y`, "Y"},
		{"backslash escapes non-format char", ref, `\-`, "-"},
		{"expanded year X", ref, "X", "+2023"},
		{"mixed format", ref, "Y-m-d", "2023-03-15"},
		{"escaped format chars in mixed", ref, `\Ym\d`, "Y03d"},
		{"full datetime format", ref, "Y-m-d H:i:s", "2023-03-15 14:30:45"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.t.Format(tc.format)
			if got != tc.want {
				t.Errorf("FormatPhp(%q) = %q, want %q", tc.format, got, tc.want)
			}
		})
	}
}

func TestFormatPhp_DST(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("timezone data not available")
	}

	summer := Time(time.Date(2023, time.July, 15, 12, 0, 0, 0, loc))
	winter := Time(time.Date(2023, time.January, 15, 12, 0, 0, 0, loc))

	if got := summer.Format("I"); got != "1" {
		t.Errorf("DST active: FormatPhp(\"I\") = %q, want \"1\"", got)
	}
	if got := winter.Format("I"); got != "0" {
		t.Errorf("DST inactive: FormatPhp(\"I\") = %q, want \"0\"", got)
	}
	if got := summer.Format("e"); got != "America/New_York" {
		t.Errorf("FormatPhp(\"e\") = %q, want \"America/New_York\"", got)
	}
}
