package ruby

import (
	"testing"
	"time"
)

func TestFormatRuby(t *testing.T) {
	// Wednesday 2023-03-15 14:30:45.123456789 UTC
	ref := Time(time.Date(2023, time.March, 15, 14, 30, 45, 123456789, time.UTC))

	ist := time.FixedZone("IST", 5*3600+30*60)  // +05:30
	est := time.FixedZone("EST", -5*3600)        // -05:00
	istRef := Time(time.Date(2023, time.March, 15, 14, 30, 45, 0, ist))
	estRef := Time(time.Date(2023, time.March, 15, 14, 30, 45, 0, est))

	day := func(month time.Month, d int) Time {
		return Time(time.Date(2023, month, d, 0, 0, 0, 0, time.UTC))
	}
	hour := func(h int) Time {
		return Time(time.Date(2023, time.March, 15, h, 0, 0, 0, time.UTC))
	}
	// Years where Jan 1 falls on each weekday for week-number edge-case testing.
	// 2023: Jan 1 = Sunday, 2024: Mon, 2019: Tue, 2020: Wed, 2015: Thu, 2021: Fri, 2022: Sat
	jan1 := func(year int) Time {
		return Time(time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC))
	}

	tests := []struct {
		name   string
		t      Time
		format string
		want   string
	}{
		// --- Year ---
		{"%Y 4-digit", ref, "%Y", "2023"},
		{"%y 2-digit", ref, "%y", "23"},
		{"%C century", ref, "%C", "20"},

		// --- Month ---
		{"%m zero-padded", ref, "%m", "03"},
		{"%B full name", ref, "%B", "March"},
		{"%b abbreviated", ref, "%b", "Mar"},
		{"%h abbreviated (alias)", ref, "%h", "Mar"},

		// --- Day ---
		{"%d zero-padded", ref, "%d", "15"},
		{"%d single digit", day(time.March, 5), "%d", "05"},
		{"%e space-padded", ref, "%e", "15"},
		{"%e single digit space-padded", day(time.March, 5), "%e", " 5"},
		{"%j day of year Jan1", day(time.January, 1), "%j", "001"},
		{"%j day of year Mar15", ref, "%j", "074"},
		{"%j day of year Dec31", day(time.December, 31), "%j", "365"},
		{"%j leap year Feb29", Time(time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)), "%j", "060"},

		// --- Hour ---
		{"%H 24h zero-padded", ref, "%H", "14"},
		{"%H midnight", hour(0), "%H", "00"},
		{"%k 24h space-padded", ref, "%k", "14"},
		{"%k midnight space-padded", hour(0), "%k", " 0"},
		{"%I 12h midnight", hour(0), "%I", "12"},
		{"%I 12h noon", hour(12), "%I", "12"},
		{"%I 12h 1pm", hour(13), "%I", "01"},
		{"%I 12h 11pm", hour(23), "%I", "11"},
		{"%l 12h space-padded midnight", hour(0), "%l", "12"},
		{"%l 12h space-padded 1pm", hour(13), "%l", " 1"},

		// --- AM/PM ---
		{"%P lowercase pm", ref, "%P", "pm"},
		{"%P lowercase am", hour(9), "%P", "am"},
		{"%p uppercase PM", ref, "%p", "PM"},
		{"%p uppercase AM", hour(9), "%p", "AM"},

		// --- Minute, second, subsecond ---
		{"%M minute", ref, "%M", "30"},
		{"%S second", ref, "%S", "45"},
		{"%L millisecond", ref, "%L", "123"},
		{"%N nanosecond default", ref, "%N", "123456789"},
		{"%3N millisecond precision", ref, "%3N", "123"},
		{"%6N microsecond precision", ref, "%6N", "123456"},
		{"%12N beyond nanosecond", ref, "%12N", "123456789000"},

		// --- Unix timestamp ---
		{"%s unix", ref, "%s", "1678890645"},

		// --- Timezone ---
		{"%z UTC", ref, "%z", "+0000"},
		{"%:z UTC with colon", ref, "%:z", "+00:00"},
		{"%::z UTC with two colons", ref, "%::z", "+00:00:00"},
		{"%z positive offset", istRef, "%z", "+0530"},
		{"%:z positive offset with colon", istRef, "%:z", "+05:30"},
		{"%::z positive offset two colons", istRef, "%::z", "+05:30:00"},
		{"%z negative offset", estRef, "%z", "-0500"},
		{"%:z negative offset with colon", estRef, "%:z", "-05:00"},
		{"%Z UTC name", ref, "%Z", "UTC"},
		{"%Z IST name", istRef, "%Z", "IST"},

		// --- Weekday ---
		{"%A full Wed", ref, "%A", "Wednesday"},
		{"%a abbreviated Wed", ref, "%a", "Wed"},
		{"%u ISO Mon=1", Time(time.Date(2023, 3, 13, 0, 0, 0, 0, time.UTC)), "%u", "1"},
		{"%u ISO Sun=7", Time(time.Date(2023, 3, 19, 0, 0, 0, 0, time.UTC)), "%u", "7"},
		{"%w Sun=0", Time(time.Date(2023, 3, 19, 0, 0, 0, 0, time.UTC)), "%w", "0"},
		{"%w Wed=3", ref, "%w", "3"},

		// --- Week numbers (%U Sunday-based) ---
		// 2023: Jan 1 = Sunday → Jan 1 is week 01
		{"%U Jan1=Sun week1", jan1(2023), "%U", "01"},
		// 2024: Jan 1 = Monday → Jan 1 is week 00
		{"%U Jan1=Mon week0", jan1(2024), "%U", "00"},
		// 2019: Jan 1 = Tuesday → Jan 1 is week 00
		{"%U Jan1=Tue week0", jan1(2019), "%U", "00"},
		// 2022: Jan 1 = Saturday → Jan 1 is week 00
		{"%U Jan1=Sat week0", jan1(2022), "%U", "00"},
		// ref = 2023-03-15 → week 11
		{"%U ref week11", ref, "%U", "11"},

		// --- Week numbers (%W Monday-based) ---
		// 2024: Jan 1 = Monday → Jan 1 is week 01
		{"%W Jan1=Mon week1", jan1(2024), "%W", "01"},
		// 2023: Jan 1 = Sunday → Jan 1 is week 00
		{"%W Jan1=Sun week0", jan1(2023), "%W", "00"},
		// 2019: Jan 1 = Tuesday → Jan 1 is week 00
		{"%W Jan1=Tue week0", jan1(2019), "%W", "00"},
		// ref = 2023-03-15 → week 11
		{"%W ref week11", ref, "%W", "11"},

		// --- ISO week date ---
		{"%G ISO year same", ref, "%G", "2023"},
		{"%G ISO year differs Jan1", Time(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)), "%G", "2015"},
		{"%V ISO week ref", ref, "%V", "11"},
		{"%V ISO week boundary", Time(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)), "%V", "53"},
		{"%g ISO year 2-digit", Time(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)), "%g", "15"},

		// --- Literals ---
		{"%n newline", ref, "%n", "\n"},
		{"%t tab", ref, "%t", "\t"},
		{"%% literal percent", ref, "%%", "%"},

		// --- Composite directives ---
		{"%c datetime", ref, "%c", "Wed Mar 15 14:30:45 2023"},
		{"%c single-digit day double-space", day(time.March, 5), "%c", "Sun Mar  5 00:00:00 2023"},
		{"%D date", ref, "%D", "03/15/23"},
		{"%x date (alias of %D)", ref, "%x", "03/15/23"},
		{"%F ISO date", ref, "%F", "2023-03-15"},
		{"%v VMS date", ref, "%v", "15-MAR-2023"},
		{"%v VMS single-digit day", day(time.March, 5), "%v", " 5-MAR-2023"},
		{"%T time", ref, "%T", "14:30:45"},
		{"%X time (alias of %T)", ref, "%X", "14:30:45"},
		{"%r 12h time", ref, "%r", "02:30:45 PM"},
		{"%r midnight", hour(0), "%r", "12:00:00 AM"},
		{"%R HH:MM", ref, "%R", "14:30"},

		// --- Padding flags ---
		{"%-d no padding", ref, "%-d", "15"},
		{"%-d no padding single digit", day(time.March, 5), "%-d", "5"},
		{"%_d space-pad", day(time.March, 5), "%_d", " 5"},
		{"%0e zero-pad override", day(time.March, 5), "%0e", "05"},
		{"%5Y width override", ref, "%5Y", "02023"},
		{"%4e space-pad to width 4", day(time.March, 5), "%4e", "   5"},
		{"%12A width on string", ref, "%12A", "   Wednesday"},
		{"%010A zero-pad on string", ref, "%010A", "0Wednesday"},
		{"%8z zero-pad width on z", ref, "%8z", "+0000000"},

		// --- Case flags ---
		{"%^A uppercase weekday", ref, "%^A", "WEDNESDAY"},
		{"%^a uppercase abbr weekday", ref, "%^a", "WED"},
		{"%^B uppercase month", ref, "%^B", "MARCH"},
		{"%^b uppercase abbr month", ref, "%^b", "MAR"},
		{"%^p uppercase AM/PM", ref, "%^p", "PM"},
		{"%#p swap PM→pm", ref, "%#p", "pm"},
		{"%#P swap pm→PM", hour(9), "%#P", "AM"},
		{"%#A swap Wednesday→WEDNESDAY", ref, "%#A", "WEDNESDAY"},
		{"%#B swap March→MARCH", ref, "%#B", "MARCH"},

		// --- Unknown directive passthrough ---
		{"%Q unknown passes through", ref, "%Q", "%Q"},
		{"%^Q unknown with flag", ref, "%^Q", "%^Q"},

		// --- Unicode in format string ---
		{"unicode passthrough", ref, "日本語%Y", "日本語2023"},

		// --- Mixed formats ---
		{"ISO date", ref, "%Y-%m-%d", "2023-03-15"},
		{"full datetime", ref, "%Y-%m-%d %H:%M:%S", "2023-03-15 14:30:45"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.t.Format(tc.format)
			if got != tc.want {
				t.Errorf("FormatRuby(%q) = %q, want %q", tc.format, got, tc.want)
			}
		})
	}
}

func TestFormatRuby_DST(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("timezone data not available")
	}

	summer := Time(time.Date(2023, time.July, 15, 12, 0, 0, 0, loc))
	winter := Time(time.Date(2023, time.January, 15, 12, 0, 0, 0, loc))

	if got := summer.Format("%Z"); got != "EDT" {
		t.Errorf("DST summer %%Z = %q, want %q", got, "EDT")
	}
	if got := winter.Format("%Z"); got != "EST" {
		t.Errorf("DST winter %%Z = %q, want %q", got, "EST")
	}
	if got := summer.Format("%:z"); got != "-04:00" {
		t.Errorf("DST summer %%:z = %q, want %q", got, "-04:00")
	}
	if got := winter.Format("%:z"); got != "-05:00" {
		t.Errorf("DST winter %%:z = %q, want %q", got, "-05:00")
	}
}
