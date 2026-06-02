package ruby

import (
	"testing"
	"time"
)

func TestParseRuby(t *testing.T) {
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

	// Direct patterns
	t.Run("date", func(t *testing.T) { roundTrip(t, ref, "%Y-%m-%d") })
	t.Run("datetime", func(t *testing.T) { roundTrip(t, ref, "%Y-%m-%d %H:%M:%S") })
	t.Run("2-digit year", func(t *testing.T) { roundTrip(t, ref, "%y-%m-%d") })
	t.Run("full names", func(t *testing.T) { roundTrip(t, ref, "%A, %d %B %Y") })
	t.Run("short names", func(t *testing.T) { roundTrip(t, ref, "%a, %d %b %Y") })
	t.Run("12h AM/PM", func(t *testing.T) { roundTrip(t, ref, "%I:%M:%S %p") })
	t.Run("12h am/pm", func(t *testing.T) { roundTrip(t, ref, "%I:%M:%S %P") })
	t.Run("space-padded day", func(t *testing.T) { roundTrip(t, ref, "%e %b %Y") })
	t.Run("no-pad day/month", func(t *testing.T) { roundTrip(t, ref, "%-d %-m %Y") })
	t.Run("numeric tz", func(t *testing.T) { roundTrip(t, istRef, "%Y-%m-%d %H:%M:%S %z") })
	t.Run("colon tz", func(t *testing.T) { roundTrip(t, istRef, "%Y-%m-%d %H:%M:%S %:z") })
	t.Run("milliseconds", func(t *testing.T) { roundTrip(t, msRef, "%Y-%m-%d %H:%M:%S.%L") })
	t.Run("nanoseconds", func(t *testing.T) { roundTrip(t, msRef, "%Y-%m-%d %H:%M:%S.%N") })
	t.Run("3-digit subsecond", func(t *testing.T) { roundTrip(t, msRef, "%Y-%m-%d %H:%M:%S.%3N") })

	// Composite patterns
	t.Run("composite %F", func(t *testing.T) { roundTrip(t, ref, "%F") })
	t.Run("composite %T", func(t *testing.T) { roundTrip(t, ref, "%T") })
	t.Run("composite %r", func(t *testing.T) { roundTrip(t, ref, "%r") })
	t.Run("composite %R", func(t *testing.T) { roundTrip(t, ref, "%R") })
	t.Run("composite %D", func(t *testing.T) { roundTrip(t, ref, "%D") })
	t.Run("composite %c", func(t *testing.T) { roundTrip(t, ref, "%c") })
	t.Run("composite %v", func(t *testing.T) { roundTrip(t, ref, "%v") })

	// Non-direct patterns
	t.Run("%s unix timestamp", func(t *testing.T) { roundTrip(t, ref, "%s") })
	t.Run("%C%y century+year", func(t *testing.T) { roundTrip(t, ref, "%C%y") })
	t.Run("%Y-%j year and day of year", func(t *testing.T) { roundTrip(t, ref, "%Y-%j") })
	t.Run("%G-%V-%u ISO week date", func(t *testing.T) { roundTrip(t, ref, "%G-%V-%u") })
	t.Run("%Y-%U-%w Sunday-based week", func(t *testing.T) { roundTrip(t, ref, "%Y-%U-%w") })
	t.Run("%Y-%W-%w Monday-based week", func(t *testing.T) { roundTrip(t, ref, "%Y-%W-%w") })
	t.Run("%u ISO weekday", func(t *testing.T) { roundTrip(t, ref, "%Y-%m-%d %u") })
	t.Run("%w Ruby weekday", func(t *testing.T) { roundTrip(t, ref, "%Y-%m-%d %w") })
	t.Run("%g-%V-%u 2-digit ISO year", func(t *testing.T) { roundTrip(t, ref, "%g-%V-%u") })

	// Direct parse
	t.Run("direct parse UTC", func(t *testing.T) {
		parsed, err := Parse("%Y-%m-%d %H:%M:%S", "2023-03-15 14:30:45")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := time.Time(parsed)
		if got.Year() != 2023 || got.Month() != 3 || got.Day() != 15 ||
			got.Hour() != 14 || got.Minute() != 30 || got.Second() != 45 {
			t.Errorf("unexpected time: %v", got)
		}
	})

	t.Run("direct parse %s", func(t *testing.T) {
		parsed, err := Parse("%s", "1678890645")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := time.Time(parsed)
		if got.Year() != 2023 || got.Month() != 3 || got.Day() != 15 ||
			got.Hour() != 14 || got.Minute() != 30 || got.Second() != 45 {
			t.Errorf("unexpected time: %v", got)
		}
	})

	t.Run("direct parse %C%y", func(t *testing.T) {
		parsed, err := Parse("%C%y-%m-%d", "2023-03-15")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := time.Time(parsed)
		if got.Year() != 2023 || got.Month() != 3 || got.Day() != 15 {
			t.Errorf("unexpected time: %v", got)
		}
	})

	t.Run("direct parse %j", func(t *testing.T) {
		parsed, err := Parse("%Y-%j", "2023-074")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := time.Time(parsed)
		if got.Year() != 2023 || got.Month() != 3 || got.Day() != 15 {
			t.Errorf("unexpected time: %v", got)
		}
	})

	t.Run("direct parse %G-%V-%u", func(t *testing.T) {
		parsed, err := Parse("%G-%V-%u", "2023-11-3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := time.Time(parsed)
		if got.Year() != 2023 || got.Month() != 3 || got.Day() != 15 {
			t.Errorf("unexpected time: %v", got)
		}
	})

	// Verify invalid inputs still error
	for _, spec := range []string{"%C", "%j", "%s", "%u", "%w", "%U", "%W", "%V", "%G", "%g"} {
		_, err := Parse(spec, "x")
		if err == nil {
			t.Errorf("Parse(%q, \"x\") expected error, got nil", spec)
		}
	}
}
