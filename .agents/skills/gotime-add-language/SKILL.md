---
name: gotime-add-language
description: >
  Guides implementing a new time formatter sub-package for the gotime module
  (github.com/ducng99/gotime) that mimics another programming language's date/time
  format string syntax. Use this skill whenever the user asks to add support for
  a new language's date/time format, implement a time formatter that works like
  another language, or create a new sub-package in the gotime module. Trigger for
  phrases like "add Python strftime support", "implement Java date formatting",
  "add <language> format", or "create a formatter" — even if the user doesn't
  say "gotime" explicitly.
---

# gotime-add-language

This skill guides creating a correct, well-tested Go sub-package for the
`github.com/ducng99/gotime` module that implements time formatting using another
language's format string syntax.

## Module pattern

The canonical example is `php/php.go`. Every language sub-package follows this shape:

```
<language>/
└── <language>.go
```

Key elements:
- `package <language>` (lowercase)
- `type <Lang>Time time.Time` — a named type wrapping `time.Time`
- One exported method: `func (t *<Lang>Time) Format(format string) string`
- One unexported dispatch function that maps a format token to a string

## Implementation requirements

### 1. Look up the official spec

Always link to the authoritative documentation for the target language in a comment
above the method. Users may want to check edge cases against the real docs.

### 2. Understand the format syntax first

Before writing any code, determine:
- **What is a format specifier?** (e.g. a `%`-prefixed character, a pattern letter, a `{}`-placeholder)
- **What is a literal character?** (characters that pass through unchanged)
- **What is the escape mechanism?** (how does the format string include a literal format character)

### 3. Choose the right parser shape

Most languages fall into one of two shapes:

**Prefix-escape** (e.g. `%Y`, `%m` in Python strftime): scan character by character;
when you see the escape prefix, consume the next character as a specifier. A doubled
escape (`%%`) typically produces a literal escape character.

**Run-length grouped** (e.g. `yyyy`, `MM`, `MMMM` in Java SimpleDateFormat): group
consecutive identical letters; the count changes the output format. Literal text is
often delimited by quotes.

The PHP implementation uses backslash-escape, which is a variant of prefix-escape.
If the target language uses a different shape, don't force the PHP loop onto it — the
parsing structure should mirror what the language actually specifies.

### 4. Map specifiers to Go time expressions

Go formats time against a fixed reference: `Mon Jan 2 15:04:05 MST 2006`.
Use `t.Format("...")` for anything expressible in that layout.

For values not representable as a Go layout string (week-of-year, day-of-year,
Unix timestamp, nanoseconds, etc.), compute them using:
- `t.YearDay()` — day of year, 1-indexed
- `t.Unix()` — Unix timestamp
- `t.Nanosecond()` — nanoseconds within the second
- `t.ISOWeek()` — returns (isoYear, isoWeek) per ISO 8601
- `t.Weekday()` — day of week (Go: Sunday=0)
- `t.Zone()` — returns (name string, offsetSeconds int)
- `time.Date(...)` — construct a reference date for week calculations

### 5. Common pitfalls to watch for

- **Week-of-year is subtle.** Most languages have at least two variants (ISO/Monday-based
  vs. Sunday-based). Verify your formula against known dates — don't guess.
- **12-hour vs 24-hour hour.** Check whether the specifier means 0-23, 0-11, or 1-12.
- **AM/PM affects the hour.** `%I` or `h` in 12-hour mode needs modular arithmetic.
- **Locale-dependent specifiers** (e.g. Python's `%c`, `%x`) have no universal Go equivalent.
  Approximate with C-locale behavior and document the limitation.
- **Rune safety.** Iterate over `[]rune(format)`, not `[]byte`, to handle multi-byte
  characters in format strings without corrupting indices.
- **Escape at end of string.** If the escape character (`\`, `%`) is the last character
  with no following character, handle it gracefully (emit it literally or skip).

## Implementing Parse

Every sub-package should also export `func Parse(format, value string) (Time, error)` — the inverse of `Format`. The architecture is always a custom token walker, **not** a conversion of the format string to a Go layout string. `time.Parse` layout conversion cannot handle tokens that interact (e.g. day-of-year, ISO week date, nano-of-day) and produces confusing errors for callers.

### Parse state struct

Accumulate parsed components in a state struct. Sentinel `-1` for unset fields (so zero is distinguishable):

```go
type fooParseState struct {
    year, month, day, hour, min, sec, nano int
    loc *time.Location

    use12h bool
    ampm   int // 0=unset 1=am 2=pm

    // Extended fields for derived date specifiers:
    hasDayOfYear bool; dayOfYear int  // language-specific: 0-based (PHP z) or 1-based (Java D)
    hasIsoWeek   bool; isoWeekYear, isoWeek int
    hasWeekday   bool; weekday int    // Go convention: 0=Sun…6=Sat
    hasUnix      bool; unixSecs int64
    // ...add more as needed
}

func newParseState() *fooParseState {
    return &fooParseState{year: -1, month: -1, day: -1, hour: -1, min: -1, sec: -1}
}
```

### Token dispatch loop

Walk format and value rune-by-rune in lockstep. Each token handler receives `rest = vRunes[vi:]` and returns `(consumed int, err error)`:

```go
func fooParse(format, value string) (time.Time, error) {
    fRunes := []rune(format)
    vRunes := []rune(value)
    vi := 0
    ps := newParseState()

    for fi := 0; fi < len(fRunes); fi++ {
        ch := fRunes[fi]

        // Language-specific escape/literal handling (see below).
        // ...

        var consumed int
        var err error
        rest := vRunes[vi:]

        switch ch {
        case 'Y':
            consumed, err = parseFixedInt(rest, &ps.year, 4)
        // ... all other tokens
        default:
            // Non-specifier: match character literally.
            if len(rest) == 0 || rest[0] != ch {
                err = fmt.Errorf("parse: expected %q at position %d", string(ch), vi)
            } else {
                consumed = 1
            }
        }

        if err != nil {
            return time.Time{}, err
        }
        vi += consumed
    }

    if vi < len(vRunes) {
        return time.Time{}, fmt.Errorf("parse: unexpected trailing text: %q", string(vRunes[vi:]))
    }
    return ps.build()
}
```

**PHP-style escape** (backslash): `if ch == '\\' { fi++; match fRunes[fi] literally; continue }`.

**Java-style quoted literals** (`'text'`, `''`=single quote): when `ch == '\''`, walk forward consuming literal characters until the closing `'`, then do `fi--; continue` so the outer `fi++` lands correctly. For run-length tokens (`yyyy`, `MM`), find the run end with an inner `j` loop, then `fi = j - 1` at the end of the case.

### Integer helpers

Write two helpers (see `php/parse.go` and `java/parse.go` for exact signatures):

- **`parseFixedInt(rest []rune, dst *int, width int) (int, error)`** — exactly N digits, no range check.
- **`parseVarInt(rest []rune, dst *int, minWidth, maxWidth int) (int, error)`** — 1..maxWidth digits. PHP adds `minVal`/`maxVal` range bounds; Java does not. Match the actual output of `Format` for this language.

For large integers (e.g. nano-of-day, Unix timestamp), use `int64` and `strconv.ParseInt`.

For fractional seconds where count = significant digits, scale to nanoseconds: `nano = parsed * 10^(9-prec)`.

### build() — assembling the time.Time

Priority order matters. Apply overrides from most-specific to least-specific:

1. **Nano-of-day / ms-of-day** — decompose into h/m/s/ns, overriding any individually parsed fields.
2. **Unix timestamp** — return `time.Unix(secs, 0)` directly if present.
3. **AM/PM adjustment** — apply after setting hour:
   - `h` (1-12): `hour = hour % 12; if pm { hour += 12 }`
   - `K` (0-11): `if pm { hour += 12 }` (no modulo step)
4. **ISO week date** — if `hasIsoWeek && isoWeek > 0 && hasWeekday`, call `isoWeekDate(year, week, weekday)`.
5. **Day of year** — `time.Date(year, January, dayOfYear, ...)` lets Go normalize to the correct month/day. PHP uses 0-based `z` so add 1: `time.Date(year, January, 1+dayOfYear, ...)`.
6. **Default** — `time.Date(year, month, day, hour, min, sec, nano, loc)`.

Default unset fields before calling `time.Date`: year→0, month→January, day→1, hour/min/sec→0. Use `max(ps.year, 0)` etc.

```go
func (ps *fooParseState) build() (time.Time, error) {
    loc := ps.loc
    if loc == nil {
        loc = time.UTC
    }
    year := max(ps.year, 0)
    month := time.Month(ps.month); if ps.month < 0 { month = time.January }
    day := ps.day;                 if day < 0     { day = 1 }
    hour := max(ps.hour, 0)
    min  := max(ps.min, 0)
    sec  := max(ps.sec, 0)
    // ... apply priority overrides above ...
    return time.Date(year, month, day, hour, min, sec, ps.nano, loc), nil
}
```

### ISO week date helper

Both PHP and Java use this same algorithm:

```go
func isoWeekDate(year, week, weekday int) time.Time {
    jan4 := time.Date(year, time.January, 4, 0, 0, 0, 0, time.UTC)
    jan4ISO := int(jan4.Weekday())
    if jan4ISO == 0 { jan4ISO = 7 }
    week1Mon := jan4.AddDate(0, 0, 1-jan4ISO)
    targetISO := weekday  // Go 0=Sun, convert: if 0 { targetISO = 7 }
    if targetISO == 0 { targetISO = 7 }
    return week1Mon.AddDate(0, 0, (week-1)*7+(targetISO-1))
}
```

Note: Go weekday 0=Sun maps to ISO weekday 7. Java `e`/`c` use Sun=1…Sat=7, so convert with `goWeekday = javaDay - 1`.

### Two-digit year

```go
func twoDigitYear(y int) int {
    if y >= 69 { return 1900 + y }
    return 2000 + y
}
```

### Timezone helpers

- **Numeric offset `±HHMM`**: parse sign + 4 digits → `time.FixedZone("", sign*(hh*3600+mm*60))`.
- **Numeric offset with colon `±HH:MM`**: same but skip the `:`.
- **GMT prefix `GMT±h[:mm]`**: consume "GMT", then optional sign+hour+optional colon+minutes. The hour may be 1 or 2 digits (e.g. `GMT+5:30` vs `GMT+05:30`).
- **ISO 8601 `±HH`, `±HHmm`, `±HH:mm`, `Z`**: `parseJavaTzISO(rest, allowZ, dst)`.
- **Abbreviation** (e.g. `UTC`, `EST`): consume letters, try `time.Parse("MST", name)`, fall back to `time.FixedZone(name, 0)` if unknown.
- **Identifier** (e.g. `America/New_York`): consume letters/digits/`/`/`_`/`-`/`+`, call `time.LoadLocation(name)`.

### Consume-and-discard tokens

Some tokens (ordinal suffix `st/nd/rd/th`, leap-year flag, DST flag, day-of-week-in-month, quarter) contribute nothing to the parsed time but their characters still need to be consumed. Parse them without storing the result, or store into a dummy variable. Always return the correct `consumed` count so `vi` advances past them.

### What to return an error for

Only return an error for:
- Genuinely non-round-trippable tokens (e.g. PHP `X` — expanded year with sign has no Go equivalent).
- Malformed input (digit expected, got letter, etc.).

Do **not** return errors for tokens like quarter, era, or weekday name that are informational but irreversible — consume and discard them instead.

## What to test

Write a `_test.go` file with a table of `(format string, expected string)` pairs verified
against the reference language. Cover:

1. **Every specifier** — at least one case per format code
2. **Edge cases for tricky specifiers:**
   - Week-of-year: test a year where Jan 1 is each day of the week
   - 12h hour: midnight (0 → should be 12), noon (12 → stays 12), 1pm (13 → 1)
   - AM/PM transitions
   - Leap year: Feb 29
   - Day of year: Jan 1 (001) and Dec 31 (365 or 366)
3. **Escape handling:** literal escape character, escaped format character
4. **Unicode in format string:** a literal CJK or emoji character should pass through untouched
5. **Timezone:** test with UTC and a named non-UTC location

### Testing Parse

Use a `roundTrip` helper — format a known `time.Time`, parse it back, format again, and compare strings (not `time.Time` values, which may differ in location pointer):

```go
roundTrip := func(t *testing.T, tm Time, format string) {
    t.Helper()
    formatted := tm.Format(format)
    parsed, err := Parse(format, formatted)
    if err != nil {
        t.Fatalf("Parse(%q, %q) error: %v", format, formatted, err)
    }
    if parsed.Format(format) != formatted {
        t.Errorf("round-trip: got %q, want %q", parsed.Format(format), formatted)
    }
}
```

Run `roundTrip` for:
- Every Format specifier that Parse supports
- UTC reference time and at least one `time.FixedZone` reference time for timezone tests
- Tokens that produce derived values (day-of-year, ISO week date, ms-of-day, etc.) — these need a matching weekday/week/year triple that actually round-trips

Add direct-parse tests (known input → check `.Year()`, `.Month()`, `.Hour()` etc.) for simple formats to catch off-by-one errors independently of Format.

For tokens that are consume-and-discard during Parse (quarter, era, ordinal suffix), include them in a round-trip format string alongside date fields that do get parsed — the round-trip string comparison will confirm the consumed count is correct.

## Verification

After writing the sub-package:
```bash
go build ./...
go test ./<language>/...
```

If the test output differs from what the reference language produces, the formula or mapping is wrong — fix it before marking done.

For Parse specifically: if a round-trip fails, the most common causes are wrong `consumed` count (leaving stale characters for the next token), incorrect field priority in `build()`, or an off-by-one in a 0-based vs 1-based specifier.
