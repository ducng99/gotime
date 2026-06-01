package java

import (
	"time"
)

type Time time.Time

// DateTimeFormatter constants - Java predefined format patterns
// See https://docs.oracle.com/en/java/javase/17/docs/api/java.base/java/time/format/DateTimeFormatter.html
// Note: Java's predefined formatters are DateTimeFormatter objects built with DateTimeFormatterBuilder,
// not simple pattern strings. These are the closest pattern equivalents.
const (
	DateTimeBasicIsoDate      = `yyyyMMdd`
	DateTimeIsoLocalDate      = `yyyy-MM-dd`
	DateTimeIsoOffsetDate     = `yyyy-MM-ddXXX`
	DateTimeIsoDate           = `yyyy-MM-ddXXX`
	DateTimeIsoLocalTime      = `HH:mm:ss`
	DateTimeIsoOffsetTime     = `HH:mm:ssXXX`
	DateTimeIsoTime           = `HH:mm:ssXXX`
	DateTimeIsoLocalDateTime  = `yyyy-MM-dd'T'HH:mm:ss`
	DateTimeIsoOffsetDateTime = `yyyy-MM-dd'T'HH:mm:ssXXX`
	DateTimeIsoZonedDateTime  = `yyyy-MM-dd'T'HH:mm:ssXXX'['VV']'`
	DateTimeIsoDateTime       = `yyyy-MM-dd'T'HH:mm:ssXXX'['VV']'`
	DateTimeIsoOrdinalDate    = `yyyy-DDD`
	DateTimeIsoWeekDate       = `YYYY-'W'ww-e`
	DateTimeIsoInstant        = `yyyy-MM-dd'T'HH:mm:ss'Z'`
	DateTimeRfc1123           = `EEE, d MMM yyyy HH:mm:ss z`
)
