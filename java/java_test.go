package java

import (
	"testing"
	"time"
)

func TestFormatJava(t *testing.T) {
	// Wednesday 2023-03-15 14:30:45.123456789 UTC
	ref := JavaTime(time.Date(2023, time.March, 15, 14, 30, 45, 123456789, time.UTC))

	ist := time.FixedZone("IST", 5*3600+30*60) // +05:30
	est := time.FixedZone("EST", -5*3600)      // -05:00
	istRef := JavaTime(time.Date(2023, time.March, 15, 14, 30, 45, 0, ist))
	estRef := JavaTime(time.Date(2023, time.March, 15, 14, 30, 45, 0, est))

	day := func(month time.Month, d int) JavaTime {
		return JavaTime(time.Date(2023, month, d, 0, 0, 0, 0, time.UTC))
	}
	hour := func(h int) JavaTime {
		return JavaTime(time.Date(2023, time.March, 15, h, 0, 0, 0, time.UTC))
	}

	tests := []struct {
		name   string
		t      JavaTime
		format string
		want   string
	}{
		// --- Era ---
		{"G AD", ref, "G", "AD"},
		{"GG AD", ref, "GG", "AD"},
		{"GGG AD", ref, "GGG", "AD"},
		{"GGGG AD", ref, "GGGG", "Anno Domini"},
		{"G BC", JavaTime(time.Date(-100, time.January, 1, 0, 0, 0, 0, time.UTC)), "G", "BC"},
		{"GGGG BC", JavaTime(time.Date(-100, time.January, 1, 0, 0, 0, 0, time.UTC)), "GGGG", "Before Christ"},

		// --- Year ---
		{"y single digit", day(time.January, 5), "y", "2023"},
		{"yy 2-digit", ref, "yy", "23"},
		{"yyyy 4-digit", ref, "yyyy", "2023"},
		{"yyyyy 5-digit", ref, "yyyyy", "02023"},
		{"Y week year", ref, "YYYY", "2023"},
		{"u extended year", ref, "u", "2023"},

		// --- Quarter ---
		{"Q single digit", ref, "Q", "1"},
		{"QQ 2-digit", ref, "QQ", "01"},
		{"QQQ abbreviated", ref, "QQQ", "Q1"},
		{"QQQQ full", ref, "QQQQ", "1st quarter"},
		{"Q2", JavaTime(time.Date(2023, time.June, 15, 0, 0, 0, 0, time.UTC)), "QQQQ", "2nd quarter"},
		{"Q3", JavaTime(time.Date(2023, time.September, 15, 0, 0, 0, 0, time.UTC)), "QQQQ", "3rd quarter"},
		{"Q4", JavaTime(time.Date(2023, time.December, 15, 0, 0, 0, 0, time.UTC)), "QQQQ", "4th quarter"},

		// --- Month ---
		{"M single digit", day(time.March, 5), "M", "3"},
		{"M no leading zero", day(time.January, 5), "M", "1"},
		{"MM zero-padded", ref, "MM", "03"},
		{"MMM abbreviated", ref, "MMM", "Mar"},
		{"MMMM full", ref, "MMMM", "March"},
		{"MMMMM first letter", ref, "MMMMM", "M"},
		{"L stand-alone month", ref, "LLL", "Mar"},
		{"LLLL stand-alone full", ref, "LLLL", "March"},

		// --- Week ---
		{"w ISO week", ref, "w", "11"},
		{"ww padded", ref, "ww", "11"},
		{"W week of month day 15", ref, "W", "3"},
		{"W week of month day 1", day(time.March, 1), "W", "1"},
		{"W week of month day 8", day(time.March, 8), "W", "2"},

		// --- Day ---
		{"d single", day(time.March, 5), "d", "5"},
		{"dd zero-padded", ref, "dd", "15"},
		{"D day of year", ref, "D", "74"},
		{"DD 2-digit", ref, "DD", "74"},
		{"DDD 3-digit", ref, "DDD", "074"},
		{"D Jan 1", day(time.January, 1), "D", "1"},
		{"D Dec 31", day(time.December, 31), "D", "365"},
		{"F day of week in month", ref, "F", "1"},

		// --- Day of week ---
		{"E abbreviated", ref, "E", "Wed"},
		{"EE abbreviated", ref, "EE", "Wed"},
		{"EEE abbreviated", ref, "EEE", "Wed"},
		{"EEEE full", ref, "EEEE", "Wednesday"},
		{"EEEEE first letter", ref, "EEEEE", "W"},
		{"e local day (Wed=4)", ref, "e", "4"},
		{"ee padded", ref, "ee", "04"},
		{"c stand-alone", ref, "c", "4"},
		{"cc stand-alone", ref, "cc", "04"},
		{"ccc abbreviated", ref, "ccc", "Wed"},
		{"cccc full", ref, "cccc", "Wednesday"},

		// --- AM/PM ---
		{"a PM", ref, "a", "PM"},
		{"a AM", hour(9), "a", "AM"},

		// --- Hour (0-23) ---
		{"H single", ref, "H", "14"},
		{"HH zero-padded", ref, "HH", "14"},
		{"H midnight", hour(0), "H", "0"},
		{"HH midnight", hour(0), "HH", "00"},

		// --- Hour (1-24) ---
		{"k single", ref, "k", "14"},
		{"kk zero-padded", ref, "kk", "14"},
		{"k midnight is 24", hour(0), "k", "24"},
		{"kk midnight is 24", hour(0), "kk", "24"},

		// --- Hour (0-11) ---
		{"K single", ref, "K", "2"},
		{"KK zero-padded", ref, "KK", "02"},
		{"K midnight", hour(0), "K", "0"},
		{"K noon", hour(12), "K", "0"},

		// --- Hour (1-12) ---
		{"h single", ref, "h", "2"},
		{"hh zero-padded", ref, "hh", "02"},
		{"h midnight is 12", hour(0), "h", "12"},
		{"h noon is 12", hour(12), "h", "12"},
		{"h 1pm", hour(13), "h", "1"},
		{"hh 1pm", hour(13), "hh", "01"},

		// --- Minute ---
		{"m single", ref, "m", "30"},
		{"mm zero-padded", ref, "mm", "30"},

		// --- Second ---
		{"s single", ref, "s", "45"},
		{"ss zero-padded", ref, "ss", "45"},

		// --- Fractional second ---
		{"S single digit", ref, "S", "1"},
		{"SS 2 digits", ref, "SS", "12"},
		{"SSS milliseconds", ref, "SSS", "123"},
		{"SSSS 4 digits", ref, "SSSS", "1234"},
		{"SSSSSSSSS nanoseconds", ref, "SSSSSSSSS", "123456789"},

		// --- Millisecond of day ---
		{"A milli of day", ref, "A", "52245123"},

		// --- Nano of day ---
		{"N nano of day", ref, "N", "52245123456789"},

		// --- Nano of second ---
		{"n nano of second", ref, "n", "123456789"},

		// --- Timezone ---
		{"z short", ref, "z", "UTC"},
		{"zzz short", ref, "zzz", "UTC"},
		{"zzzz long", ref, "zzzz", "UTC"},
		{"Z RFC 822", ref, "Z", "+0000"},
		{"ZZZ RFC 822", ref, "ZZZ", "+0000"},
		{"ZZZZ GMT", ref, "ZZZZ", "GMT"},
		{"Z positive offset", istRef, "Z", "+0530"},
		{"Z negative offset", estRef, "Z", "-0500"},
		{"O GMT offset", ref, "O", "GMT"},
		{"OOOO GMT offset long", ref, "OOOO", "GMT"},
		{"O IST short hour", istRef, "O", "GMT+5:30"},
		{"OOOO IST padded", istRef, "OOOO", "GMT+05:30"},
		{"X UTC is Z", ref, "X", "Z"},
		{"XX UTC is Z", ref, "XX", "Z"},
		{"XXX UTC is Z", ref, "XXX", "Z"},
		{"X positive offset", istRef, "X", "+0530"},
		{"XX positive offset", istRef, "XX", "+0530"},
		{"XXX positive offset", istRef, "XXX", "+05:30"},
		{"x lowercase (never Z)", ref, "x", "+00"},
		{"xx lowercase", ref, "xx", "+0000"},
		{"xxx lowercase", ref, "xxx", "+00:00"},
		{"v generic tz", ref, "v", "UTC"},
		{"VV tz ID", ref, "VV", "UTC"},

		// --- Quoted text ---
		{"quoted literal", ref, "'year' yyyy", "year 2023"},
		{"escaped quote", ref, "''", "'"},
		{"quote around text", ref, "'quoted'", "quoted"},
		{"quote with escaped inside", ref, "HH'o''clock'", "14o'clock"},

		// --- Mixed formats ---
		{"ISO date", ref, "yyyy-MM-dd", "2023-03-15"},
		{"full datetime", ref, "yyyy-MM-dd HH:mm:ss", "2023-03-15 14:30:45"},
		{"with fractional", ref, "yyyy-MM-dd HH:mm:ss.SSS", "2023-03-15 14:30:45.123"},
		{"with timezone", ref, "yyyy-MM-dd HH:mm:ss Z", "2023-03-15 14:30:45 +0000"},
		{"with IST timezone", istRef, "yyyy-MM-dd HH:mm:ss Z", "2023-03-15 14:30:45 +0530"},
		{"with ISO tz", ref, "yyyy-MM-dd'T'HH:mm:ssXXX", "2023-03-15T14:30:45Z"},

		// --- Unicode in format string ---
		{"unicode in quoted text", ref, "'日期' yyyy-MM-dd", "日期 2023-03-15"},

		// --- Unknown pattern letters ---
		{"unknown letter passes through", ref, "J", "J"},
		{"unknown repeated", ref, "JJJ", "JJJ"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.t.Format(tc.format)
			if got != tc.want {
				t.Errorf("Format(%q) = %q, want %q", tc.format, got, tc.want)
			}
		})
	}
}

func TestFormatJava_DST(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("timezone data not available")
	}

	summer := JavaTime(time.Date(2023, time.July, 15, 12, 0, 0, 0, loc))
	winter := JavaTime(time.Date(2023, time.January, 15, 12, 0, 0, 0, loc))

	if got := summer.Format("z"); got != "EDT" {
		t.Errorf("DST summer z = %q, want %q", got, "EDT")
	}
	if got := winter.Format("z"); got != "EST" {
		t.Errorf("DST winter z = %q, want %q", got, "EST")
	}
	if got := summer.Format("Z"); got != "-0400" {
		t.Errorf("DST summer Z = %q, want %q", got, "-0400")
	}
	if got := winter.Format("Z"); got != "-0500" {
		t.Errorf("DST winter Z = %q, want %q", got, "-0500")
	}
}
