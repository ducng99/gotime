package php

import (
	"testing"
	"time"
)

func TestParsePhp(t *testing.T) {
	ref := Time(time.Date(2023, time.March, 15, 14, 30, 45, 0, time.UTC))
	ist := time.FixedZone("IST", 5*3600+30*60)
	istRef := Time(time.Date(2023, time.March, 15, 14, 30, 45, 0, ist))
	msRef := Time(time.Date(2023, time.March, 15, 14, 30, 45, 123000000, time.UTC))

	roundTrip := func(t *testing.T, tm Time, format string) {
		t.Helper()
		formatted := tm.Format(format)
		parsed, err := Parse(format, formatted)
		if err != nil {
			t.Fatalf("Parse(%q, %q) error: %v", format, formatted, err)
		}
		got := parsed.Format(format)
		if got != formatted {
			t.Errorf("round-trip: got %q, want %q", got, formatted)
		}
	}

	t.Run("date", func(t *testing.T) { roundTrip(t, ref, "Y-m-d") })
	t.Run("datetime", func(t *testing.T) { roundTrip(t, ref, "Y-m-d H:i:s") })
	t.Run("unpadded day/month", func(t *testing.T) { roundTrip(t, ref, "j n Y") })
	t.Run("full names", func(t *testing.T) { roundTrip(t, ref, "l, d F Y") })
	t.Run("short names", func(t *testing.T) { roundTrip(t, ref, "D, d M Y") })
	t.Run("12h with AM/PM", func(t *testing.T) { roundTrip(t, ref, "h:i:s A") })
	t.Run("12h no-pad am/pm", func(t *testing.T) { roundTrip(t, ref, "g:i:s a") })
	t.Run("2-digit year", func(t *testing.T) { roundTrip(t, ref, "y-m-d") })
	t.Run("offset -0700", func(t *testing.T) { roundTrip(t, istRef, "Y-m-d H:i:s O") })
	t.Run("offset -07:00", func(t *testing.T) { roundTrip(t, istRef, "Y-m-d H:i:s P") })
	t.Run("UTC Z offset", func(t *testing.T) { roundTrip(t, ref, "Y-m-d H:i:s p") })
	t.Run("milliseconds", func(t *testing.T) { roundTrip(t, msRef, `Y-m-d\TH:i:s.vP`) })
	t.Run("ATOM constant", func(t *testing.T) { roundTrip(t, istRef, DateTimeATOM) })
	t.Run("RFC3339 constant", func(t *testing.T) { roundTrip(t, istRef, DateTimeRFC3339) })
	t.Run("RFC1123 constant", func(t *testing.T) { roundTrip(t, ref, DateTimeRFC1123) })
	t.Run("COOKIE constant", func(t *testing.T) { roundTrip(t, ref, DateTimeCOOKIE) })
	t.Run("c full datetime", func(t *testing.T) { roundTrip(t, istRef, "c") })
	t.Run("r RFC2822", func(t *testing.T) { roundTrip(t, ref, "r") })

	// Newly-supported tokens
	t.Run("ordinal S", func(t *testing.T) { roundTrip(t, ref, "jS F Y") })
	t.Run("weekday N", func(t *testing.T) { roundTrip(t, ref, "Y-m-d N") })
	t.Run("weekday w", func(t *testing.T) { roundTrip(t, ref, "Y-m-d w") })
	t.Run("day of year z", func(t *testing.T) { roundTrip(t, ref, "Y z H:i:s") })
	t.Run("ISO week W+o+N", func(t *testing.T) { roundTrip(t, ref, "o-W-N") })
	t.Run("days in month t", func(t *testing.T) { roundTrip(t, ref, "Y-m-t") })
	t.Run("leap year L", func(t *testing.T) { roundTrip(t, ref, "Y-m-d L") })
	t.Run("DST indicator I", func(t *testing.T) { roundTrip(t, ref, "Y-m-d H:i:s I") })
	t.Run("Swatch time B", func(t *testing.T) { roundTrip(t, ref, "Y-m-d.B") })
	t.Run("tz seconds Z", func(t *testing.T) { roundTrip(t, istRef, "Y-m-d H:i:s Z") })
	t.Run("tz identifier e", func(t *testing.T) { roundTrip(t, ref, "Y-m-d H:i:s e") })
	t.Run("unix timestamp U", func(t *testing.T) { roundTrip(t, ref, "U") })

	// Direct parse: known input → expected time components
	t.Run("direct parse UTC", func(t *testing.T) {
		parsed, err := Parse("Y-m-d H:i:s", "2023-03-15 14:30:45")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := time.Time(parsed)
		if got.Year() != 2023 || got.Month() != 3 || got.Day() != 15 ||
			got.Hour() != 14 || got.Minute() != 30 || got.Second() != 45 {
			t.Errorf("unexpected time: %v", got)
		}
	})

	t.Run("expanded year X", func(t *testing.T) { roundTrip(t, ref, "X-m-d") })
	t.Run("expanded year X ISO8601Expanded", func(t *testing.T) { roundTrip(t, istRef, DateTimeISO8601Expanded) })
}
