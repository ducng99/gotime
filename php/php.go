package php

import "time"

type Time time.Time

// DateTimeInterface constants - PHP date format constants
const (
	DateTimeATOM            = `Y-m-d\TH:i:sP`
	DateTimeCOOKIE          = `l, d-M-Y H:i:s T`
	DateTimeISO8601         = `Y-m-d\TH:i:sO`
	DateTimeISO8601Expanded = `X-m-d\TH:i:sP`
	DateTimeRFC822          = `D, d M y H:i:s O`
	DateTimeRFC850          = `l, d-M-y H:i:s T`
	DateTimeRFC1036         = `D, d M y H:i:s O`
	DateTimeRFC1123         = `D, d M Y H:i:s O`
	DateTimeRFC7231         = `D, d M Y H:i:s \G\M\T`
	DateTimeRFC2822         = `D, d M Y H:i:s O`
	DateTimeRFC3339         = `Y-m-d\TH:i:sP`
	DateTimeRFC3339Extended = `Y-m-d\TH:i:s.vP`
	DateTimeRSS             = `D, d M Y H:i:s O`
	DateTimeW3C             = `Y-m-d\TH:i:sP`
)
