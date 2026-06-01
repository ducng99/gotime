package java

import (
	"testing"
	"time"
)

func TestParseJava(t *testing.T) {
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

	t.Run("basic date", func(t *testing.T) { roundTrip(t, ref, "yyyy-MM-dd") })
	t.Run("datetime", func(t *testing.T) { roundTrip(t, ref, "yyyy-MM-dd HH:mm:ss") })
	t.Run("2-digit year", func(t *testing.T) { roundTrip(t, ref, "yy-MM-dd") })
	t.Run("unpadded day/month", func(t *testing.T) { roundTrip(t, ref, "d M yyyy") })
	t.Run("short month name", func(t *testing.T) { roundTrip(t, ref, "d MMM yyyy") })
	t.Run("full month name", func(t *testing.T) { roundTrip(t, ref, "d MMMM yyyy") })
	t.Run("short weekday", func(t *testing.T) { roundTrip(t, ref, "EEE, dd MMM yyyy HH:mm:ss") })
	t.Run("full weekday", func(t *testing.T) { roundTrip(t, ref, "EEEE, dd MMMM yyyy") })
	t.Run("12h AM/PM", func(t *testing.T) { roundTrip(t, ref, "hh:mm:ss a") })
	t.Run("12h unpadded", func(t *testing.T) { roundTrip(t, ref, "h:mm:ss a") })
	t.Run("ISO offset XXX", func(t *testing.T) { roundTrip(t, istRef, "yyyy-MM-dd'T'HH:mm:ssXXX") })
	t.Run("ISO offset XX", func(t *testing.T) { roundTrip(t, istRef, "yyyy-MM-dd'T'HH:mm:ssXX") })
	t.Run("numeric tz Z", func(t *testing.T) { roundTrip(t, istRef, "yyyy-MM-dd HH:mm:ss Z") })
	t.Run("milliseconds SSS", func(t *testing.T) { roundTrip(t, msRef, "yyyy-MM-dd HH:mm:ss.SSS") })
	t.Run("quoted literal", func(t *testing.T) { roundTrip(t, ref, "yyyy-MM-dd'T'HH:mm:ss") })
	t.Run("IsoLocalDate constant", func(t *testing.T) { roundTrip(t, ref, DateTimeIsoLocalDate) })
	t.Run("IsoLocalTime constant", func(t *testing.T) { roundTrip(t, ref, DateTimeIsoLocalTime) })
	t.Run("IsoLocalDateTime constant", func(t *testing.T) { roundTrip(t, ref, DateTimeIsoLocalDateTime) })
	t.Run("IsoOffsetDateTime constant", func(t *testing.T) { roundTrip(t, istRef, DateTimeIsoOffsetDateTime) })
	t.Run("Rfc1123 constant", func(t *testing.T) { roundTrip(t, ref, DateTimeRfc1123) })

	// Direct parse
	t.Run("direct parse UTC", func(t *testing.T) {
		parsed, err := Parse("yyyy-MM-dd HH:mm:ss", "2023-03-15 14:30:45")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := time.Time(parsed)
		if got.Year() != 2023 || got.Month() != 3 || got.Day() != 15 ||
			got.Hour() != 14 || got.Minute() != 30 || got.Second() != 45 {
			t.Errorf("unexpected time: %v", got)
		}
	})

	// Unsupported patterns
	for _, spec := range []string{"G", "Y", "Q", "q", "w", "W", "D", "F", "e", "c", "k", "K", "A", "N", "n", "O", "V", "v"} {
		_, err := Parse(spec, "x")
		if err == nil {
			t.Errorf("Parse(%q, \"x\") expected error, got nil", spec)
		}
	}
}
