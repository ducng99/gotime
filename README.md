# gotime

Format Go `time.Time` values using other languages' date format strings.

## Installation

```bash
go get github.com/ducng99/gotime
```

## Usage

### PHP Format

```go
package main

import (
    "fmt"
    "time"

    "github.com/ducng99/gotime/php"
)

func main() {
    t := php.PhpTime(time.Now())
    fmt.Println(t.Format("Y-m-d H:i:s"))  // 2026-05-27 14:30:00
}
```

### Ruby Format

```go
package main

import (
    "fmt"
    "time"

    "github.com/ducng99/gotime/ruby"
)

func main() {
    t := ruby.RubyTime(time.Now())
    fmt.Println(t.Format("%Y-%m-%d %H:%M:%S"))  // 2026-05-27 14:30:00
}
```

## Supported Format Specifiers

- **PHP**: Full support for PHP `date()` format specifiers (see [PHP docs](https://www.php.net/manual/en/datetime.format.php))
- **Ruby**: Full support for Ruby `strftime` format specifiers including flags, width, and case modifiers (see [Ruby docs](https://docs.ruby-lang.org/en/master/strftime_formatting_rdoc.html))
