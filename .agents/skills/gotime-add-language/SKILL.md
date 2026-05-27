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
- One exported method: `func (t *<Lang>Time) Format<Lang>(format string) string`
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

## Verification

After writing the sub-package:
```bash
go build ./...
go test ./<language>/...
```

If the test output differs from what the reference language produces, the formula or
mapping is wrong — fix it before marking done.
